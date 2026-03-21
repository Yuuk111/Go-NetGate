package limit

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Yuuk111/Go-NetGate/internal/xff"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

/*
分布式令牌桶
*/
// Core: 分布式的 Lua 令牌桶脚本
// Redis 执行 Lua 是单线程的，保证了操作的原子性，适合实现分布式限流
const luaTokenBucketScript = `
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = 1

-- 获取当前令牌桶状态
local bucket = redis.call("HMGET", key, "tokens", "last_time")
local tokens_str = bucket[1]
local last_time_str = bucket[2]

local tokens, last_time

-- 如果桶不存在，初始化为满桶状态
if tokens_str == false then
	tokens = capacity
	last_time = now
else
	tokens = tonumber(tokens_str)
	last_time = tonumber(last_time_str)
end

-- 惰性计算令牌数量
local delta_time = math.max(0, now - last_time)
local generated_tokens = math.floor(delta_time * rate)
tokens = math.min(capacity, tokens + generated_tokens)

-- 判断是否有足够的令牌
if tokens >= requested then
	tokens = tokens - requested
	redis.call("HSET", key, "tokens", tokens, "last_time", now)
	redis.call("EXPIRE", key, math.ceil(capacity / rate * 2)) -- 设置过期时间，避免内存泄漏
	return 1 -- 放行
else
	redis.call("HSET", key, "tokens", tokens, "last_time", now)
	redis.call("EXPIRE", key, math.ceil(capacity / rate * 2)) -- 设置过期时间，避免内存泄漏
	return 0 -- 拒绝
end
`

const redisKeyPrefix = "netgate:ratelimit:"

// RedisRateLimiter 基于 Redis 的分布式令牌桶限流器
type RedisRateLimiter struct {
	redisClient   *redis.Client
	rate          rate.Limit
	burst         int
	script        *redis.Script
	localFallback *IPRateLimiter //本地内存限流器，作为 Redis 故障时的降级方案
}

// NewRedisRateLimiter 实例化全局 IP 限流器
func NewRedisRateLimiter(rdb *redis.Client, rate rate.Limit, burst int) *RedisRateLimiter {
	return &RedisRateLimiter{
		redisClient: rdb,
		rate:        rate,
		burst:       burst,
		//预加载 Lua 脚本，提升性能，后续调用只传 SHA1 摘要
		script:        redis.NewScript(luaTokenBucketScript),
		localFallback: NewIPRateLimiter(rate, burst), //初始化本地内存限流器
	}
}

func (redisl *RedisRateLimiter) RedisRateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, err := xff.GetClientIP(r)
		if err != nil {
			log.Printf("❌ [Redis Limiter] 解析访客 IP 失败: %v", err)
			next.ServeHTTP(w, r) //解析 IP 失败，放行请求，避免误伤
			return
		}
		//加前缀防冲突
		key := redisKeyPrefix + ip
		now := time.Now().Unix()

		//调用 Lua 脚本执行令牌桶算法，传入当前速率,桶容量,当前时间
		// 传入r.Context()，确保在请求取消时能够及时中断 Redis 操作，避免资源浪费
		// 还可以传入定时器的上下文，防止 Redis 响应过慢导致请求堆积，最后触发网关的全局超时保护机制等

		ctxTimeout, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond) //设置 Redis 操作的超时时间，防止长时间等待
		defer cancel()
		result, err := redisl.script.Run(ctxTimeout, redisl.redisClient, []string{key}, redisl.rate, redisl.burst, now).Int()
		cancel()
		// 大坑，result默认返回int64

		if err != nil { //降级策略：当 Redis 出现错误时，默认放行请求，避免误伤正常流量
			// 后续改为降级到本地内存限流，保证在 Redis 故障时仍然能够提供一定的保护能力
			log.Printf("❌ [Redis Limiter] Redis 限流器异常，自动放行: %v", err)
			next.ServeHTTP(w, r)
			return

		}

		if result == 0 {
			//拒绝请求，返回 429 Too Many Requests
			log.Printf("⛔ [Redis Limiter] 触发分布式限流，IP %s 请求过于频繁，已被拒绝", ip)
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"Too Many Requests. Distributed Rate Limit Blocked!", "message": "Your IP is being rate limited. Please try again later."}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}
