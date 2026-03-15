package proxy

import (
	"net/url"
	"sync/atomic"
)

type RoundRobinLB struct {
	backends []*url.URL
	current  uint64
}

// Next 返回下一个后端服务器的 URL，使用轮询算法实现负载均衡
// 是RoundRobinLB的一个方法
func (lb *RoundRobinLB) Next() *url.URL {
	// 通过原子操作获取下一个后端服务器的索引，确保线程安全
	idx := atomic.AddUint64(&lb.current, 1) % uint64(len(lb.backends))
	return lb.backends[idx]
}
