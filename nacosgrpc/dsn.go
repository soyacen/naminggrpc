package nacosgrpc

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

type DSN struct {
	ClientParam     vo.NacosClientParam        // Nacos client parameters
	RegisterParam   vo.RegisterInstanceParam   // Registration instance parameters
	DeregisterParam vo.DeregisterInstanceParam // Deregistration instance parameters
	SubscribeParam  vo.SubscribeParam          // Subscription parameters
}

type DsnParser func(ctx context.Context, kind string, dsn string) (*DSN, error)

var DefaultDsnParser DsnParser = ParseDsn

func ParseDsn(ctx context.Context, kind string, dsn string) (*DSN, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("nacosgrpc: failed to parse dsn %s", dsn)
	}

	// Parse host and port
	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		host = u.Host
		portStr = "8848" // Default Nacos port
	}
	port, _ := strconv.ParseUint(portStr, 10, 64)

	// Parse query parameters
	q := u.Query()
	namespace := q.Get("namespace")
	if namespace == "" {
		namespace = "public" // Default namespace
	}

	group := q.Get("group")
	if group == "" {
		group = "DEFAULT_GROUP" // Default group
	}

	// Parse timeout
	timeoutMs := uint64(10000) // Default timeout 10 seconds
	if t := q.Get("timeout"); t != "" {
		if tv, err := strconv.ParseUint(t, 10, 64); err == nil {
			timeoutMs = tv
		}
	}

	// Construct client configuration
	clientConfig := constant.NewClientConfig()
	clientConfig.Username = u.User.Username()
	clientConfig.Password, _ = u.User.Password()
	clientConfig.NamespaceId = namespace
	clientConfig.TimeoutMs = timeoutMs
	clientConfig.NotLoadCacheAtStart = true

	// Construct server configuration
	serverConfigs := []constant.ServerConfig{
		*constant.NewServerConfig(host, port),
	}

	clientParam := vo.NacosClientParam{
		ClientConfig:  clientConfig,
		ServerConfigs: serverConfigs,
	}

	// Parse service name
	serviceName := strings.Trim(u.Path, "/")
	if serviceName == "" {
		return nil, fmt.Errorf("nacosgrpc: service name is required in path")
	}

	d := &DSN{
		ClientParam: clientParam,
	}

	if kind == "registrar" {
		// Parse instance IP and port
		svcIP := q.Get("ip")
		if svcIP == "" {
			return nil, fmt.Errorf("nacosgrpc: service IP is required in query parameters")
		}
		svcPortStr := q.Get("port")
		svcPort, err := strconv.ParseUint(svcPortStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("nacosgrpc: invalid service port in query parameters")
		}

		// Parse weight configuration
		weight := 10.0 // Default weight
		if w := q.Get("weight"); w != "" {
			if wv, err := strconv.ParseFloat(w, 64); err == nil {
				weight = wv
			}
		}

		// Parse ephemeral instance configuration
		ephemeral := true // Default to ephemeral instance
		if e := q.Get("ephemeral"); e != "" {
			if ev, err := strconv.ParseBool(e); err == nil {
				ephemeral = ev
			}
		}

		// Parse cluster name and metadata
		cluster := q.Get("cluster")
		metadata := make(map[string]string)

		// Process metadata parameters starting with meta.
		for k, v := range q {
			if strings.HasPrefix(k, "meta.") {
				metadata[strings.TrimPrefix(k, "meta.")] = v[0]
			}
		}

		d.RegisterParam = vo.RegisterInstanceParam{
			Ip:          svcIP,
			Port:        svcPort,
			Weight:      weight,
			Enable:      true, // Default enabled
			Healthy:     true, // Default healthy
			Metadata:    metadata,
			ClusterName: cluster,
			ServiceName: serviceName,
			GroupName:   group,
			Ephemeral:   ephemeral,
		}
		d.DeregisterParam = vo.DeregisterInstanceParam{
			Ip:          svcIP,
			Port:        svcPort,
			Cluster:     cluster,
			ServiceName: serviceName,
			GroupName:   group,
			Ephemeral:   ephemeral,
		}
	}

	if kind == "resolver" {
		clusters := q.Get("clusters")
		d.SubscribeParam = vo.SubscribeParam{
			ServiceName: serviceName,
			GroupName:   group,
			Clusters:    strings.Split(clusters, ","),
		}
	}

	return d, nil
}
