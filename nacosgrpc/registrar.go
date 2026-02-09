// Package nacosgrpc provides gRPC service registration implementation based on Nacos
package nacosgrpc

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/soyacen/naminggrpc"
)

var _ naminggrpc.Registrar = (*registrar)(nil)

// init registers the nacos factory with naminggrpc
func init() {
	naminggrpc.Register("nacos", &Factory{})
}

// registrar represents the Nacos service registrar structure
type registrar struct {
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
func (r *registrar) Register(ctx context.Context) error {
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
func (r *registrar) Deregister(ctx context.Context) error {
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
func NewRegistrar(dsn string) (naminggrpc.Registrar, error) {
	// Parse registrar DSN configuration
	parsed, err := DefaultDsnParser(context.Background(), "registrar", dsn)
	if err != nil {
		return nil, err
	}

	// Create Nacos naming client
	client, err := clients.NewNamingClient(parsed.ClientParam)
	if err != nil {
		return nil, errors.Wrapf(err, "nacosgrpc: failed to create naming client")
	}

	// Return registrar instance
	return &registrar{
		namingClient:    client,
		registerParam:   parsed.RegisterParam,
		deregisterParam: parsed.DeregisterParam,
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
