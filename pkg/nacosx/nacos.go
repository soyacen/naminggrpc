package nacosx

import (
	"github.com/cockroachdb/errors"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/soyacen/gox/conc/lazyload"
)

func (options *Options) AsNacosClientParam() vo.NacosClientParam {
	serverConfig := constant.NewServerConfig(
		options.GetIpAddr().GetValue(),
		options.GetPort().GetValue(),
	)
	if options.GetScheme() != nil {
		serverConfig.Scheme = options.GetScheme().GetValue()
	}
	if options.GetContextPath() != nil {
		serverConfig.ContextPath = options.GetContextPath().GetValue()
	}
	if options.GetGrpcPort() != nil {
		serverConfig.GrpcPort = options.GetGrpcPort().GetValue()
	}
	serverConfigs := []constant.ServerConfig{*serverConfig}

	clientConfig := constant.NewClientConfig()
	// 设置客户端配置
	if options.GetTimeoutMs() != nil {
		clientConfig.TimeoutMs = options.GetTimeoutMs().GetValue()
	}
	if options.GetBeatInterval() != nil {
		clientConfig.BeatInterval = options.GetBeatInterval().GetValue()
	}
	if options.GetNamespace() != nil {
		clientConfig.NamespaceId = options.GetNamespace().GetValue()
	}
	if options.GetAppName() != nil {
		clientConfig.AppName = options.GetAppName().GetValue()
	}
	if options.GetAppKey() != nil {
		clientConfig.AppKey = options.GetAppKey().GetValue()
	}
	if options.GetEndpoint() != nil {
		clientConfig.Endpoint = options.GetEndpoint().GetValue()
	}
	if options.GetRegionId() != nil {
		clientConfig.RegionId = options.GetRegionId().GetValue()
	}
	if options.GetAccessKey() != nil {
		clientConfig.AccessKey = options.GetAccessKey().GetValue()
	}
	if options.GetSecretKey() != nil {
		clientConfig.SecretKey = options.GetSecretKey().GetValue()
	}
	if options.GetCacheDir() != nil {
		clientConfig.CacheDir = options.GetCacheDir().GetValue()
	}
	if options.GetDisableUseSnapshot() != nil {
		clientConfig.DisableUseSnapShot = options.GetDisableUseSnapshot().GetValue()
	}
	if options.GetUpdateThreadNum() != nil {
		clientConfig.UpdateThreadNum = int(options.GetUpdateThreadNum().GetValue())
	}
	if options.GetNotLoadCacheAtStart() != nil {
		clientConfig.NotLoadCacheAtStart = options.GetNotLoadCacheAtStart().GetValue()
	}
	if options.GetUpdateCacheWhenEmpty() != nil {
		clientConfig.UpdateCacheWhenEmpty = options.GetUpdateCacheWhenEmpty().GetValue()
	}
	if options.GetUsername() != nil {
		clientConfig.Username = options.GetUsername().GetValue()
	}
	if options.GetPassword() != nil {
		clientConfig.Password = options.GetPassword().GetValue()
	}
	if options.GetLogDir() != nil {
		clientConfig.LogDir = options.GetLogDir().GetValue()
	}
	if options.GetLogLevel() != nil {
		clientConfig.LogLevel = options.GetLogLevel().GetValue()
	}
	if options.GetAppendToStdout() != nil {
		clientConfig.AppendToStdout = options.GetAppendToStdout().GetValue()
	}
	if options.GetTlsCfg() != nil {
		clientConfig.TLSCfg = constant.TLSConfig{
			Enable:             true,
			CaFile:             options.GetTlsCfg().GetCaFile().GetValue(),
			CertFile:           options.GetTlsCfg().GetCertFile().GetValue(),
			KeyFile:            options.GetTlsCfg().GetKeyFile().GetValue(),
			ServerNameOverride: options.GetTlsCfg().GetServerName().GetValue(),
		}
	}
	if options.GetAsyncUpdateService() != nil {
		clientConfig.AsyncUpdateService = options.GetAsyncUpdateService().GetValue()
	}
	if options.GetEndpointContextPath() != nil {
		clientConfig.EndpointContextPath = options.GetEndpointContextPath().GetValue()
	}
	if options.GetEndpointQueryParams() != nil {
		clientConfig.EndpointQueryParams = options.GetEndpointQueryParams().GetValue()
	}
	if options.GetClusterName() != nil {
		clientConfig.ClusterName = options.GetClusterName().GetValue()
	}
	if options.GetAppConnLabels() != nil {
		clientConfig.AppConnLabels = options.GetAppConnLabels()
	}

	return vo.NacosClientParam{ClientConfig: clientConfig, ServerConfigs: serverConfigs}
}

func NewConfigClients(config *Config) *lazyload.Group[config_client.IConfigClient] {
	return &lazyload.Group[config_client.IConfigClient]{
		New: func(key string) (config_client.IConfigClient, error) {
			configs := config.GetConfigs()
			options, ok := configs[key]
			if !ok {
				return nil, errors.Errorf("nacosx: config %s not found", key)
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
				return nil, errors.Errorf("nacosx: nacos %s not found", key)
			}
			return NewNamingClient(options)
		},
	}
}

func NewConfigClient(options *Options) (config_client.IConfigClient, error) {
	return clients.NewConfigClient(options.AsNacosClientParam())
}

func NewNamingClient(options *Options) (naming_client.INamingClient, error) {
	return clients.NewNamingClient(options.AsNacosClientParam())
}
