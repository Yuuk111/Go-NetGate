package router

import (
	"net/http"
	"sort"
	"strings"
)

type RouteEntry struct {
	Prefix  string
	Handler http.Handler
}

type Router struct {
	routes []*RouteEntry
}

func NewRouter() *Router {
	return &Router{
		routes: make([]*RouteEntry, 0),
	}
}

// AddRoute 添加路由规则，prefix 是 URL 前缀，handler 是对应的处理器
func (r *Router) AddRoute(prefix string, handler http.Handler) {
	r.routes = append(r.routes, &RouteEntry{
		Prefix:  prefix,
		Handler: handler,
	})

	// 每次添加完路由后，根据前缀长度进行排序，确保最长前缀优先匹配
	// 保证精确匹配
	sort.Slice(r.routes, func(i, j int) bool {
		return len(r.routes[i].Prefix) > len(r.routes[j].Prefix)
	})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	for _, route := range r.routes {
		if strings.HasPrefix(path, route.Prefix) {
			route.Handler.ServeHTTP(w, req)
			// 匹配到路由后直接返回，不再继续匹配
			return
		}
	}
	// 如果没有任何路由匹配，返回 404 Not Found
	http.NotFound(w, req)
}
