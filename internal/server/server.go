package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/tjfoc/gmsm/gmtls"
)

// startServer 创建监听器并运行 HTTP Server
func StartServer(ctx context.Context, port string, tlsmode string, gmConfig *gmtls.Config, stdCert string, stdKey string, handler http.Handler) error {

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

	errChan := make(chan error, 1) //创建一个错误通道，用于接收服务器运行过程中发生的错误
	// 启动服务器的协程，根据 TLS 模式选择不同的监听方式
	go func() {
		if tlsmode == "gmtls" {
			// 国密模式
			log.Printf("🛡️  启用国密 GmTLS 模式监听端口 %s", port)
			ln, err := gmtls.Listen("tcp", port, gmConfig)
			if err != nil {
				errChan <- err
				return
			}
			defer ln.Close() //申请到系统资源后，确保在函数退出时释放
			errChan <- server.Serve(ln)
		} else {
			// 标准 TLS 模式
			log.Printf("🛡️  启用标准 TLS 模式监听端口 %s", port)
			errChan <- server.ListenAndServeTLS(stdCert, stdKey)
		}
	}()
	select { //主协程等待两个信号：服务器运行错误 或者 接收到退出信号
	case err := <-errChan: //信号A，启动时发生错误，立即返回错误
		return err
	case <-ctx.Done(): //信号B，接收到main.go 传入的 Ctrl+C 信号
		log.Println("⛔ [Server] 收到退出信号，已停止接收新请求,正在安全关闭服务器...")
		// 设定打烊倒计时：给还在处理的请求最多 15 秒的时间完成，超过这个时间就强制关闭连接，避免资源占用过久
		// 到时间自动触发 cancel()，从而通知所有使用这个 shutdownCtx 的协程安全退出
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel() //确保在函数退出时取消上下文，释放资源

		// 执行优雅停机，Shutdown 会阻塞，直到所有活跃连接完成或者 shutdownCtx 超时
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("❌ [Server] 服务器优雅停机超时或出错，强制切断残余连接: %v", err)
			return err
		}
		log.Println("✅ [Server] 服务器已安全关闭，所有连接已断开，退出程序")
		return nil
	}
}
