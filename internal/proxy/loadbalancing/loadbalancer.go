package loadbalancing

import (
	"hash/fnv"
	"log"
	"net/url"
	"sync/atomic"
)

// LoadBalancer 定义了负载均衡器接口，提供获取下一个后端服务器 URL 的方法
type LoadBalancer interface {
	Next(clientIP string) *url.URL
}

// RoundRobinLB 基础轮询负载均衡器

type RoundRobinLB struct {
	backends []*url.URL
	current  uint64
}

// Next 返回下一个后端服务器的 URL，使用轮询算法实现负载均衡
// 是RoundRobinLB的一个方法
func (lb *RoundRobinLB) Next(clientIP string) *url.URL {
	// 通过原子操作获取下一个后端服务器的索引，确保线程安全
	idx := atomic.AddUint64(&lb.current, 1) % uint64(len(lb.backends))
	return lb.backends[idx]
}

// IpHashLB 基于客户端 IP 地址哈希的负载均衡器
type IpHashLB struct {
	backends []*url.URL
}

func (lb *IpHashLB) Next(clientIP string) *url.URL {
	if len(lb.backends) == 0 {
		return nil
	}

	// 计算 IP 哈希值
	hashVal := hashIP(clientIP)
	// 通过哈希值对服务器数量取模获得索引
	idx := hashVal % uint32(len(lb.backends))
	return lb.backends[idx]
}

// hashIP 将客户端 IP 地址转换为一个哈希值，用于在 IP 哈希负载均衡算法中选择后端服务器
func hashIP(ip string) uint32 {
	// 使用 FNV-1a 算法将 IP 字符串转换为一个 32 位的哈希值
	h := fnv.New32a()
	h.Write([]byte(ip)) //把 IP 地址转换为字节切片并写入哈希计算器
	return h.Sum32()    //返回计算得到的哈希值
}

func NewLoadBalancer(algo string, parsedURLs []*url.URL) LoadBalancer {
	switch algo {
	case "RR": //基础轮询
		return &RoundRobinLB{
			backends: parsedURLs,
			current:  0,
		}
	case "IPHash": //基于 IP 哈希
		return &IpHashLB{
			backends: parsedURLs,
		}
	default: //默认使用基础轮询
		log.Printf("未知的负载均衡算法 '%s'，默认使用基础轮询算法", algo)
		return &RoundRobinLB{
			backends: parsedURLs,
			current:  0,
		}
	}
}
