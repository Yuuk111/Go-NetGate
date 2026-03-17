package server

import (
	"log"
	"net/http"

	"github.com/tjfoc/gmsm/gmtls"
)

// startServer 创建监听器并运行 HTTP Server
func StartServer(port string, tlsmode string, gmConfig *gmtls.Config, stdCert string, stdKey string, handler http.Handler) error {
	if tlsmode == "gmtls" {
		// 国密模式
		log.Printf("🛡️  启用国密 GmTLS 模式监听端口 %s", port)
		ln, err := gmtls.Listen("tcp", port, gmConfig)
		if err != nil {
			return err
		}
		defer ln.Close() //申请到系统资源后，确保在函数退出时释放
		server := &http.Server{
			Handler: handler,
		}
		return server.Serve(ln)
	} else {
		// 标准 TLS 模式
		log.Printf("🛡️  启用标准 TLS 模式监听端口 %s", port)
		server := &http.Server{
			Addr:    port,
			Handler: handler,
		}
		return server.ListenAndServeTLS(stdCert, stdKey)
	}
}
