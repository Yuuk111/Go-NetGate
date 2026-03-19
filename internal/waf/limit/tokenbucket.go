package limit

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Yuuk111/Go-NetGate/internal/xff"
	"golang.org/x/time/rate"
)

// IPRateLimiter 维护所有访客的令牌桶
type IPRateLimiter struct {
	ips   map[string]*visitor // 存储每个访客的IP地址和对应的令牌桶
	mux   sync.RWMutex        // 读写锁，保证并发访问时的安全性
	ra    rate.Limit          // 令牌生成速率，例如每秒生成多少个令牌
	burst int                 // 令牌桶的容量，即允许的最大突发请求数
}

// visitor 代表一个访客的令牌桶和最后访问时间(用于清理闲置内存)
type visitor struct {
	limiter  *rate.Limiter // 令牌桶
	lastSeen time.Time
	mux      sync.Mutex // 互斥锁，保护 lastSeen 字段的并发访问
}

// NewIPRateLimiter 创建一个新的 IPRateLimiter 限流器实例
func NewIPRateLimiter(ra rate.Limit, burst int) *IPRateLimiter {
	limiter := &IPRateLimiter{
		ips:   make(map[string]*visitor),
		ra:    ra,
		burst: burst,
	}
	// 启动一个后台协程，定期清理闲置的访客记录，避免内存泄漏
	go limiter.cleanupVisitors()
	return limiter
}

func (i *IPRateLimiter) getVisitor(ip string) *rate.Limiter {
	i.mux.RLock()
	v, exists := i.ips[ip]
	i.mux.RUnlock()
	if exists { //已经存在，更新最后访问时间并返回令牌桶
		v.mux.Lock()
		v.lastSeen = time.Now()
		v.mux.Unlock()
		return v.limiter
	}
	//不存在，创建新的令牌桶并存储
	i.mux.Lock()
	defer i.mux.Unlock()
	//双重检查，防止在获取写锁期间其他协程已经创建了这个访客的记录
	v, exists = i.ips[ip]
	if !exists {
		limiter := rate.NewLimiter(i.ra, i.burst)
		i.ips[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}
	//如果已经存在了，说明在获取写锁期间其他协程已经创建了这个访客的记录，直接返回即可
	return v.limiter
}

// cleanupVisitors 定期清理闲置的访客记录，避免内存泄漏
func (i *IPRateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute) //每分钟执行一次清理

		i.mux.Lock()
		for ip, v := range i.ips {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(i.ips, ip) //删除闲置超过3分钟的访客记录
			}
		}
		i.mux.Unlock()
	}
}

func (i *IPRateLimiter) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, err := xff.GetClientIP(r) //获取访客IP地址
		if err != nil {
			log.Printf("解析 IP 失败: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		limiter := i.getVisitor(ip)
		if !limiter.Allow() { //若未获取到令牌，说明请求过于频繁，返回 429 Too Many Requests
			log.Printf("[WAF] IP %s 请求过于频繁，已被限流", ip)
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"Too Many Requests. CC Attack Blocked!", "message": "Your IP is being rate limited. Please try again later."}`))
			return
		}
		//获取到令牌，继续处理请求
		next.ServeHTTP(w, r)
	})
}
