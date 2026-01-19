# 使用官方Go镜像作为基础镜像
FROM golang:1.25-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum文件（如果存在）到工作目录
COPY go.mod go.sum ./

# 下载项目依赖
RUN go mod download

# 将源代码复制到容器中
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o project .

# 使用轻量级alpine镜像作为最终镜像
FROM alpine:latest

# 安装ca-certificates，如果应用需要连接HTTPS服务
RUN apk --no-cache add ca-certificates

# 设置时区（可选）
RUN apk add --no-cache tzdata

# 创建非root用户运行应用
RUN addgroup -g 65532 nonroot &&\
    adduser -D -u 65532 -G nonroot nonroot

# 设置工作目录
WORKDIR /root/

# 从builder镜像中复制编译好的二进制文件
COPY --from=builder /app/project .

# 更改文件权限给nonroot用户
RUN chown -R nonroot:nonroot /root/
USER nonroot

# 暴露端口（如果应用需要监听端口的话）
# EXPOSE 8080

# 启动应用
CMD ["./project"]