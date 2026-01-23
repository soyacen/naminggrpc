# Grocer Nacos Resolver 实现总结

## 1. 自动注册
- 包导入时自动注册到 gRPC resolver 注册表
- 支持 `nacos://` URL scheme

## 2. 灵活配置
支持的 URL 参数：
- `namespace`: Nacos 命名空间 ID
- `group`: 服务分组（默认 DEFAULT_GROUP）
- `cluster`: 集群名称
- `healthy`: 是否仅返回健康实例
- `poll-interval`: 轮询间隔（默认 30s）
- `timeout`: 连接超时（默认 30s）

## 3. 完整的生命周期管理
- 创建 Nacos 客户端连接
- 定期轮询获取服务实例
- 自动更新 gRPC 连接池
- 优雅的资源清理

## 4. 错误处理和日志
- 通过 grpclog 输出所有错误和信息
- 连接失败时自动重试
- 详细的日志便于调试

## 使用方式

### 基础使用

```go
import _ "github.com/soyacen/grocer/pkg/grpcx/resolverx/nacos"

conn, err := grpc.NewClient("nacos://127.0.0.1:8848/my-service")
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

client := pb.NewMyServiceClient(conn)
```

### 完整示例

```go
conn, err := grpc.NewClient(
    "nacos://127.0.0.1:8848/user-service?" +
    "namespace=prod&" +
    "group=api&" +
    "cluster=beijing&" +
    "poll-interval=30s",
)
```
