package server

import (
	"net/http"

	"github.com/tjfoc/gmsm/gmtls"
)

func StartServer(port string, gmConfig *gmtls.Config, handler http.Handler) error {
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
