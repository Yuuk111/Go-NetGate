package limit

import (
	"log"
	"net/http"
)

// GlobalConcurrencyLimiter 全局并发限流器 (Load Shedding)
type GlobalConcurrencyLimiter struct {
	// 这是一个带缓冲的通道，充当“令牌槽”或“排队位”
	semaphore chan struct{}
}

// NewGlobalConcurrencyLimiter 创建全局限流器，maxConcurrent 是网关能承受的最大同时处理数
func NewGlobalConcurrencyLimiter(maxConcurrent int) *GlobalConcurrencyLimiter {
	return &GlobalConcurrencyLimiter{
		semaphore: make(chan struct{}, maxConcurrent),
	}
}

// GlobalLimitMiddleware 全局并发限流中间件
func (l *GlobalConcurrencyLimiter) GlobalLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// 尝试向 channel 中塞入一个空结构体
		select {
		case l.semaphore <- struct{}{}:
			// 成功塞入！说明当前并发量还没有达到最大值

			// 当请求处理完毕，必须把位置空出来！
			defer func() {
				<-l.semaphore
			}()

			// 放行给下一层
			next.ServeHTTP(w, r)

		default:
			// channel 满了，当前并发已达物理极限
			// 触发负载抛弃，快速失败，保护网关和后端不被压垮
			log.Print("⛔ [系统过载] 触发全局并发限流！当前并发已满，拒绝新连接")

			// HTTP 503 Service Unavailable
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error": "503 Service Unavailable. System is overloaded, please try again later."}`))
		}
	})
}
