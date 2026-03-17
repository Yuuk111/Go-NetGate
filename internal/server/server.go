package server

import (
	"log"
	"net/http"
	"time"

	"github.com/tjfoc/gmsm/gmtls"
)

// startServer 创建监听器并运行 HTTP Server
func StartServer(port string, tlsmode string, gmConfig *gmtls.Config, stdCert string, stdKey string, handler http.Handler) error {

	server := &http.Server{
		Addr:    port,
		Handler: handler,

		// 读取头部时的超时时间，防止慢速攻击
		ReadHeaderTimeout: 5 * time.Second, // 客户端必须在5秒发送完HTTP Header
		// 读取整个请求的超时时间，防止慢速攻击
		ReadTimeout: 10 * time.Second, // 客户端必须在10秒发送完HTTP Body
		// 写响应的超时时间，防止后端处理过慢导致资源占用
		WriteTimeout: 15 * time.Second, // 客户端必须在15秒内接收完响应，否则断开连接
		// 空闲连接的超时时间，防止资源泄露
		IdleTimeout: 90 * time.Second, // 连接空闲超过90秒就关闭
	}
	if tlsmode == "gmtls" {
		// 国密模式
		log.Printf("🛡️  启用国密 GmTLS 模式监听端口 %s", port)
		ln, err := gmtls.Listen("tcp", port, gmConfig)
		if err != nil {
			return err
		}
		defer ln.Close() //申请到系统资源后，确保在函数退出时释放
		return server.Serve(ln)
	} else {
		// 标准 TLS 模式
		log.Printf("🛡️  启用标准 TLS 模式监听端口 %s", port)
		return server.ListenAndServeTLS(stdCert, stdKey)
	}
}
