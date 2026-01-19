# grocer

`grocer` 是一个Go微服务基础设施库，提供了一系列常用的中间件和服务集成组件，帮助快速构建分布式应用。该库集成了数据库、缓存、消息队列、搜索引擎等多种服务的支持。

## 特性

- 统一的配置管理
- 依赖注入支持 (使用 Google Wire)
- 标准化的中间件集成
- OpenTelemetry 集成支持
- Protobuf 配置定义

## 组件

### dbx
数据库访问层，提供 MySQL 和 PostgreSQL 的统一接口。

### esx
Elasticsearch 集成，用于搜索和分析功能。

### goosex
提供 Goose 数据库迁移工具的集成。

### grpcx
gRPC 服务框架支持，包括客户端和服务端的便捷封装。

### idx
唯一 ID 生成器，基于 ULID 实现。

### jeagerx
Jaeger 分布式追踪系统集成，便于链路监控和调试。

### kafkax
Apache Kafka 消息队列集成，提供生产者和消费者支持。

### mongox
MongoDB 数据库集成，支持连接管理和查询封装。

### nacosx
Nacos 服务发现和配置中心集成。

### otelx
OpenTelemetry 集成，提供分布式追踪和指标收集功能。

### promx
Prometheus 指标收集集成。

### protobufx
Protobuf 工具类，提供 TLS 配置等通用消息定义。

### redisx
Redis 缓存集成，支持单机、集群和哨兵模式。

### registryx
服务注册与发现组件。

## 快速开始

### 安装

```bash
go get github.com/soyacen/grocer/pkg
```

### 使用示例

```go
package main

import (
    "github.com/soyacen/grocer/pkg/redisx"
    "github.com/soyacen/grocer/pkg/dbx"
)

func main() {
    // 使用 redisx 连接 Redis
    redisClient := redisx.NewClient(&redisx.Config{
        // 配置参数
    })
    
    // 使用 dbx 连接数据库
    db := dbx.Connect(&dbx.Config{
        // 配置参数
    })
    
    // ... 其他业务逻辑
}
```

## 架构设计

`grocer` 采用模块化设计，每个子包都是一个独立的功能模块。通过 Google Wire 实现依赖注入，使得组件间的耦合度更低，更易于测试和维护。

## 贡献

欢迎提交 Issue 和 Pull Request 来改进项目。

## 许可证

MIT License