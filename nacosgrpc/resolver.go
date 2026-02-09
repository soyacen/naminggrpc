// Package nacosgrpc provides gRPC naming resolver implementation based on Nacos
package nacosgrpc

import (
    "cmp"
    "net"
    "net/url"
    "slices"
    "strconv"
    "strings"

    "github.com/cockroachdb/errors"
    "github.com/nacos-group/nacos-sdk-go/v2/clients"
    "github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
    "github.com/nacos-group/nacos-sdk-go/v2/common/constant"
    "github.com/nacos-group/nacos-sdk-go/v2/model"
    "github.com/nacos-group/nacos-sdk-go/v2/vo"
    "google.golang.org/grpc/attributes"
    "google.golang.org/grpc/resolver"
)

// scheme defines the protocol scheme name for this resolver
const scheme = "nacos"

// ResolverDSNParser defines the function for parsing resolver DSN
var ResolverDSNParser = parseResolverDSN

// init automatically registers the resolver builder when the package is imported
func init() {
    resolver.Register(&nacosResolverBuilder{})
}

// nacosResolverBuilder is the builder structure for Nacos resolver
type nacosResolverBuilder struct{}

// Build creates a new resolver instance based on target address
// Parameters:
//   - tgt: target address information
//   - cc: client connection interface
//   - opts: build options
//
// Returns:
//   - resolver.Resolver: resolver instance
//   - error: error information
func (b *nacosResolverBuilder) Build(tgt resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
    // Parse DSN configuration
    parsed, err := ResolverDSNParser(tgt.URL)
    if err != nil {
        return nil, errors.Wrapf(err, "naming: failed to parse dsn")
    }

    // Create Nacos naming client
    client, err := clients.NewNamingClient(parsed.ClientParam)
    if err != nil {
        return nil, errors.Wrapf(err, "naming: failed to create naming client")
    }

    // Create resolver instance
    r := &nacosResolver{
        target:       tgt,
        cc:           cc,
        namingClient: client,
        param:        parsed.SubscribeParam,
    }

    // Start the resolver
    if err := r.start(); err != nil {
        return nil, err
    }

    return r, nil
}

// Scheme returns the protocol scheme supported by this resolver
func (b *nacosResolverBuilder) Scheme() string {
    return scheme
}

// nacosResolver is the concrete implementation of Nacos resolver
type nacosResolver struct {
    target       resolver.Target             // Target address information
    cc           resolver.ClientConn         // Client connection interface
    namingClient naming_client.INamingClient // Nacos naming client
    param        vo.SubscribeParam           // Subscription parameters
}

// start starts the resolver, performs initial service discovery and subscription
func (r *nacosResolver) start() error {
    // Initial retrieval of service instance list
    instances, err := r.namingClient.SelectInstances(vo.SelectInstancesParam{
        ServiceName: r.param.ServiceName,
        GroupName:   r.param.GroupName,
        Clusters:    r.param.Clusters,
        HealthyOnly: true,
    })
    if err != nil {
        return errors.Wrapf(err, "naming: failed to select instances")
    }

    // Update connection state
    r.updateState(instances)

    // Set subscription callback function
    r.param.SubscribeCallback = func(services []model.Instance, err error) {
        if err != nil {
            return
        }
        r.updateState(services)
    }

    // Subscribe to service changes
    if err := r.namingClient.Subscribe(&r.param); err != nil {
        return errors.Wrapf(err, "naming: failed to subscribe")
    }
    return nil
}

// updateState updates resolver state, converting service instances to gRPC address format
// Parameters:
//   - instances: Nacos service instance list
func (r *nacosResolver) updateState(instances []model.Instance) {
    addrs := make([]resolver.Address, 0, len(instances))

    // Iterate through all instances, filter healthy and enabled instances
    for _, inst := range instances {
        if !inst.Healthy || !inst.Enable {
            continue
        }

        // Construct gRPC address
        addr := resolver.Address{
            Addr: net.JoinHostPort(inst.Ip, strconv.FormatUint(inst.Port, 10)),
        }

        // Add instance attribute information
        addr.Attributes = attributes.New("ServiceName", inst.ServiceName).
            WithValue("weight", inst.Weight).
            WithValue("cluster", inst.ClusterName).
            WithValue("ephemeral", inst.Ephemeral)

        // Add custom metadata
        for k, v := range inst.Metadata {
            addr.Attributes = addr.Attributes.WithValue(k, v)
        }

        addrs = append(addrs, addr)
    }

    // Sort addresses to ensure consistency
    slices.SortFunc(addrs, func(a, b resolver.Address) int {
        return cmp.Compare(a.Addr, b.Addr)
    })

    // Update client connection state
    r.cc.UpdateState(resolver.State{Addresses: addrs})
}

// ResolveNow resolves the target address immediately (currently implemented as no-op)
func (r *nacosResolver) ResolveNow(resolver.ResolveNowOptions) {}

// Close closes the resolver, unsubscribes and closes the client
func (r *nacosResolver) Close() {
    _ = r.namingClient.Unsubscribe(&r.param)
    r.namingClient.CloseClient()
}

// ResolverDSN contains configuration information required by the resolver
type ResolverDSN struct {
    ClientParam    vo.NacosClientParam // Nacos client parameters
    SubscribeParam vo.SubscribeParam   // Subscription parameters
}

// parseResolverDSN parses URL-formatted DSN configuration
// Supported URL format: nacos://[username[:password]@]host[:port]/service_name?param=value
// Parameters:
//   - u: URL object
//
// Returns:
//   - *ResolverDSN: parsed configuration object
//   - error: error information
func parseResolverDSN(u url.URL) (*ResolverDSN, error) {
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

    clusters := q.Get("clusters")

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

    // Parse service name
    serviceName := strings.Trim(u.Path, "/")
    if serviceName == "" {
        return nil, errors.New("naming: service name is required in path")
    }

    // Return parsed result
    return &ResolverDSN{
        ClientParam: vo.NacosClientParam{
            ClientConfig:  clientConfig,
            ServerConfigs: serverConfigs,
        },
        SubscribeParam: vo.SubscribeParam{
            ServiceName: serviceName,
            GroupName:   group,
            Clusters:    strings.Split(clusters, ","),
        },
    }, nil
}