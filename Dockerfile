# 使用官方 Go 镜像来构建应用程序
FROM golang:1.23 AS builder

# 设置工作目录
WORKDIR /app

# 复制模块依赖文件并安装依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制项目文件
COPY . .

# 构建应用程序
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 使用 Alpine 作为基础镜像，减小镜像体积
FROM alpine:latest

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制应用程序
COPY --from=builder /app/main .
COPY --from=builder /app/templates ./templates

# 暴露端口
EXPOSE 8080

# 运行应用程序
CMD ["./main"]
