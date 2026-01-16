package nacosx

import (
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/soyacen/goconc/lazyload"
)

func NewConfigClients(config *Config) *lazyload.Group[config_client.IConfigClient] {
	return &lazyload.Group[config_client.IConfigClient]{
		New: func(key string) (config_client.IConfigClient, error) {
			configs := config.GetConfigs()
			options, ok := configs[key]
			if !ok {
				return nil, fmt.Errorf("nacos %s not found", key)
			}
			return NewConfigClient(options)
		},
	}
}

func NewNamingClients(config *Config) *lazyload.Group[naming_client.INamingClient] {
	return &lazyload.Group[naming_client.INamingClient]{
		New: func(key string) (naming_client.INamingClient, error) {
			configs := config.GetConfigs()
			options, ok := configs[key]
			if !ok {
				return nil, fmt.Errorf("nacos %s not found", key)
			}
			return NewNamingClient(options)
		},
	}
}

func NewConfigClient(options *Options) (config_client.IConfigClient, error) {
	sc := []constant.ServerConfig{
		*constant.NewServerConfig(options.GetAddress().GetValue(), uint64(options.GetPort().GetValue())),
	}
	cc := *constant.NewClientConfig(
		constant.WithNamespaceId(options.GetNamespace().GetValue()),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogLevel("warn"),
	)
	nacosClientParam := vo.NacosClientParam{
		ClientConfig:  &cc,
		ServerConfigs: sc,
	}
	return clients.NewConfigClient(nacosClientParam)
}

func NewNamingClient(options *Options) (naming_client.INamingClient, error) {
	cc := constant.NewClientConfig(
		constant.WithNamespaceId(options.GetNamespace().GetValue()),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithUpdateCacheWhenEmpty(true),
	)
	sc := []constant.ServerConfig{
		*constant.NewServerConfig(options.GetAddress().GetValue(), uint64(options.GetPort().GetValue())),
	}
	clientParam := vo.NacosClientParam{
		ClientConfig:  cc,
		ServerConfigs: sc,
	}
	return clients.NewNamingClient(clientParam)
}
