package proxy

import (
	"context"
	"log"
	"net"
	"net/http/httputil"
	"net/url"

	"github.com/Yuuk111/Go-NetGate/internal/proxy/loadbalancing"
)

// NewBalancedReverseProxy 创建一个支持负载均衡的反向代理
func NewBalancedReverseProxy(ctx context.Context, algo string, targets []string) (*httputil.ReverseProxy, error) {
	//预处理，把字符串转换为 url.URL 结构体
	targetURLs := make([]*url.URL, 0, len(targets)) //申请长度为 len(targets) 的切片，元素类型是 *url.URL
	//len()怎么处理的？
	for _, target := range targets {
		u, err := url.Parse(target)
		if err != nil {
			panic("无效的后端地址： " + target)
		}
		targetURLs = append(targetURLs, u)
	}
	//实例化负载均衡器
	lb := loadbalancing.NewLoadBalancer(ctx, algo, targetURLs)
	//创建自定义 Transport
	customTransport := SetTransport()
	//创建自定义错误处理器
	customErrorHandler := SetErrorHandler()
	//Rewrite 重写发送向后端的请求
	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			//获取客户端 IP 地址
			clientIP, _, err := net.SplitHostPort(pr.In.RemoteAddr)
			if err != nil { //如果无法解析客户端 IP 地址，记录日志但继续处理请求
				log.Printf("无法解析客户端 IP 地址: %v", err)
				return
			}
			target := lb.Next(clientIP) //从负载均衡器获取下一个后端服务器
			// 2. 魔法方法：自动重写请求的目的地
			// 它会自动把 pr.Out 的 Scheme, Host 替换为 target 的
			// 并且会安全地拼接 Path 和 RawQuery
			if target == nil {
				log.Printf("❌ [Load Balancer] 没有可用的后端服务器，无法处理请求！")
				pr.Out.URL.Scheme = "http" //默认使用 http 协议，虽然这个值不会被真正使用，因为后续的 Transport 会直接返回错误响应
				pr.Out.URL.Host = "0.0.0.0"
				return
			}

			pr.SetURL(target)
			// 3. 魔法方法：自动设置 X-Forwarded-For, X-Forwarded-Host, X-Forwarded-Proto 等头部
			pr.SetXForwarded()
			pr.Out.Header.Set("X-Real-IP", clientIP) //设置 X-Real-IP 头部为客户端 IP 地址

			pr.Out.Header.Set("X-WAF-Protected", "true")

		},
		//自定义连接池和超时设置 详见 SetTransport() 函数
		Transport:    customTransport,
		ErrorHandler: customErrorHandler,
	}
	return proxy, nil
}

// 创建一个反向代理
func NewReverseProxy(target string) (*httputil.ReverseProxy, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(targetURL)
			pr.Out.Header.Set("X-WAF-Protected", "true")
		},
	}
	return proxy, nil
}
