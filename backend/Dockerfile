# 使用多阶段构建
# 第一阶段：构建阶段
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装依赖
RUN apk add --no-cache git

# 复制go mod和sum文件
COPY go.mod go.sum ./

# 下载所有依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd

# 第二阶段：运行阶段
FROM alpine:latest

# 安装基本工具和CA证书
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为亚洲/上海
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
  echo "Asia/Shanghai" > /etc/timezone

# 创建非root用户
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 复制配置文件和其他必要文件
COPY --from=builder /app/config ./config
COPY --from=builder /app/docs ./docs

# 创建数据目录和日志目录
RUN mkdir -p /app/data /app/log && \
  chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露API端口
EXPOSE 25610

# 运行应用
CMD ["./main"]
