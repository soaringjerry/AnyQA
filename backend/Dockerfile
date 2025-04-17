# backend/Dockerfile

# ---- Build Stage ----
# 使用官方 Go Alpine 镜像作为构建环境，匹配 go.mod 文件
FROM golang:1.22.0-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 Go 模块文件
COPY go.mod go.sum ./
# 下载依赖项。利用 Docker 缓存，只有在 go.mod/go.sum 变化时才重新下载
RUN go mod download

# 复制后端源代码
COPY . .

# 构建 Go 应用，禁用 CGO，创建静态链接的可执行文件
# 将可执行文件命名为 'server'
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./main.go

# ---- Run Stage ----
# 使用一个非常小的基础镜像
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/server .

# （可选）如果你的应用需要读取配置文件或其他静态资源，也需要在这里复制
# 例如: COPY --from=builder /app/config/config.go.example ./config/config.go.example
# 注意：生产环境通常通过环境变量注入配置，而不是复制示例文件

# 暴露后端服务监听的端口
EXPOSE 8080

# 设置容器启动时执行的命令
ENTRYPOINT ["./server"]