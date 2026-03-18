package loadbalancing

import (
	"context"
	"hash/fnv"
	"log"
	"net"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

// LoadBalancer 定义了负载均衡器接口，提供获取下一个后端服务器 URL 的方法
type LoadBalancer interface {
	Next(clientIP string) *url.URL
}

// ==================
// 服务器状态封装
// ==================
type Backend struct {
	URL   *url.URL
	alive bool         // 存活状态
	mux   sync.RWMutex // 读写锁，保证并发安全；不能用互斥锁不然会退化成串行或者TOCTOU(检查与使用的时间差)问题
}

// SetAlive 设置服务器的存活状态，alive=true 表示服务器可用，alive=false 表示服务器不可用
func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.alive = alive
	b.mux.Unlock() //解锁
}

// IsAlive 返回服务器的存活状态，true 表示服务器可用，false 表示服务器不可用
func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	defer b.mux.RUnlock() //解锁
	return b.alive
}

// ==================
// 心跳检测协程
// ==================
func startHealthCheck(ctx context.Context, backends []*Backend) {
	// 定时器，每 10 秒执行一次健康检查
	ticker := time.NewTicker(10 * time.Second)
	// 不能马上defer 因为外层函数startHealthCheck不会阻塞，心跳检测是在一个独立的协程中运行的，所以需要在协程内停止定时器
	// 如果直接在外层函数 defer ticker.Stop()，那么这个定时器会在 startHealthCheck 函数返回时就被停止了，导致心跳检测协程无法正常工作
	go func() {
		defer ticker.Stop() // 确保在协程退出时停止定时器，释放资源
		for {
			select { // 等待定时器触发或者其他退出信号
			case <-ticker.C:
				log.Println("✅ [Health Check] 心跳检测协程正常运行，正在检查后端服务器状态...")
				for _, b := range backends {
					// 通过 TCP 连接检查服务器是否存活，超时时间为 2 秒
					conn, err := net.DialTimeout("tcp", b.URL.Host, 2*time.Second)
					if err != nil {
						if b.IsAlive() {
							log.Printf("🛑 [Health Check] [熔断] 后端节点宕机剔除:%s", b.URL.Host)
							b.SetAlive(false) //连接失败，标记服务器不可用
						}
					} else {
						conn.Close() //连接成功就关闭连接，释放资源
						if !b.IsAlive() {
							log.Printf("✅ [Health Check][恢复] 后端节点恢复上线:%s", b.URL.Host)
							b.SetAlive(true) //连接成功，标记服务器可用
						}
					}
				}
			case <-ctx.Done():
				log.Println("🛑 [Health Check] 收到退出信号，心跳检测协程安全销毁")
				return //退出协程 并且会触发 defer ticker.Stop() 来停止定时器，释放资源
			}
		}
	}()
}

// RoundRobinLB 基础轮询负载均衡器

type RoundRobinLB struct {
	backends []*Backend
	current  uint64
}

// Next 返回下一个后端服务器的 URL，使用轮询算法实现负载均衡
// 是RoundRobinLB的一个方法
func (lb *RoundRobinLB) Next(clientIP string) *url.URL {
	total := uint64(len(lb.backends))
	if total == 0 {
		return nil
	}
	// 循环查找下一个存活的服务器，最多尝试 total 次
	for i := uint64(0); i < total; i++ {
		// 通过原子操作获取下一个服务器的索引，确保线程安全
		idx := atomic.AddUint64(&lb.current, 1) % uint64(total)
		if lb.backends[idx].IsAlive() {
			return lb.backends[idx].URL
		}
	}

	log.Println("❌ [Load Balancer] [致命错误] 所有后端节点均已宕机！")
	return nil //所有服务器都不可用，返回 nil
}

// IpHashLB 基于客户端 IP 地址哈希的负载均衡器
type IpHashLB struct {
	backends []*Backend
}

// Next 返回下一个后端服务器的 URL，使用 IP 哈希算法实现负载均衡
func (lb *IpHashLB) Next(clientIP string) *url.URL {
	total := uint32(len(lb.backends))
	if total == 0 {
		return nil
	}
	// 计算 IP 哈希值
	hashVal := hashIP(clientIP)
	// IPhash 降级策略：线性探测，如果选中的服务器不可用，就尝试下一个服务器，最多尝试 total 次
	for i := uint32(0); i < total; i++ {
		idx := (hashVal + i) % total
		if lb.backends[idx].IsAlive() {
			return lb.backends[idx].URL
		}
	}

	log.Println("❌ [Load Balancer] [致命错误] 所有后端节点均已宕机！")
	return nil //所有服务器都不可用，返回 nil
}

// hashIP 将客户端 IP 地址转换为一个哈希值，用于在 IP 哈希负载均衡算法中选择后端服务器
func hashIP(ip string) uint32 {
	// 使用 FNV-1a 算法将 IP 字符串转换为一个 32 位的哈希值
	h := fnv.New32a()
	h.Write([]byte(ip)) //把 IP 地址转换为字节切片并写入哈希计算器
	return h.Sum32()    //返回计算得到的哈希值
}

func NewLoadBalancer(algo string, parsedURLs []*url.URL) LoadBalancer {
	//包装后端服务器列表，添加存活状态和锁
	backends := make([]*Backend, 0, len(parsedURLs))
	for _, u := range parsedURLs {
		backends = append(backends, &Backend{
			URL:   u,
			alive: true, //默认服务器都是存活的，后续通过心跳检测更新状态
		})
	}
	// 启动心跳检测协程，定期检查服务器的存活状态
	// 如果加入配置文件热更新，传入的 context 可以改为动态生成的，调用 cancel() 来安全销毁旧的心跳检测协程
	startHealthCheck(context.Background(), backends) //传入一个背景上下文，表示这个协程会一直运行直到程序退出

	switch algo {
	case "RR": //基础轮询
		return &RoundRobinLB{
			backends: backends,
			current:  0,
		}
	case "IPHash": //基于 IP 哈希
		return &IpHashLB{
			backends: backends,
		}
	default: //默认使用基础轮询
		log.Printf("⚠️ [Load Balancer] 未知的负载均衡算法 '%s'，默认使用基础轮询算法", algo)
		return &RoundRobinLB{
			backends: backends,
			current:  0,
		}
	}
}
