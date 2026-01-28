package mongox

import (
	"context"
	"fmt"

	"github.com/soyacen/gox/conc/lazyload"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mongooptions "go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/v2/mongo/otelmongo"
)

// ToClientOptions converts the protobuf Options to MongoDB ClientOptions
// It applies the URI from the options and sets up OpenTelemetry monitoring if enabled
func (options *Options) ToClientOptions() *mongooptions.ClientOptions {
	opts := mongooptions.Client()
	opts = opts.ApplyURI(options.GetUri().GetValue())
	if options.GetEnableOtel().GetValue() {
		// Set up OpenTelemetry monitor to track MongoDB operations
		opts.SetMonitor(otelmongo.NewMonitor())
	}
	return opts
}

// NewClients creates a lazy loading group for MongoDB clients
// It allows creating multiple clients based on the provided configuration map
func NewClients(ctx context.Context, config *Config) *lazyload.Group[*mongo.Client] {
	return &lazyload.Group[*mongo.Client]{
		New: func(key string) (*mongo.Client, error) {
			configs := config.GetConfigs()
			options, ok := configs[key]
			if !ok {
				// Return error if the specified configuration key doesn't exist
				return nil, fmt.Errorf("mongo config %s not found", key)
			}
			return NewClient(ctx, options)
		},
		Finalize: func(ctx context.Context, obj *mongo.Client) error {
			return obj.Disconnect(ctx)
		},
	}
}

// NewClient creates a new MongoDB client with the given options
// It connects to the MongoDB instance using the provided configuration
func NewClient(ctx context.Context, options *Options) (*mongo.Client, error) {
	return mongo.Connect(options.ToClientOptions())
}
