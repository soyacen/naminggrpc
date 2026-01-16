package mongox

import (
	"context"
	"fmt"

	"github.com/soyacen/goconc/lazyload"
	"go.mongodb.org/mongo-driver/v2/mongo"
	mongooptions "go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/v2/mongo/otelmongo"
)

func (options *Options) ToClientOptions() *mongooptions.ClientOptions {
	opts := mongooptions.Client()
	opts = opts.ApplyURI(options.GetUri().GetValue())
	if options.GetEnableOtel().GetValue() {
		opts.SetMonitor(otelmongo.NewMonitor())
	}
	return opts
}

func NewClients(ctx context.Context, config *Config) *lazyload.Group[*mongo.Client] {
	return &lazyload.Group[*mongo.Client]{
		New: func(key string) (*mongo.Client, error) {
			configs := config.GetConfigs()
			options, ok := configs[key]
			if !ok {
				return nil, fmt.Errorf("nacos %s not found", key)
			}
			return NewClient(ctx, options)
		},
	}
}

func NewClient(ctx context.Context, options *Options) (*mongo.Client, error) {
	return mongo.Connect(options.ToClientOptions())
}
