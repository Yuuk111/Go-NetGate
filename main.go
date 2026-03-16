package main

import (
	"log"

	"github.com/Yuuk111/Go-NetGate/internal/config"
	"github.com/Yuuk111/Go-NetGate/internal/gmtls"
	"github.com/Yuuk111/Go-NetGate/internal/proxy"
	myserver "github.com/Yuuk111/Go-NetGate/internal/server"
	"github.com/Yuuk111/Go-NetGate/internal/waf"
)

func main() {
	// 1. 加载配置
	cmdConfig, err := config.LoadFileConfig()
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}
	// 2. 初始化国密 TLS 配置
	gmConfig, err := gmtls.LoadGMTLSConfig(cmdConfig.Gm.SignCertFile, cmdConfig.Gm.SignKeyFile, cmdConfig.Gm.EncCertFile, cmdConfig.Gm.EncKeyFile)
	if err != nil {
		log.Fatalf("国密证书配置加载失败: %v", err)
	}

	// 3. 读取配置项
	ListenPort := ":" + cmdConfig.Server.ListenPort    // 读取监听端口
	TargetURLs := cmdConfig.LoadBalance.Backends       // 读取后端服务器列表
	LoadBalanceAlgo := cmdConfig.LoadBalance.Algorithm // 读取负载均衡算法

	// 4. 创建反向代理
	// proxy, err := proxy.NewReverseProxy(TargetURL)
	proxy, err := proxy.NewBalancedReverseProxy(LoadBalanceAlgo, TargetURLs)
	if err != nil {
		log.Fatalf("反向代理初始化失败: %v", err)
	}

	// 5. 组装处理链：WAF -> Proxy
	// 使用 WAF 中间件包装 proxy
	handler := waf.WafMiddleware(proxy)

	// 6. 启动服务
	log.Printf("Go语言国密WAF启动，监听 %s，转发至 %#v", ListenPort, TargetURLs)
	if err := myserver.StartServer(ListenPort, gmConfig, handler); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
