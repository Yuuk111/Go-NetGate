package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Yuuk111/Go-NetGate/internal/config"
	"github.com/Yuuk111/Go-NetGate/internal/gmtls"
	"github.com/Yuuk111/Go-NetGate/internal/proxy"
	myserver "github.com/Yuuk111/Go-NetGate/internal/server"
	"github.com/Yuuk111/Go-NetGate/internal/waf"
	"github.com/Yuuk111/Go-NetGate/internal/waf/limit"
	"github.com/redis/go-redis/v9"

	tjgmtls "github.com/tjfoc/gmsm/gmtls"
)

func main() {
	//===========================
	// 创建监听系统信号的 Context
	// 当接收到 SIGINT (Ctrl+C) 或 SIGTERM 信号时, ctx 会瞬间自动 cancle()，触发 ctx.Done()，从而通知所有使用这个 ctx 的协程安全退出
	// 这种方式比传统的 signal.Notify + channel 更加优雅和安全，避免了忘记关闭 channel 导致的资源泄露问题
	//===========================
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop() //退出时释放资源

	//===========================
	//Go 语言铁律：只要你看到 WithTimeout、WithCancel、NotifyContext
	//下面必须紧跟一行 defer cancel/stop()。谁污染，谁治理。
	//===========================

	// 1. 加载配置
	cmdConfig, err := config.LoadFileConfig()
	if err != nil {
		log.Fatalf("❌ [Config] 配置加载失败: %v", err)
	}

	// 初始化 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "",
		Password: "",
		DB:       0,
		PoolSize: 100,
	})
	// 2. 读取配置项
	ListenPort := ":" + cmdConfig.Server.ListenPort    // 读取监听端口
	TargetURLs := cmdConfig.LoadBalance.Backends       // 读取后端服务器列表
	LoadBalanceAlgo := cmdConfig.LoadBalance.Algorithm // 读取负载均衡算法
	TLSMode := cmdConfig.Server.TLSMode                // 读取 TLS 模式

	// 3. 初始化国密 TLS 配置
	var gmConfig *tjgmtls.Config
	if TLSMode == "gmtls" {
		gmConfig, err = gmtls.LoadGMTLSConfig(cmdConfig.Gm.SignCertFile, cmdConfig.Gm.SignKeyFile, cmdConfig.Gm.EncCertFile, cmdConfig.Gm.EncKeyFile)
		if err != nil {
			log.Fatalf("❌ [Config] 国密证书配置加载失败: %v", err)
		}
	}

	// 4. 创建反向代理
	// proxy, err := proxy.NewReverseProxy(TargetURL)
	proxyhandler, err := proxy.NewBalancedReverseProxy(ctx, LoadBalanceAlgo, TargetURLs)
	if err != nil {
		log.Fatalf("❌ [Server] 反向代理初始化失败: %v", err)
	}

	// 5. 组装处理链：WAF -> Proxy
	// 创建限流器实例，后续可以加config支持动态调整限流参数
	rateLimiter := limit.NewIPRateLimiter(5, 10) //每个IP每秒最多5个请求，令牌桶容量为10

	// 使用 WAF 中间件包装 proxy
	handler := rateLimiter.RateLimitMiddleware(waf.WafMiddleware(proxyhandler))

	// 6. 启动服务
	log.Printf("✅ [Server] Go语言国密WAF启动，监听 %s，转发至 %#v", ListenPort, TargetURLs)
	if err := myserver.StartServer(
		ctx,
		ListenPort,
		TLSMode,
		gmConfig,
		cmdConfig.Tls.CertFile,
		cmdConfig.Tls.KeyFile,
		handler); err != nil {
		log.Fatalf("❌ [Server] 服务遇到错误: %v", err)
	}
}
