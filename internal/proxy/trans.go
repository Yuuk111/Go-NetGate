package proxy

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

func SetTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,  // TCP 握手超时
			KeepAlive: 30 * time.Second, // 操作系统级别的 TCP 连接保活机制，防止长时间不活跃的连接被关闭以及检测死连接
		}).DialContext,

		// 连接池配置：避免每次请求都重新 TCP 握手，提升性能
		MaxIdleConns:        1000,             // 全局最大空闲连接数
		MaxIdleConnsPerHost: 100,              // 关键 每个后端主机的最大空闲连接数
		IdleConnTimeout:     90 * time.Second, // 空闲连接的超时时间

		// 熔断与超时控制，保护网关不被拖死
		TLSHandshakeTimeout:   3 * time.Second,  // TLS 握手超时
		ExpectContinueTimeout: 3 * time.Second,  // 发送完请求后，等待后端 Continue 响应头的最大时间
		ResponseHeaderTimeout: 10 * time.Second, // 等待后端响应头的最大时间
	}
}

func SetErrorHandler() func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("[NetGate 降级] 请求 %s 转发后端失败: %v", r.URL.Path, err)

		// 准备返回给客户端的 HTTP 状态码和错误信息
		statusCode := http.StatusBadGateway // 默认502 后端宕机/拒绝连接
		errMsg := "NetGate: 后端服务不可用"

		// 简单错误分类，判断是否 Transport 触发超时
		if err != nil && (strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline")) {
			statusCode = http.StatusGatewayTimeout // 504 网关超时
			errMsg = "NetGate: 后端服务超时"
		}

		// 返回 JSON 格式的错误响应
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-NetGate-Error", "Degraded") // 塞一个网关自定义头
		w.WriteHeader(statusCode)

		jsonResponse := fmt.Sprintf(`{"code": %d, "error": "%s", "data": null}`, statusCode, errMsg)
		w.Write([]byte(jsonResponse))
	}
}
