package waf

import (
	"net/http"
	"time"

	"github.com/Yuuk111/Go-NetGate/internal/insight"
	"github.com/Yuuk111/Go-NetGate/internal/xff"
	"github.com/Yuuk111/Go-NetGate/pb" // 替换为你的实际 pb 路径
	"github.com/google/uuid"           //
)

func InsightMiddleware(sender insight.LogSender) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 生成一个唯一的请求 ID，便于在日志中追踪
			traceID := uuid.New().String()

			// 从 XFF 中提取客户端 IP 地址
			clientIP, _ := xff.GetClientIP(r)

			// 构造日志项
			logItem := &pb.LogItem{
				TraceId:       traceID,
				SourceIp:      clientIP,
				DestinationIp: r.Host,
				Method:        r.Method,
				Path:          r.URL.Path,
				Query:         r.URL.RawQuery,
				Payload:       "", // 暂时不记录io.TeepReader(r.Body)的内容，后续可以根据需要添加
				Timestamp:     int32(time.Now().Unix()),
			}

			if sender != nil {
				sender.SendLog(logItem) //异步发送日志，不等待结果，避免阻塞请求处理
			}

			next.ServeHTTP(w, r)

		})
	}
}
