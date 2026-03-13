package waf

import (
	"net/http"
)

func checker(req *http.Request) bool {
	query := req.URL.Query().Encode()
	_ = query
	// 这里可以添加更多的检测逻辑，例如检查请求头、请求体等
	return false
}
