// Package nacosgrpc 实现了基于 Nacos 的 gRPC 服务注册器
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

var RegistrarDSNParser = parseRegistrarDSN

func init() {
	naminggrpc.Register("nacos")
}

// Registrar 是 Nacos 服务注册器结构体
type Registrar struct {
	namingClient    naming_client.INamingClient // Nacos 命名客户端
	registerParam   vo.RegisterInstanceParam    // 注册实例参数
	deregisterParam vo.DeregisterInstanceParam  // 注销实例参数
}

// Register 在 Nacos 中注册服务实例
// 参数:
//   - ctx: 上下文对象
//
// 返回值:
//   - error: 注册过程中的错误信息
func (r *Registrar) Register(ctx context.Context) error {
	// 调用 Nacos 客户端注册实例
	ok, err := r.namingClient.RegisterInstance(r.registerParam)
	if err != nil {
		return errors.Wrapf(err, "naming: failed to register %s", r.registerParam.ServiceName)
	}

	// 检查注册是否成功
	if !ok {
		return errors.Errorf("naming: failed to register %s", r.registerParam.ServiceName)
	}

	return nil
}

// Deregister 从 Nacos 中注销服务实例
// 参数:
//   - ctx: 上下文对象
//
// 返回值:
//   - error: 注销过程中的错误信息
func (r *Registrar) Deregister(ctx context.Context) error {
	// 调用 Nacos 客户端注销实例
	ok, err := r.namingClient.DeregisterInstance(r.deregisterParam)
	if err != nil {
		return errors.Wrapf(err, "naming: failed to deregister %s", r.deregisterParam.ServiceName)
	}

	// 检查注销是否成功
	if !ok {
		return errors.Errorf("naming: failed to deregister %s", r.deregisterParam.ServiceName)
	}

	return nil
}

// NewNacosRegistrar 创建新的 Nacos 服务注册器实例
// 参数:
//   - dsn: 数据源名称，格式为 nacos://[username[:password]@]host[:port]/service_name?param=value
//
// 返回值:
//   - *Registrar: 注册器实例
//   - error: 创建过程中的错误信息
func NewRegistrar(dsn string) (*Registrar, error) {
	// 解析 DSN URL
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, errors.Wrapf(err, "naming: failed to parse dsn %s", dsn)
	}

	// 解析注册器 DSN 配置
	parsed, err := RegistrarDSNParser(*u)
	if err != nil {
		return nil, err
	}

	// 创建 Nacos 命名客户端
	client, err := clients.NewNamingClient(parsed.clientParam)
	if err != nil {
		return nil, errors.Wrapf(err, "naming: failed to create naming client")
	}

	// 返回注册器实例
	return &Registrar{
		namingClient:    client,
		registerParam:   parsed.registerParam,
		deregisterParam: parsed.deregisterParam,
	}, nil
}

// RegistrarDSN 包含注册器所需的配置信息
type RegistrarDSN struct {
	clientParam     vo.NacosClientParam        // Nacos 客户端参数
	registerParam   vo.RegisterInstanceParam   // 注册实例参数
	deregisterParam vo.DeregisterInstanceParam // 注销实例参数
}

// parseRegistrarDSN 解析注册器的 DSN 配置
// 支持的 URL 格式: nacos://[username[:password]@]host[:port]/service_name?param=value
// 参数:
//   - u: URL 对象
//
// 返回值:
//   - *nacosRegistrarDSN: 解析后的配置对象
//   - error: 错误信息
func parseRegistrarDSN(u url.URL) (*RegistrarDSN, error) {
	// 解析主机和端口
	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		host = u.Host
		portStr = "8848" // 默认 Nacos 端口
	}
	port, _ := strconv.ParseUint(portStr, 10, 64)

	// 解析查询参数
	q := u.Query()
	namespace := q.Get("namespace")
	if namespace == "" {
		namespace = "public" // 默认命名空间
	}

	group := q.Get("group")
	if group == "" {
		group = "DEFAULT_GROUP" // 默认分组
	}

	// 解析超时时间
	timeoutMs := uint64(10000) // 默认超时 10 秒
	if t := q.Get("timeout"); t != "" {
		if tv, err := strconv.ParseUint(t, 10, 64); err == nil {
			timeoutMs = tv
		}
	}

	// 构造客户端配置
	clientConfig := constant.NewClientConfig()
	clientConfig.Username = u.User.Username()
	clientConfig.Password, _ = u.User.Password()
	clientConfig.NamespaceId = namespace
	clientConfig.TimeoutMs = timeoutMs
	clientConfig.NotLoadCacheAtStart = true

	// 构造服务器配置
	serverConfigs := []constant.ServerConfig{
		*constant.NewServerConfig(host, port),
	}

	// 解析服务名称
	serviceName := strings.Trim(u.Path, "/")
	if serviceName == "" {
		return nil, errors.New("naming: service name is required in path")
	}

	// 解析实例 IP 和端口
	ip := q.Get("ip")
	regPortStr := q.Get("port")
	var regPort uint64
	if regPortStr != "" {
		regPort, _ = strconv.ParseUint(regPortStr, 10, 64)
	}

	// 解析权重配置
	weight := 10.0 // 默认权重
	if w := q.Get("weight"); w != "" {
		if wv, err := strconv.ParseFloat(w, 64); err == nil {
			weight = wv
		}
	}

	// 解析临时实例配置
	ephemeral := true // 默认为临时实例
	if e := q.Get("ephemeral"); e != "" {
		if ev, err := strconv.ParseBool(e); err == nil {
			ephemeral = ev
		}
	}

	// 解析集群名称和元数据
	cluster := q.Get("cluster")
	metadata := make(map[string]string)

	// 处理以 meta. 开头的元数据参数
	for k, v := range q {
		if strings.HasPrefix(k, "meta.") {
			metadata[strings.TrimPrefix(k, "meta.")] = v[0]
		}
	}

	// 返回解析结果
	return &RegistrarDSN{
		clientParam: vo.NacosClientParam{
			ClientConfig:  clientConfig,
			ServerConfigs: serverConfigs,
		},
		registerParam: vo.RegisterInstanceParam{
			Ip:          ip,
			Port:        regPort,
			Weight:      weight,
			Enable:      true, // 默认启用
			Healthy:     true, // 默认健康
			Metadata:    metadata,
			ClusterName: cluster,
			ServiceName: serviceName,
			GroupName:   group,
			Ephemeral:   ephemeral,
		},
		deregisterParam: vo.DeregisterInstanceParam{
			Ip:          ip,
			Port:        regPort,
			Cluster:     cluster,
			ServiceName: serviceName,
			GroupName:   group,
			Ephemeral:   ephemeral,
		},
	}, nil
}

type Factory struct{}

func (f *Factory) New(ctx context.Context, dsn string) (naminggrpc.Registrar, error) {
}
