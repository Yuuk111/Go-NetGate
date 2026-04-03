# 编译
FROM golang:1.26.1-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# 构建可执行文件
RUN go build -ldflags="-s -w" -o netgate main.go

# 运行阶段
FROM alpine:latest

# 安装 tzdata 和 ca-certificates，设置时区为上海
RUN apk --no-cache add tzdata ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

COPY --from=builder /builder/netgate .

EXPOSE 8443

CMD ["./netgate"]