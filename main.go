package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/tjfoc/gmsm/gmtls"
)

// 配置常量
const (
	SignCertFile = "./certs/SS.crt"
	SignKeyFile  = "./certs/SS.key"
	EncCertFile  = "./certs/SE.crt"
	EncKeyFile   = "./certs/SE.key"
	ListenAddr   = ":8443"
)

func main() {
	// 1. 初始化国密 TLS 配置
	gmConfig, err := loadGMTLSConfig(SignCertFile, SignKeyFile, EncCertFile, EncKeyFile)
	if err != nil {
		log.Fatalf("国密证书配置加载失败: %v", err)
	}

	//获取后端地址
	var TargetURL string
	flag.StringVar(&TargetURL, "t", "http://localhost:9001", "后端目标服务器地址")
	flag.Parse()

	// 2. 创建反向代理
	proxy, err := newReverseProxy(TargetURL)
	if err != nil {
		log.Fatalf("反向代理初始化失败: %v", err)
	}

	// 3. 组装处理链：WAF -> Proxy
	// 使用 WAF 中间件包装 proxy
	handler := wafMiddleware(proxy)

	// 4. 启动服务
	log.Printf("Go语言国密WAF启动，监听 %s，转发至 %s", ListenAddr, TargetURL)
	if err := startServer(ListenAddr, gmConfig, handler); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// ---------------------------------------------------------------------
// 功能函数定义
// ---------------------------------------------------------------------

// loadGMTLSConfig 加载国密双证书并返回 TLS 配置
func loadGMTLSConfig(signCertPath, signKeyPath, encCertPath, encKeyPath string) (*gmtls.Config, error) {
	// 加载签名证书
	sigCert, err := gmtls.LoadX509KeyPair(signCertPath, signKeyPath)
	if err != nil {
		return nil, err
	}

	// 加载加密证书
	encCert, err := gmtls.LoadX509KeyPair(encCertPath, encKeyPath)
	if err != nil {
		return nil, err
	}

	// 配置 gmtls
	// 注意：Certificates 数组中，第一个必须是签名证书，第二个是加密证书
	config := &gmtls.Config{
		Certificates: []gmtls.Certificate{sigCert, encCert},
		GMSupport:    &gmtls.GMSupport{},
	}

	return config, nil
}

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

// wafMiddleware WAF 中间件
func wafMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 获取 URL 参数进行检测
		query := r.URL.Query().Encode()

		if containsEvil(query) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("WAF Blocked: Malicious Request Detected"))
			// 拦截后直接返回，不调用 next
			return
		}

		// 通过 WAF 检测，转交给下一个处理器（即 Proxy）
		next.ServeHTTP(w, r)
	})
}

// startServer 创建监听器并运行 HTTP Server
func startServer(addr string, gmConfig *gmtls.Config, handler http.Handler) error {
	// 创建国密监听器
	ln, err := gmtls.Listen("tcp", addr, gmConfig)
	if err != nil {
		return err
	}
	defer ln.Close()

	// 创建 HTTP 服务器
	server := &http.Server{
		Handler: handler,
	}

	return server.Serve(ln)
}

// containsEvil 简单的恶意特征检测函数
func containsEvil(query string) bool {
	// 定义恶意关键词列表
	evilKeywords := []string{"union", "select", "alert", "script", "../"}

	lowerQuery := strings.ToLower(query)
	for _, keyword := range evilKeywords {
		if strings.Contains(lowerQuery, keyword) {
			log.Printf("检测到攻击特征: %s", keyword)
			return true
		}
	}
	return false
}
