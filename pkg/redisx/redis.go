package redisx

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/soyacen/gox/conc/lazyload"
)

// ToUniversalOptions converts PB-defined Options to Redis UniversalOptions
func (options *Options) ToUniversalOptions() *redis.UniversalOptions {
	opts := &redis.UniversalOptions{
		Addrs:                        options.Addrs,
		ClientName:                   options.GetClientName().GetValue(),
		DB:                           int(options.GetDb().GetValue()),
		Dialer:                       nil, // Dialer is a function type that cannot be converted from proto
		OnConnect:                    nil, // OnConnect is a function type that cannot be converted from proto
		Protocol:                     int(options.GetProtocol().GetValue()),
		Username:                     options.GetUsername().GetValue(),
		Password:                     options.GetPassword().GetValue(),
		CredentialsProvider:          nil, // CredentialsProvider is a function type that cannot be converted from proto
		CredentialsProviderContext:   nil, // CredentialsProviderContext is a function type that cannot be converted from proto
		StreamingCredentialsProvider: nil, // StreamingCredentialsProvider is an interface type that cannot be converted from proto
		SentinelUsername:             options.GetSentinelUsername().GetValue(),
		SentinelPassword:             options.GetSentinelPassword().GetValue(),
		MaxRetries:                   int(options.GetMaxRetries().GetValue()),
		MinRetryBackoff:              options.GetMinRetryBackoff().AsDuration(),
		MaxRetryBackoff:              options.GetMaxRetryBackoff().AsDuration(),
		DialTimeout:                  options.GetDialTimeout().AsDuration(),
		ReadTimeout:                  options.GetReadTimeout().AsDuration(),
		WriteTimeout:                 options.GetWriteTimeout().AsDuration(),
		ContextTimeoutEnabled:        options.GetContextTimeoutEnabled().GetValue(),
		ReadBufferSize:               int(options.GetReadBufferSize().GetValue()),
		WriteBufferSize:              int(options.GetWriteBufferSize().GetValue()),
		PoolFIFO:                     options.GetPoolFifo().GetValue(),
		PoolSize:                     int(options.GetPoolSize().GetValue()),
		PoolTimeout:                  options.GetPoolTimeout().AsDuration(),
		MinIdleConns:                 int(options.GetMinIdleConns().GetValue()),
		MaxIdleConns:                 int(options.GetMaxIdleConns().GetValue()),
		MaxActiveConns:               int(options.GetMaxActiveConns().GetValue()),
		ConnMaxIdleTime:              options.GetConnMaxIdleTime().AsDuration(),
		ConnMaxLifetime:              options.GetConnMaxLifetime().AsDuration(),
		TLSConfig:                    options.GetTlsConfig().AsConfig(),
		MaxRedirects:                 int(options.GetClusterOptions().GetMaxRedirects().GetValue()),
		ReadOnly:                     options.GetClusterOptions().GetReadOnly().GetValue(),
		RouteByLatency:               options.GetClusterOptions().GetRouteByLatency().GetValue(),
		RouteRandomly:                options.GetClusterOptions().GetRouteRandomly().GetValue(),
		MasterName:                   options.GetFailoverOptions().GetMasterName().GetValue(),
		DisableIndentity:             options.GetDisableIdentity().GetValue(), // Note: This is the old field name for backward compatibility
		DisableIdentity:              options.GetDisableIdentity().GetValue(),
		IdentitySuffix:               options.GetIdentitySuffix().GetValue(),
		FailingTimeoutSeconds:        int(options.GetFailingTimeoutSeconds().GetValue()),
		UnstableResp3:                options.GetUnstableResp3().GetValue(),
		IsClusterMode:                options.GetIsClusterMode().GetValue(),
		MaintNotificationsConfig:     nil, // MaintNotificationsConfig is a complex structure, temporarily set to nil
	}
	return opts
}

// New creates a redis client collection using lazy loading
func NewClients(ctx context.Context, config *Config) *lazyload.Group[redis.UniversalClient] {
	return &lazyload.Group[redis.UniversalClient]{
		New: func(key string) (redis.UniversalClient, error) {
			options, ok := config.GetConfigs()[key]
			if !ok {
				return nil, errors.Errorf("redisx: config %s not found", key)
			}
			return NewClient(ctx, options)
		},
	}
}

// NewClient creates a new Redis client with the given options
// It performs a ping test to ensure the connection works
// If tracing or metrics are enabled, it instruments the client accordingly
func NewClient(ctx context.Context, options *Options) (redis.UniversalClient, error) {
	// Create a new universal Redis client using the provided options
	client := redis.NewUniversalClient(options.ToUniversalOptions())

	// Test the connection by pinging the server
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, errors.Wrap(err, "redisx: failed to ping redis")
	}

	// Enable tracing if configured
	if options.GetEnableTracing().GetValue() {
		if err := redisotel.InstrumentTracing(client); err != nil {
			return nil, errors.Wrap(err, "redisx: failed to instrument tracing")
		}
	}

	// Enable metrics if configured
	if options.GetEnableMetrics().GetValue() {
		if err := redisotel.InstrumentMetrics(client); err != nil {
			return nil, errors.Wrap(err, "redisx: failed to instrument metrics")
		}
	}

	// Return the configured client
	return client, nil
}
