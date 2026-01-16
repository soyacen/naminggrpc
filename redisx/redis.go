package redisx

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/soyacen/goconc/lazyload"
)

// ConvertToOptions 将PB定义的Options转换为Redis UniversalOptions
func (options *Options) ToUniversalOptions() *redis.UniversalOptions {
	var tlsConfig *tls.Config
	if info := options.GetTlsConfig(); info != nil {
		cert, err := tls.LoadX509KeyPair(info.GetCertFile(), info.GetKeyFile())
		if err != nil {
			panic(err)
		}

		// 创建基础 TLS 配置
		tlsConfig = &tls.Config{Certificates: []tls.Certificate{cert}}

		// 处理 CA 证书
		if info.GetCaFile() != "" {
			caCert, err := os.ReadFile(info.GetCaFile())
			if err != nil {
				panic(err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caCertPool
		}

		// 设置服务器名称
		if info.GetServerName() != "" {
			tlsConfig.ServerName = info.GetServerName()
		}
	}
	opts := &redis.UniversalOptions{
		Addrs:                        options.Addrs,
		ClientName:                   options.GetClientName().GetValue(),
		DB:                           int(options.GetDb().GetValue()),
		Dialer:                       nil, // Dialer 是函数类型，无法从proto转换
		OnConnect:                    nil, // OnConnect 是函数类型，无法从proto转换
		Protocol:                     int(options.GetProtocol().GetValue()),
		Username:                     options.GetUsername().GetValue(),
		Password:                     options.GetPassword().GetValue(),
		CredentialsProvider:          nil, // CredentialsProvider 是函数类型，无法从proto转换
		CredentialsProviderContext:   nil, // CredentialsProviderContext 是函数类型，无法从proto转换
		StreamingCredentialsProvider: nil, // StreamingCredentialsProvider 是接口类型，无法从proto转换
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
		TLSConfig:                    tlsConfig,
		MaxRedirects:                 int(options.GetClusterOptions().GetMaxRedirects().GetValue()),
		ReadOnly:                     options.GetClusterOptions().GetReadOnly().GetValue(),
		RouteByLatency:               options.GetClusterOptions().GetRouteByLatency().GetValue(),
		RouteRandomly:                options.GetClusterOptions().GetRouteRandomly().GetValue(),
		MasterName:                   options.GetFailoverOptions().GetMasterName().GetValue(),
		DisableIndentity:             options.GetDisableIdentity().GetValue(), // 注意：这是旧字段名，保留向后兼容性
		DisableIdentity:              options.GetDisableIdentity().GetValue(),
		IdentitySuffix:               options.GetIdentitySuffix().GetValue(),
		FailingTimeoutSeconds:        int(options.GetFailingTimeoutSeconds().GetValue()),
		UnstableResp3:                options.GetUnstableResp3().GetValue(),
		IsClusterMode:                options.GetIsClusterMode().GetValue(),
		MaintNotificationsConfig:     nil, // MaintNotificationsConfig 是复杂结构，暂时设为nil
	}
	return opts
}

// New 创建redis客户端集合
func NewClients(ctx context.Context, config *Config) *lazyload.Group[redis.UniversalClient] {
	return &lazyload.Group[redis.UniversalClient]{
		New: func(key string) (redis.UniversalClient, error) {
			options, ok := config.GetConfigs()[key]
			if !ok {
				return nil, fmt.Errorf("redis %s not found", key)
			}
			return NewClient(ctx, options)
		},
	}
}

func NewClient(ctx context.Context, options *Options) (redis.UniversalClient, error) {
	client := redis.NewUniversalClient(options.ToUniversalOptions())
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	if options.GetEnableTracing().GetValue() {
		if err := redisotel.InstrumentTracing(client); err != nil {
			return nil, err
		}
	}
	if options.GetEnableMetrics().GetValue() {
		if err := redisotel.InstrumentMetrics(client); err != nil {
			return nil, err
		}
	}
	return client, nil
}
