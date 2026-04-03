package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Yuuk111/Go-NetGate/internal/config"
	"github.com/Yuuk111/Go-NetGate/internal/gmtls"
	"github.com/Yuuk111/Go-NetGate/internal/insight"
	"github.com/Yuuk111/Go-NetGate/internal/proxy"
	"github.com/Yuuk111/Go-NetGate/internal/proxy/router"
	"github.com/Yuuk111/Go-NetGate/internal/waf"
	"github.com/Yuuk111/Go-NetGate/internal/waf/limit"
	"github.com/redis/go-redis/v9"

	myserver "github.com/Yuuk111/Go-NetGate/internal/server"
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

	// 加载配置
	cmdConfig, err := config.LoadFileConfig()
	if err != nil {
		log.Fatalf("❌ [Config] 配置加载失败: %v", err)
	}

	// 初始化 Insight 异步上报引擎
	insightSender, err := insight.NewGRPCReporter(cmdConfig.InsightAgent.ServerAddr, cmdConfig.InsightAgent.BufferSize)
	if err != nil {
		log.Printf("❌ [Insight] 无法连接到日志分析服务器: %v \n", err)
	} else {
		log.Println("✅ [Insight] 安全审计日志引擎初始化成功")
		defer insightSender.Close() //确保在 main 函数退出时关闭 Insight 上报引擎
	}

	// 初始化 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cmdConfig.Redis.Addr,
		Password: cmdConfig.Redis.Password,
		DB:       cmdConfig.Redis.DB,
		PoolSize: cmdConfig.Redis.PoolSize,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("❌ [Redis] 无法连接到 Redis: %v, 网关启动失败", err)
	}
	defer rdb.Close() //确保在 main 函数退出时关闭 Redis 连接
	log.Println("✅ [Redis] 成功连接到 Redis")

	// 读取配置项
	ListenPort := ":" + cmdConfig.Server.ListenPort // 读取监听端口
	TLSMode := cmdConfig.Server.TLSMode             // 读取 TLS 模式

	// 初始化国密 TLS 配置
	var gmConfig *tjgmtls.Config
	if TLSMode == "gmtls" {
		gmConfig, err = gmtls.LoadGMTLSConfig(cmdConfig.Gm.SignCertFile, cmdConfig.Gm.SignKeyFile, cmdConfig.Gm.EncCertFile, cmdConfig.Gm.EncKeyFile)
		if err != nil {
			log.Fatalf("❌ [Config] 国密证书配置加载失败: %v", err)
		}
	}

	//===========================
	// 初始化智能路由引擎
	//===========================
	router := router.NewRouter()
	targetRoutes := make([]string, 0, len(cmdConfig.RouteRules))

	for _, routeRule := range cmdConfig.RouteRules {
		// 将每条路由规则添加到路由引擎中
		proxyHandler, err := proxy.NewBalancedReverseProxy(ctx, routeRule.Algorithm, routeRule.Backends)
		if err != nil {
			log.Fatalf("❌ [Router] 路由规则 %s 初始化失败: %v", routeRule.Path, err)
		}
		targetRoutes = append(targetRoutes, routeRule.Path)
		router.AddRoute(routeRule.Path, proxyHandler)
		log.Printf("✅ [Router] 路由规则添加成功: 前缀=%s, 算法=%s, 后端=%#v", routeRule.Path, routeRule.Algorithm, routeRule.Backends)
	}

	// // 创建反向代理
	// // proxy, err := proxy.NewReverseProxy(TargetURL)
	// proxyhandler, err := proxy.NewBalancedReverseProxy(ctx, LoadBalanceAlgo, TargetURLs)
	// if err != nil {
	// 	log.Fatalf("❌ [Server] 反向代理初始化失败: %v", err)
	// }

	// 组装处理链：限流器 -> WAF -> ->Agent -> 反向代理
	// 创建单机限流器实例，后续可以加config支持动态调整限流参数
	// rateLimiter := limit.NewIPRateLimiter(cmdConfig.SingleRateLimit.Rate, cmdConfig.SingleRateLimit.Burst) //每个IP每秒最多5个请求，令牌桶容量为10

	// 初始化 Insight 中间件
	insightMW := waf.InsightMiddleware(insightSender)

	// 创建分布式 Redis 限流器实例，后续可以加config支持动态调整限流参数
	redisRateLimiter := limit.NewRedisRateLimiter(rdb, cmdConfig.RedisRateLimit.Rate, cmdConfig.RedisRateLimit.Burst) //每个IP每秒最多5个请求，令牌桶容量为10
	// 使用洋葱模型组装中间件链
	var handler http.Handler
	if insightSender != nil {
		handler = redisRateLimiter.RedisRateLimitMiddleware(waf.WafMiddleware(insightMW(router)))
	} else {
		handler = redisRateLimiter.RedisRateLimitMiddleware(waf.WafMiddleware(router))
	}

	// 启动服务
	log.Printf("✅ [Server] Go语言国密WAF启动，监听 %s，转发至路由 %#v", ListenPort, targetRoutes)
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
