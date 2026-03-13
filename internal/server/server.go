package server

import (
	"net/http"

	"github.com/tjfoc/gmsm/gmtls"
)

// startServer 创建监听器并运行 HTTP Server
func StartServer(port string, gmConfig *gmtls.Config, handler http.Handler) error {
	//创建国密监听器，处理国密握手和加密通信
	ln, err := gmtls.Listen("tcp", port, gmConfig)
	if err != nil {
		return err
	}
	defer ln.Close() //申请到系统资源后，确保在函数退出时释放

	server := &http.Server{
		Handler: handler,
	}
	return server.Serve(ln)
}
