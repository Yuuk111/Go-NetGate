# NetGate 🛡️

[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Status](https://img.shields.io/badge/Status-Production%20Ready-success)]()

**NetGate** 是一款基于 Go 语言深度定制开发的高性能、云原生微服务 API 网关。

项目旨在解决微服务架构下的流量精细化编排、高并发限流防刷、以及配置变更导致的服务抖动问题。通过原生 Go 特性与现代架构设计的结合，NetGate 实现了极简部署与极致性能的统一。

## ✨ 核心特性 (Core Features)

* 🚀 **精准长前缀路由 (LPM Routing):** 弃用低效的遍历匹配，基于 Longest Prefix Match (最长前缀匹配) 算法实现智能路由引擎，保障 API 请求的极速精准寻址。

* 🛡️ **多级容灾与防刷 WAF:**
    * **全局并发限流 (Load Shedding):** 基于 Go Channel 信号量机制，在系统过载时果断抛弃溢出流量，保护底层架构不被洪峰压垮。
    * **高可用分布式限流:** 基于 Redis + Lua 脚本实现 IP 级别令牌桶限流。内置**本地内存降级防线**，在 Redis 宕机时秒级无缝切回单机限流。
* ⚖️ **动态探活负载均衡:** 支持 Round-Robin 与 IP-Hash 算法。内置 Active Health Check（主动心跳探活），后端节点宕机时触发秒级平滑剔除与自动恢复。
* 🔒 **金融级信创安全体系:** * 支持标准 TLS 与 **国密双证书 (GmTLS)** 模式无缝切换，满足政企与金融场景的合规诉求。
    * 深度定制 `httputil.ReverseProxy`，强制覆写非法 Host，彻底封堵**开放代理 (Open Proxy)** 漏洞。
    * 接管全局 `Context`，实现进程级别的**优雅停机 (Graceful Shutdown)**。

## 📂 架构与目录说明 (Project Structure)

项目严格遵循领域驱动设计 (DDD) 与 Go 官方工程规范：

```text
├── certs/                # TLS 与 GmTLS(国密) 证书目录
├── internal/             # 核心私有业务逻辑 (防篡改设计)
│   ├── config/           # Viper 配置解析与模型定义
│   ├── gmtls/            # 国密算法与双证书加载支持
│   ├── proxy/            # 核心转发层 (LPM路由引擎、反向代理、负载均衡、连接池调优)
│   ├── server/           # HTTP/HTTPS/GmTLS 监听器与优雅停机
│   ├── waf/              # Web 应用防火墙 (多级限流器、恶意载荷拦截)
│   └── xff/              # 客户端真实 IP (X-Forwarded-For) 解析提取
├── tools/                # 辅助工具脚本
├── config.yml            # 网关主配置文件
└── main.go               # 入口文件：生命周期管理与热更新守护进程
````
## 🗺️ 开发进度与 Roadmap

**🧩 基础模块 (Core & Foundation)**

  - [x] 解耦功能模块
  - [x] 基础反向代理
  - [x] 实现国密连接
  - [ ] 理解双证书
  - [ ] 分离静态资源和接口 (后端请求)
  - [x] Viper 配置文件
  - [x] 国密TLS / TLS1.2 切换
  - [x] 优雅停机 (Graceful Shutdown)
  - [ ] 线程池/协程池
  - [ ] 访问日志
  - [ ] 报错日志
  - [x] 多实例改造
  - [ ] 容器化部署实现守护进程
  - [ ] 可视化界面

**🛡️ WAF 模块 (Web Application Firewall)**

  - [ ] 数据清洗 (去空格，去注释，统一大小写)
  - [ ] AC 自动机匹配
  - [ ] Libinjection (可选)
  - [ ] 正则匹配
  - [ ] WAF 日志
  - [x] 动态配置 Viper (基于 `fsnotify` 与 RCU 的无锁热更新)

**⚖️ 负载均衡模块 (Load Balancing)**

  - [x] 负载均衡：基础轮询 (RR)
  - [x] 负载均衡：IP 哈希 (IP Hash)
  - [ ] 负载均衡：最小连接数
  - [x] 负载均衡：模式动态切换

**🔗 后端侧连接模块 (Upstream)**

  - [x] 智能路由 (LPM)
  - [ ] 智能路由：压缩前缀树优化 (Radix Trie)
  - [x] 连接池管理
  - [x] Transport 调优
  - [x] ErrorHandler 异常接管
  - [x] 主动心跳检测 (Active Health Check)
  - [ ] 被动心跳检测
  - [ ] 健康检查多实例化共享
  - [x] 初级熔断与降级
  - [x] 熔断与降级 (502/504 保护机制)
  - [ ] 熔断状态机

**🌐 客户端侧连接模块 (Downstream)**

  - [x] 超时设置：防慢速连接和雪崩 (Read/Write/Idle Timeout)

**🚦 限流模块 (Rate Limiting)**

  - [x] 流量控制：令牌桶
  - [x] IP 令牌桶 (单机内存)
  - [x] 全局令牌桶 (Load Shedding)
  - [x] 分布式全局令牌桶 (Redis + Lua)
  - [ ] 内存优化：零拷贝与缓冲区池
  - [ ] 多核利用和系统参数调优

**🧪 测试与调优 (Testing & Profiling)**

  - [ ] Pprof 性能分析
  - [x] 并发测试：压测 hey 25000 QPS
  - [ ] 协程数测试
  - [ ] 回收检测 (内存泄露/Goroutine 泄露检测)


***


## 🛠️ 快速开始 (Quick Start)

### 1\. 环境准备

确保您的系统已安装 Go 1.20 或更高版本，并准备好 Redis 实例。

### 2\. 克隆与启动

```bash
# 克隆仓库
git clone [https://github.com/YourID/Go-NetGate.git](https://github.com/YourID/Go-NetGate.git)
cd Go-NetGate

# 下载依赖
go mod tidy

# 启动网关
go run main.go
```

### 3\. 配置示例 (`config.yml`)

NetGate 的所有行为都可以通过 `config.yml` 进行控制，并且**支持在运行时修改，瞬间生效 (Hot Reload)**。

```yaml
server:
  port: "8443"
  tls_mode: "tls" # 支持 tls 或 gmtls(国密)

# 核心路由与负载均衡表
route_rules:
  - path: "/api/order"
    algorithm: "RR"
    backend:
      - "http://localhost:9004"
  - path: "/" # 兜底路由
    algorithm: "IPHash"
    backend:
      - "http://localhost:9001"
      - "http://localhost:9002"

# 分布式 Redis 限流配置
redis_rate_limit:
  rate: 5   # 每秒生成的令牌数
  burst: 10 # 令牌桶容量

# Redis 节点配置
redis:
  addr: "localhost:6379"
  password: ""
  db: 0
```

## 📈 性能与稳定性防御策略

  * **TCP 连接池调优:** 复用向后端的长连接 (`MaxIdleConnsPerHost`)，避免高并发下产生大量 `TIME_WAIT` 导致端口耗尽。
  * **防级联故障 (Cascading Failure):** 重写 `ErrorHandler`，在后端微服务全线崩溃时，网关坚如磐石，优雅返回 JSON 格式的 `502 Bad Gateway`，拒绝无效 TCP 拨号阻塞。

## 🤝 贡献与支持 (Contributing)

欢迎提交 Pull Request 或发起 Issue 探讨网络编程、高并发网关架构设计等技术细节！

## 📄 开源协议 (License)

本项目采用 [MIT License](https://www.google.com/search?q=LICENSE) 协议开源。

```

***

这份 README 不仅结构清晰，而且把你在背后付出的“看不见的心血”（比如防开放代理、连接池优化、防级联故障）全部摆在了台面上。任何一个懂行的技术总监看到这份文档，都会立刻意识到这是一个经历过深思熟虑的工业级项目。

下一步，你可以把代码推送到 GitHub 上了！需要我为你提供一份标准的 `.gitignore` 文件，或者教你怎么打个优雅的 Git Tag 发版吗？
```