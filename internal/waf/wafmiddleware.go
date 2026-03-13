package waf

import (
	"net/http"
)

func WafMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if checker(r) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("WAF Blocked: Malicious Request Detected"))
			//检测到恶意请求，直接返回，不调用 next
			return
		}

		//通过 WAF 检测，转交给下一个处理器（即 Proxy）
		next.ServeHTTP(w, r)
	})
}
