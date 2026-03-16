package proxy

import (
	"net"
	"net/http"
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
