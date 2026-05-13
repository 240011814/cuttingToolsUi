# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 安装 git
RUN apk add --no-cache git

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源码
COPY . .

# 静态编译
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Runtime stage
FROM alpine:latest

WORKDIR /app

# 复制可执行文件
COPY --from=builder /app/main .

# 赋予执行权限
RUN chmod +x ./main

# 暴露端口
EXPOSE 8080

# 启动
CMD ["./main"]