# naminggrpc

[naminggrpc](https://github.com/soyacen/naminggrpc) 是一个基于 Nacos 的 gRPC 服务注册与发现库，提供了服务注册器和命名解析器的实现。

## 功能特性

- 🚀 基于 Nacos 的服务注册与发现
- 🔧 支持 gRPC 服务注册器 (Registrar)
- 🔍 支持 gRPC 命名解析器 (Resolver)
- ⚡ 自动服务发现和负载均衡
- 🛠️ 灵活的 DSN 配置方式
- 📦 易于集成和扩展

## 安装

```bash
go get github.com/soyacen/naminggrpc
```

## 快速开始

### 服务注册器使用示例

```go
package main

import (
    "context"
    "log"
    
    "github.com/soyacen/naminggrpc/nacosgrpc"
)

func main() {
    // 创建服务注册器
    registrar, err := nacosgrpc.NewRegistrar("nacos://localhost:8848/my-service?group=DEFAULT_GROUP&namespace=public")
    if err != nil {
        log.Fatal(err)
    }
    
    // 注册服务
    ctx := context.Background()
    if err := registrar.Register(ctx); err != nil {
        log.Fatal(err)
    }
    
    // 服务运行...
    
    // 注销服务
    if err := registrar.Deregister(ctx); err != nil {
        log.Fatal(err)
    }
}
```

### 命名解析器使用示例

```go
package main

import (
    "context"
    "log"
    "time"
    
    "google.golang.org/grpc"
    "github.com/soyacen/naminggrpc/nacosgrpc"
    pb "your-service-package" // 替换为你的 protobuf 生成包
)

func main() {
    // 使用 nacos scheme 连接服务
    conn, err := grpc.Dial("nacos://localhost:8848/my-service?group=DEFAULT_GROUP", 
        grpc.WithInsecure(),
        grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`))
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    // 创建客户端
    client := pb.NewYourServiceClient(conn)
    
    // 使用客户端调用服务
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    
    resp, err := client.YourMethod(ctx, &pb.YourRequest{})
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Response: %v", resp)
}
```

## DSN 配置格式

### 注册器 DSN 格式

```
nacos://[username[:password]@]host[:port]/service_name?param=value
```

### 解析器 DSN 格式

```
nacos://[username[:password]@]host[:port]/service_name?param=value
```

### 支持的参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `namespace` | 命名空间 ID | `public` |
| `group` | 服务分组 | `DEFAULT_GROUP` |
| `timeout` | 超时时间(毫秒) | `10000` |
| `ip` | 实例 IP 地址 | - |
| `port` | 实例端口号 | - |
| `weight` | 实例权重 | `10.0` |
| `ephemeral` | 是否为临时实例 | `true` |
| `cluster` | 集群名称 | - |
| `clusters` | 订阅的集群列表(逗号分隔) | - |
| `meta.*` | 自定义元数据 | - |

### 示例 DSN

```bash
# 基本配置
nacos://localhost:8848/my-service

# 带认证的配置
nacos://username:password@192.168.1.100:8848/my-service

# 完整配置
nacos://localhost:8848/my-service?namespace=dev&group=MY_GROUP&timeout=5000&weight=5.0&ephemeral=true&cluster=DEFAULT&meta.version=v1.0.0
```

## 接口设计

### Registrar 接口

```go
type Registrar interface {
    Register(ctx context.Context) error
    Deregister(ctx context.Context) error
}
```

### Factory 工厂模式

```go
type Factory interface {
    New(ctx context.Context, dsn string) (Registrar, error)
}
```

## 测试

运行单元测试：

```bash
go test ./...
```

运行特定包的测试：

```bash
go test ./nacosgrpc
```

## 依赖

- [nacos-sdk-go/v2](https://github.com/nacos-group/nacos-sdk-go) - Nacos Go SDK
- [grpc-go](https://github.com/grpc/grpc-go) - gRPC Go 实现

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 致谢

感谢以下开源项目：

- [Nacos](https://nacos.io/) - 动态服务发现、配置和服务管理平台
- [gRPC](https://grpc.io/) - Google 的高性能 RPC 框架