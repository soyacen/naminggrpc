// Package nacosgrpc 实现了基于 Nacos 的 gRPC 命名解析器
package nacosgrpc

import (
	"cmp"
	"net"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

// scheme 定义了该解析器的协议方案名称
const scheme = "nacos"

// ResolverDSNParser 定义了解析器 DSN 的解析函数
var ResolverDSNParser = parseResolverDSN

// 初始化函数，在包被导入时自动注册解析器构建器
func init() {
	resolver.Register(&nacosResolverBuilder{})
}

// nacosResolverBuilder 是 Nacos 解析器的构建器结构体
type nacosResolverBuilder struct{}

// Build 根据目标地址创建一个新的解析器实例
// 参数:
//   - tgt: 目标地址信息
//   - cc: 客户端连接接口
//   - opts: 构建选项
//
// 返回值:
//   - resolver.Resolver: 解析器实例
//   - error: 错误信息
func (b *nacosResolverBuilder) Build(tgt resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	// 解析 DSN 配置
	parsed, err := ResolverDSNParser(tgt.URL)
	if err != nil {
		return nil, errors.Wrapf(err, "naming: failed to parse dsn")
	}

	// 创建 Nacos 命名客户端
	client, err := clients.NewNamingClient(parsed.ClientParam)
	if err != nil {
		return nil, errors.Wrapf(err, "naming: failed to create naming client")
	}

	// 创建解析器实例
	r := &nacosResolver{
		target:       tgt,
		cc:           cc,
		namingClient: client,
		param:        parsed.SubscribeParam,
	}

	// 启动解析器
	if err := r.start(); err != nil {
		return nil, err
	}

	return r, nil
}

// Scheme 返回该解析器支持的协议方案
func (b *nacosResolverBuilder) Scheme() string {
	return scheme
}

// nacosResolver 是 Nacos 解析器的具体实现
type nacosResolver struct {
	target       resolver.Target             // 目标地址信息
	cc           resolver.ClientConn         // 客户端连接接口
	namingClient naming_client.INamingClient // Nacos 命名客户端
	param        vo.SubscribeParam           // 订阅参数
}

// start 启动解析器，进行初始服务发现和订阅
func (r *nacosResolver) start() error {
	// 初始获取服务实例列表
	instances, err := r.namingClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: r.param.ServiceName,
		GroupName:   r.param.GroupName,
		Clusters:    r.param.Clusters,
		HealthyOnly: true,
	})
	if err != nil {
		return errors.Wrapf(err, "naming: failed to select instances")
	}

	// 更新连接状态
	r.updateState(instances)

	// 设置订阅回调函数
	r.param.SubscribeCallback = func(services []model.Instance, err error) {
		if err != nil {
			return
		}
		r.updateState(services)
	}

	// 订阅服务变化
	if err := r.namingClient.Subscribe(&r.param); err != nil {
		return errors.Wrapf(err, "naming: failed to subscribe")
	}
	return nil
}

// updateState 更新解析器状态，将服务实例转换为 gRPC 地址格式
// 参数:
//   - instances: Nacos 服务实例列表
func (r *nacosResolver) updateState(instances []model.Instance) {
	addrs := make([]resolver.Address, 0, len(instances))

	// 遍历所有实例，过滤健康且启用的实例
	for _, inst := range instances {
		if !inst.Healthy || !inst.Enable {
			continue
		}

		// 构造 gRPC 地址
		addr := resolver.Address{
			Addr: net.JoinHostPort(inst.Ip, strconv.FormatUint(inst.Port, 10)),
		}

		// 添加实例属性信息
		addr.Attributes = attributes.New("ServiceName", inst.ServiceName).
			WithValue("weight", inst.Weight).
			WithValue("cluster", inst.ClusterName).
			WithValue("ephemeral", inst.Ephemeral)

		// 添加自定义元数据
		for k, v := range inst.Metadata {
			addr.Attributes = addr.Attributes.WithValue(k, v)
		}

		addrs = append(addrs, addr)
	}

	// 按地址排序确保一致性
	slices.SortFunc(addrs, func(a, b resolver.Address) int {
		return cmp.Compare(a.Addr, b.Addr)
	})

	// 更新客户端连接状态
	r.cc.UpdateState(resolver.State{Addresses: addrs})
}

// ResolveNow 立即解析目标地址（当前实现为空操作）
func (r *nacosResolver) ResolveNow(resolver.ResolveNowOptions) {}

// Close 关闭解析器，取消订阅并关闭客户端
func (r *nacosResolver) Close() {
	_ = r.namingClient.Unsubscribe(&r.param)
	r.namingClient.CloseClient()
}

// ResolverDSN 包含解析器所需的配置信息
type ResolverDSN struct {
	ClientParam    vo.NacosClientParam // Nacos 客户端参数
	SubscribeParam vo.SubscribeParam   // 订阅参数
}

// parseResolverDSN 解析 URL 格式的 DSN 配置
// 支持的 URL 格式: nacos://[username[:password]@]host[:port]/service_name?param=value
// 参数:
//   - u: URL 对象
//
// 返回值:
//   - *ResolverDSN: 解析后的配置对象
//   - error: 错误信息
func parseResolverDSN(u url.URL) (*ResolverDSN, error) {
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

	clusters := q.Get("clusters")

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

	// 返回解析结果
	return &resolverDSN{
		ClientParam: vo.NacosClientParam{
			ClientConfig:  clientConfig,
			ServerConfigs: serverConfigs,
		},
		SubscribeParam: vo.SubscribeParam{
			ServiceName: serviceName,
			GroupName:   group,
			Clusters:    strings.Split(clusters, ","),
		},
	}, nil
}
