# 使用一个基础镜像，这里选择的是官方的 Go 镜像
FROM golang:1.18 AS builder

# 设置工作目录
WORKDIR /app

# 将应用程序的源代码复制到容器中
COPY go.mod ./
RUN go mod download
COPY *.go ./

# 编译应用程序
RUN CGO_ENABLED=0 GOOS=linux go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /nginxbackend-autoreport

# 使用一个更小的镜像作为最终镜像
FROM alpine:latest

# 将编译好的应用程序复制到最终镜像中
COPY --from=builder /nginxbackend-autoreport /usr/local/bin/nginxbackend-autoreport

# 启动应用程序
CMD ["nginxbackend-autoreport"]
