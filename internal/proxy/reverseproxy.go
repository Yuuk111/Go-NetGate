package proxy

import (
	"net/http/httputil"
	"net/url"
)

// 配置反向代理
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

/*
// 配置反向代理
func newReverseProxy(target string) (*httputil.ReverseProxy, error) {
	urlObj, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(urlObj)

	// 1. 保留并重写 Director (篡改请求头)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		remoteIP, _, _ := net.SplitHostPort(req.RemoteAddr)
		req.Header.Set("Host", req.Host)
		req.Header.Set("X-Real-IP", remoteIP)
		req.Header.Set("X-Forwarded-Proto", "https")
	}

	// 2. 核心：自定义底层发送器 (Transport)
	proxy.Transport = &http.Transport{
		// 限制和后端建立 TCP 连接的时间：最多等 3 秒
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,

		// 限制 TLS 握手时间 (如果后端也是 HTTPS)
		TLSHandshakeTimeout: 3 * time.Second,

		// 核心熔断点：发送完请求后，最多等后端 10 秒钟返回响应头！
		// 如果后端卡死超过 10 秒没理我，直接掐断，不干了！
		ResponseHeaderTimeout: 10 * time.Second,

		// 连接池配置：避免每次请求都重新 TCP 握手，提升性能
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}

	// 3. 错误兜底机制 (ErrorHandler)
	// 如果上面的 Transport 触发了超时，或者后端彻底宕机连不上，就会走到这里
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("[WAF 拦截] 后端目标服务器异常或超时: %v", err)

		// 优雅地给前端返回 504 状态码，而不是让客户端死等或者看到连接重置
		w.WriteHeader(http.StatusGatewayTimeout)
		w.Write([]byte("504 Gateway Timeout: The backend server is down or too slow."))
	}

	return proxy, nil
}
*/
