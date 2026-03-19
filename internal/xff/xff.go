package xff

import (
	"net"
	"net/http"
	"strings"
)

// GetClientIP 从 X-Forwarded-For 头部获取客户端 IP 地址
func GetClientIP(r *http.Request) (string, error) {
	xFF := r.Header.Get("X-Forwarded-For")
	if xFF != "" {
		ips := strings.Split(xFF, ",")
		// X-Forwarded-For 头部可能包含多个 IP 地址，按照标准格式是 "client, proxy1, proxy2"，我们取第一个 IP 地址作为客户端 IP 地址
		return strings.TrimSpace(ips[0]), nil
	}
	//如果没有 X-Forwarded-For 头部，取X-Real-IP
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return strings.TrimSpace(xRealIP), nil
	}
	//如果都没有，直接取 RemoteAddr 的 IP 部分
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err //如果无法解析 RemoteAddr，返回 nil
	}
	return ip, nil
}
