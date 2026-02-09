// Package nacosgrpc provides gRPC service registration implementation based on Nacos
package nacosgrpc

import (
	"context"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/soyacen/naminggrpc"
)

// RegistrarDSNParser is the function variable for parsing registrar DSN
var RegistrarDSNParser = parseRegistrarDSN

// init registers the nacos factory with naminggrpc
func init() {
	naminggrpc.Register("nacos", &Factory{})
}

// Registrar represents the Nacos service registrar structure
type Registrar struct {
	namingClient    naming_client.INamingClient // Nacos naming client
	registerParam   vo.RegisterInstanceParam    // Registration instance parameters
	deregisterParam vo.DeregisterInstanceParam  // Deregistration instance parameters
}

// Register registers a service instance in Nacos
// Parameters:
//   - ctx: context object
//
// Returns:
//   - error: error information during registration
func (r *Registrar) Register(ctx context.Context) error {
	// Call Nacos client to register instance
	ok, err := r.namingClient.RegisterInstance(r.registerParam)
	if err != nil {
		return errors.Wrapf(err, "nacosgrpc: failed to register %s", r.registerParam.ServiceName)
	}

	// Check if registration was successful
	if !ok {
		return errors.Errorf("nacosgrpc: failed to register %s", r.registerParam.ServiceName)
	}

	return nil
}

// Deregister deregisters a service instance from Nacos
// Parameters:
//   - ctx: context object
//
// Returns:
//   - error: error information during deregistration
func (r *Registrar) Deregister(ctx context.Context) error {
	// Call Nacos client to deregister instance
	ok, err := r.namingClient.DeregisterInstance(r.deregisterParam)
	if err != nil {
		return errors.Wrapf(err, "nacosgrpc: failed to deregister %s", r.deregisterParam.ServiceName)
	}

	// Check if deregistration was successful
	if !ok {
		return errors.Errorf("nacosgrpc: failed to deregister %s", r.deregisterParam.ServiceName)
	}

	return nil
}

// NewRegistrar creates a new Nacos service registrar instance
// Parameters:
//   - dsn: data source name, format: nacos://[username[:password]@]host[:port]/service_name?param=value
//
// Returns:
//   - *Registrar: registrar instance
//   - error: error information during creation
func NewRegistrar(dsn string) (*Registrar, error) {
	// Parse DSN URL
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, errors.Wrapf(err, "nacosgrpc: failed to parse dsn %s", dsn)
	}

	// Parse registrar DSN configuration
	parsed, err := RegistrarDSNParser(*u)
	if err != nil {
		return nil, err
	}

	// Create Nacos naming client
	client, err := clients.NewNamingClient(parsed.clientParam)
	if err != nil {
		return nil, errors.Wrapf(err, "nacosgrpc: failed to create naming client")
	}

	// Return registrar instance
	return &Registrar{
		namingClient:    client,
		registerParam:   parsed.registerParam,
		deregisterParam: parsed.deregisterParam,
	}, nil
}

// RegistrarDSN contains configuration information required by the registrar
type RegistrarDSN struct {
	clientParam     vo.NacosClientParam        // Nacos client parameters
	registerParam   vo.RegisterInstanceParam   // Registration instance parameters
	deregisterParam vo.DeregisterInstanceParam // Deregistration instance parameters
}

// parseRegistrarDSN parses the DSN configuration for the registrar
// Supported URL format: nacos://[username[:password]@]host[:port]/service_name?param=value
// Parameters:
//   - u: URL object
//
// Returns:
//   - *RegistrarDSN: parsed configuration object
//   - error: error information
func parseRegistrarDSN(u url.URL) (*RegistrarDSN, error) {
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

	// Parse service name
	serviceName := strings.Trim(u.Path, "/")
	if serviceName == "" {
		return nil, errors.New("nacosgrpc: service name is required in path")
	}

	// Parse instance IP and port
	svcIP := q.Get("ip")
	if svcIP == "" {
		return nil, errors.New("nacosgrpc: service IP is required in query parameters")
	}
	svcPortStr := q.Get("port")
	svcPort, err := strconv.ParseUint(svcPortStr, 10, 64)
	if err != nil {
		return nil, errors.New("nacosgrpc: invalid service port in query parameters")
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

	// Return parsed result
	return &RegistrarDSN{
		clientParam: vo.NacosClientParam{
			ClientConfig:  clientConfig,
			ServerConfigs: serverConfigs,
		},
		registerParam: vo.RegisterInstanceParam{
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
		},
		deregisterParam: vo.DeregisterInstanceParam{
			Ip:          svcIP,
			Port:        svcPort,
			Cluster:     cluster,
			ServiceName: serviceName,
			GroupName:   group,
			Ephemeral:   ephemeral,
		},
	}, nil
}

// Factory implements the registrar factory interface
type Factory struct{}

// New creates a new registrar instance
// Parameters:
//   - ctx: context object
//   - dsn: data source name
//
// Returns:
//   - naminggrpc.Registrar: registrar interface
//   - error: error information
func (f *Factory) New(ctx context.Context, dsn string) (naminggrpc.Registrar, error) {
	return NewRegistrar(dsn)
}
