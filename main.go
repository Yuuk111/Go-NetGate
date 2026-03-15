package main

import (
	"log"

	"github.com/Yuuk111/Go-NetGate/internal/config"
	"github.com/Yuuk111/Go-NetGate/internal/gmtls"
	"github.com/Yuuk111/Go-NetGate/internal/proxy"
	myserver "github.com/Yuuk111/Go-NetGate/internal/server"
	"github.com/Yuuk111/Go-NetGate/internal/waf"
)

// 配置常量
const (
	SignCertFile = "./certs/SS.crt"
	SignKeyFile  = "./certs/SS.key"
	EncCertFile  = "./certs/SE.crt"
	EncKeyFile   = "./certs/SE.key"
)

func main() {
	// 1. 初始化国密 TLS 配置
	gmConfig, err := gmtls.LoadGMTLSConfig(SignCertFile, SignKeyFile, EncCertFile, EncKeyFile)
	if err != nil {
		log.Fatalf("国密证书配置加载失败: %v", err)
	}

	//获取后端地址
	cmdConfig := config.LoadConfig()
	TargetURL := cmdConfig.TargetURL
	TargetURLs := cmdConfig.BackendServers
	ListenPort := ":" + cmdConfig.ListenPort

	// 2. 创建反向代理
	// proxy, err := proxy.NewReverseProxy(TargetURL)
	proxy, err := proxy.NewBalancedReverseProxy(TargetURLs)
	if err != nil {
		log.Fatalf("反向代理初始化失败: %v", err)
	}

	// 3. 组装处理链：WAF -> Proxy
	// 使用 WAF 中间件包装 proxy
	handler := waf.WafMiddleware(proxy)

	// 4. 启动服务
	log.Printf("Go语言国密WAF启动，监听 %s，转发至 %s", ListenPort, TargetURL)
	if err := myserver.StartServer(ListenPort, gmConfig, handler); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
