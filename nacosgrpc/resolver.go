// Package nacosgrpc provides gRPC naming resolver implementation based on Nacos
package nacosgrpc

import (
	"cmp"
	"context"
	"net"
	"slices"
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/attributes"
	grpcresolver "google.golang.org/grpc/resolver"
)

// scheme defines the protocol scheme name for this resolver
const scheme = "nacos"

// init automatically registers the resolver builder when the package is imported
func init() {
	grpcresolver.Register(&Builder{})
	grpc.Dial("", grpc.WithResolvers())
}

// Builder is the Builder structure for Nacos resolver
type Builder struct{}

// Build creates a new resolver instance based on target address
// Parameters:
//   - tgt: target address information
//   - cc: client connection interface
//   - opts: build options
//
// Returns:
//   - resolver.Resolver: resolver instance
//   - error: error information
func (b *Builder) Build(tgt grpcresolver.Target, cc grpcresolver.ClientConn, opts grpcresolver.BuildOptions) (grpcresolver.Resolver, error) {
	// Parse DSN configuration
	parsed, err := DefaultDsnParser(context.Background(), "resolver", tgt.URL.String())
	if err != nil {
		return nil, errors.Wrapf(err, "naming: failed to parse dsn")
	}

	// Create Nacos naming client
	client, err := clients.NewNamingClient(parsed.ClientParam)
	if err != nil {
		return nil, errors.Wrapf(err, "naming: failed to create naming client")
	}

	// Create resolver instance
	r := &resolver{
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
func (b *Builder) Scheme() string {
	return scheme
}

// resolver is the concrete implementation of Nacos resolver
type resolver struct {
	target       grpcresolver.Target         // Target address information
	cc           grpcresolver.ClientConn     // Client connection interface
	namingClient naming_client.INamingClient // Nacos naming client
	param        vo.SubscribeParam           // Subscription parameters
}

// start starts the resolver, performs initial service discovery and subscription
func (r *resolver) start() error {
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
func (r *resolver) updateState(instances []model.Instance) {
	addrs := make([]grpcresolver.Address, 0, len(instances))

	// Iterate through all instances, filter healthy and enabled instances
	for _, inst := range instances {
		if !inst.Healthy || !inst.Enable {
			continue
		}

		// Construct gRPC address
		addr := grpcresolver.Address{
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
	slices.SortFunc(addrs, func(a, b grpcresolver.Address) int {
		return cmp.Compare(a.Addr, b.Addr)
	})

	// Update client connection state
	r.cc.UpdateState(grpcresolver.State{Addresses: addrs})
}

// ResolveNow resolves the target address immediately (currently implemented as no-op)
func (r *resolver) ResolveNow(grpcresolver.ResolveNowOptions) {}

// Close closes the resolver, unsubscribes and closes the client
func (r *resolver) Close() {
	_ = r.namingClient.Unsubscribe(&r.param)
	r.namingClient.CloseClient()
}
