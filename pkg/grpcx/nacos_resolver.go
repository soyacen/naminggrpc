package grpcx

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/go-playground/form"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/pkg/errors"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/resolver"
)

func init() {
	resolver.Register(&nacosResolverBuilder{})
}

const nacosResolverSchemeName = "nacos"

type nacosResolverBuilder struct{}

type nacosResolver struct {
	cancelFunc context.CancelFunc
	client     naming_client.INamingClient
}

type nacosResolverTarget struct {
	Addr         string
	Service      string
	GroupName    string
	Namespace    string
	ClusterName  string
	Timeout      time.Duration
	PollInterval time.Duration
}

func (b *nacosResolverBuilder) Build(tgt resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	dsn := strings.Join([]string{nacosResolverSchemeName + ":/", tgt.URL.Host, tgt.URL.Path + "?" + tgt.URL.RawQuery}, "/")
	target, err := parseNacosResolverURL(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "Wrong nacos URL")
	}

	cli, err := createNacosClient(target)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't connect to the Nacos API")
	}

	ctx, cancel := context.WithCancel(context.Background())
	pipe := make(chan []model.Instance, 1)
	go watchNacosService(ctx, cli, target, pipe)
	go populateEndpoints(ctx, cc, pipe)

	return &nacosResolver{cancelFunc: cancel, client: cli}, nil
}

func (b *nacosResolverBuilder) Scheme() string {
	return nacosResolverSchemeName
}

func (r *nacosResolver) ResolveNow(resolver.ResolveNowOptions) {}

func (r *nacosResolver) Close() {
	r.cancelFunc()
	if r.client != nil {
		r.client.CloseClient()
	}
}

func parseNacosResolverURL(u string) (*nacosResolverTarget, error) {
	rawURL, err := url.Parse(u)
	if err != nil {
		return nil, errors.Wrap(err, "Malformed URL")
	}

	if rawURL.Scheme != nacosResolverSchemeName || len(rawURL.Host) == 0 || len(strings.TrimLeft(rawURL.Path, "/")) == 0 {
		return nil, errors.Errorf("Malformed URL('%s'). Must be in the next format: 'nacos://host:port/service?param=value'", u)
	}

	t := &nacosResolverTarget{
		Addr:         rawURL.Host,
		Service:      strings.TrimLeft(rawURL.Path, "/"),
		GroupName:    constant.DEFAULT_GROUP,
		Timeout:      30 * time.Second,
		PollInterval: 30 * time.Second,
	}

	decoder := form.NewDecoder()
	decoder.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
		return time.ParseDuration(vals[0])
	}, time.Duration(0))

	err = decoder.Decode(t, rawURL.Query())
	if err != nil {
		return nil, errors.Wrap(err, "Malformed URL parameters")
	}

	return t, nil
}

func createNacosClient(t *nacosResolverTarget) (naming_client.INamingClient, error) {
	hostParts := strings.Split(t.Addr, ":")
	if len(hostParts) != 2 {
		return nil, errors.Errorf("Invalid host format, expected host:port, got %s", t.Addr)
	}

	host := hostParts[0]
	var port uint64
	_, err := fmt.Sscanf(hostParts[1], "%d", &port)
	if err != nil {
		return nil, errors.Wrap(err, "Invalid port format")
	}

	cc := constant.NewClientConfig(
		constant.WithNamespaceId(t.Namespace),
		constant.WithTimeoutMs(uint64(t.Timeout.Milliseconds())),
		constant.WithBeatInterval(5000),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogDir("/tmp/nacos/log"),
		constant.WithLogLevel("info"),
	)

	sc := []constant.ServerConfig{
		*constant.NewServerConfig(
			host,
			port,
			constant.WithScheme("http"),
			constant.WithContextPath("/nacos"),
		),
	}

	cli, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create Nacos client")
	}

	return cli, nil
}

func watchNacosService(ctx context.Context, client naming_client.INamingClient, t *nacosResolverTarget, output chan<- []model.Instance) {
	res := make(chan []model.Instance, 1)
	// Subscribe to service updates
	subscribeParam := &vo.SubscribeParam{
		ServiceName: t.Service,
		GroupName:   t.GroupName,
		Clusters:    []string{t.ClusterName},
		SubscribeCallback: func(services []model.Instance, err error) {
			if err != nil {
				grpclog.Errorf("[Nacos resolver] Error in subscription callback for service=%s; error=%v", t.Service, err)
				return
			}
			grpclog.Infof("[Nacos resolver] %d endpoints received for service=%s", len(services), t.Service)
			select {
			case res <- services:
			case <-ctx.Done():
				return
			}
		},
	}

	// Start subscription in a goroutine
	go func() {
		if err := client.Subscribe(subscribeParam); err != nil {
			grpclog.Errorf("[Nacos resolver] Couldn't subscribe to service=%s; error=%v", t.Service, err)
		}
	}()

	// Handle unsubscribe when context is cancelled
	go func() {
		<-ctx.Done()
		if err := client.Unsubscribe(subscribeParam); err != nil {
			grpclog.Errorf("[Nacos resolver] Couldn't unsubscribe to service=%s; error=%v", t.Service, err)
		}
	}()

	// Polling loop for fetching instances periodically
	go func() {
		// Create polling ticker
		ticker := time.NewTicker(t.PollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				instances, err := client.SelectInstances(vo.SelectInstancesParam{
					ServiceName: t.Service,
					GroupName:   t.GroupName,
					Clusters:    []string{t.ClusterName},
					HealthyOnly: true,
				})
				if err != nil {
					grpclog.Errorf("[Nacos resolver] Couldn't fetch endpoints. service=%s; error=%v", t.Service, err)
					continue
				}
				grpclog.Infof("[Nacos resolver] %d endpoints fetched for service=%s", len(instances), t.Service)

				select {
				case output <- instances:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func populateEndpoints(ctx context.Context, clientConn resolver.ClientConn, input <-chan []model.Instance) {
	for {
		select {
		case instances, ok := <-input:
			if !ok {
				// Channel was closed, exit the loop
				grpclog.Info("[Nacos resolver] Input channel closed, ending population")
				return
			}

			instanceSet := make(map[string]model.Instance, len(instances))
			for _, instance := range instances {
				// Filter out unhealthy or disabled instances
				if !instance.Enable || !instance.Healthy || instance.Weight <= 0 {
					continue
				}
				addr := fmt.Sprintf("%s:%d", instance.Ip, instance.Port)
				instanceSet[addr] = instance
			}

			addresses := make([]resolver.Address, 0, len(instanceSet))
			for addr, instance := range instanceSet {
				address := resolver.Address{
					Addr: addr,
				}
				address.Attributes = attributes.New("instance_id", instance.InstanceId)
				address.Attributes = address.Attributes.WithValue("cluster_name", instance.ClusterName)
				address.Attributes = address.Attributes.WithValue("service_name", instance.ServiceName)
				for key, value := range instance.Metadata {
					address.Attributes = address.Attributes.WithValue(key, value)
				}
				addresses = append(addresses, address)
			}

			slices.SortFunc(addresses, func(a, b resolver.Address) int { return strings.Compare(a.Addr, b.Addr) })

			if err := clientConn.UpdateState(resolver.State{Addresses: addresses}); err != nil {
				grpclog.Errorf("[Nacos resolver] Couldn't update client connection. error=%v", err)
				continue
			}

			grpclog.Infof("[Nacos resolver] Updated state with %d healthy endpoints", len(addresses))
		case <-ctx.Done():
			grpclog.Info("[Nacos resolver] Watch has been finished")
			return
		}
	}
}
