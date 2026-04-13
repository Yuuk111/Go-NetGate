# NetGate 🛡️

[](https://go.dev/)
[](https://www.google.com/search?q=https://python.org/)
[](https://opensource.org/licenses/MIT)
[](https://www.google.com/search?q=)

**NetGate** 是一款基于 Go 语言深度定制开发的高性能、云原生微服务 API 网关，并创新性地搭载了基于 Python 与大模型的 **AI 旁路安全审计大脑**。

项目旨在解决微服务架构下的流量精细化编排、高并发限流防刷、以及零 0-day 威胁感知问题。通过原生 Go 特性、云原生容器化编排与现代 AI Agent 架构的结合，NetGate 实现了极简部署、极致性能与智能防御的完美统一。

## ✨ 核心特性 (Core Features)

  * 🚀 **精准长前缀路由 (LPM Routing):** 弃用低效的遍历匹配，基于 Longest Prefix Match (最长前缀匹配) 算法实现智能路由引擎，保障 API 请求的极速精准寻址。
  * 🛡️ **多级容灾与防刷 WAF:**
      * **全局并发限流 (Load Shedding):** 基于 Go Channel 信号量机制，在系统过载时果断抛弃溢出流量，保护底层架构不被洪峰压垮。
      * **高可用分布式限流:** 基于 Redis + Lua 脚本实现 IP 级别令牌桶限流。内置**本地内存降级防线**，在 Redis 宕机时秒级无缝切回单机限流。
  * 🧠 **全异步 AI 旁路审计引擎 (Insight Brain):**
      * **微秒级流量镜像:** 基于 gRPC Client Streaming 构建单向非阻塞数据流，将网关日志异步旁路至检测端，主链路**零性能损耗**。
      * **ReAct 智能态势感知:** 搭载异步 Python Agent，赋予大模型多轮思考与逻辑推理能力，实现对 0-day 漏洞探测、高级 APT 攻击的自动化定性与溯源。
  * ⚖️ **动态探活负载均衡:** 支持 Round-Robin 与 IP-Hash 算法。内置 Active Health Check（主动心跳探活），后端节点宕机时触发秒级平滑剔除与自动恢复。
  * 🔒 **金融级信创安全体系:** \* 支持标准 TLS 与 **国密双证书 (GmTLS)** 模式无缝切换，满足政企与金融场景的合规诉求。
      * 深度定制 `httputil.ReverseProxy`，强制覆写非法 Host，彻底封堵**开放代理 (Open Proxy)** 漏洞。
      * 接管全局 `Context`，实现进程级别的**优雅停机 (Graceful Shutdown)**。

## 📂 架构与目录说明 (Project Structure)

项目严格遵循领域驱动设计 (DDD) 与 12-Factor App 云原生规范：

```text
├── certs/                # TLS 与 GmTLS(国密) 证书目录 (挂载)
├── internal/             # 核心私有业务逻辑 (防篡改设计)
│   ├── config/           # Viper 配置解析与模型定义 (支持环境变量注入)
│   ├── gmtls/            # 国密算法与双证书加载支持
│   ├── insight/          # gRPC 异步日志上报引擎
│   ├── proxy/            # 核心转发层 (LPM路由引擎、反向代理、负载均衡)
│   ├── server/           # HTTP/HTTPS/GmTLS 监听器与优雅停机
│   └── waf/              # Web 应用防火墙 (多级限流器、恶意载荷拦截)
├── app/                  # AI 旁路审计中枢 (Python)
│   ├── agent/            # 基于 AsyncOpenAI 的 ReAct 状态机
│   ├── grpc_server/      # aio.grpc 流量接收端点
│   └── web/              # 基于 aiohttp + SSE 的大屏可视化面板
├── docker-compose.yml    # 微服务一键编排配置
├── Dockerfile            # Go 网关多阶段构建脚本
├── config.yml            # 网关主拓扑配置文件 (挂载)
├── .env.example          # 机密环境变量模板
└── main.go               # 入口文件：生命周期管理与热更新守护进程
```

## 🗺️ 开发进度与 Roadmap

**🧩 基础与路由模块 (Core & Foundation)**

  - [x] 基础反向代理与长前缀智能路由 (LPM)
  - [x] 国密连接 (GmTLS) 与 TLS1.2/1.3 切换
  - [x] Viper 配置解析与环境变量覆盖 (`.env`)
  - [x] 优雅停机 (Graceful Shutdown) 与信号劫持
  - [ ] 智能路由：压缩前缀树优化 (Radix Trie)

**🛡️ WAF 与限流模块 (Firewall & Rate Limiting)**

  - [x] 全局令牌桶并发限流 (Load Shedding)
  - [x] 分布式全局令牌桶 (Redis + Lua)
  - [x] IP 令牌桶 (单机内存降级)
  - [x] 动态配置 Viper (基于 RCU 的无锁热更新)
  - [ ] AC 自动机恶意载荷匹配

**🧠 AI 安全大脑 (Insight Agent)**

  - [x] Go 端 gRPC 异步 Client Streaming 上报
  - [x] Python 端 Asyncio 生产者-消费者解耦队列
  - [x] 基于大模型的 ReAct 架构日志推理分析
  - [x] SSE 协议的实时 LLM 思考链路大屏 (Dashboard)
  - [ ] 接入 LangGraph 构建图节点级确定性工作流
  - [ ] eBPF XDP 内核级联动精准封禁恶意 IP

**⚖️ 负载均衡与上游连接 (Load Balancing & Upstream)**

  - [x] 负载均衡：基础轮询 (RR) 与 IP 哈希
  - [x] 主动心跳检测 (Active Health Check)
  - [x] ErrorHandler 异常接管与防级联故障 (502/504)
  - [x] 连接池调优防 TIME\_WAIT 耗尽

**🌐 客户端侧连接模块 (Downstream)**

  - [x] 严苛超时设置：防慢速攻击 (Read/Write/Idle Timeout)

**🐳 云原生部署体系 (Cloud Native Deployment)**

  - [x] Go 核心网关多阶段极简构建 (Multi-stage Build)
  - [x] Python Agent 隔离编译环境构建
  - [x] Docker Compose 微服务全栈联动编排

## 🛠️ 快速开始 (Quick Start)

### 1\. 环境准备

确保您的系统已安装 `Docker` 与 `Docker Compose`。

### 2\. 容器化一键启动 (推荐)

NetGate 采用了配置分离的最佳实践，机密数据通过环境变量注入。

```bash
# 1. 克隆仓库
git clone https://github.com/YourID/Go-NetGate.git
cd Go-NetGate

# 2. 准备环境变量 (请在 .env 中填入你的大模型 API_KEY 等机密信息)
cp .env.example .env

# 3. 一键编译与启动网关、Redis、AI 审计大脑
docker-compose up -d --build

# 4. 查看 AI 审计大屏
# 打开浏览器访问 http://localhost:50052
```

### 3\. 本地原生调试

```bash
# 下载依赖
go mod tidy
# 启动 Go 网关
go run main.go

# (另起终端) 启动 Python Agent
cd app && pip install -r requirements.txt
python main.py
```

## 📈 性能与稳定性防御策略

  * **全链路异步隔离:** 网关上报日志与 AI 大模型响应速度彻底解耦，即便模型 API 响应长达 10 秒，网关的 10w+ QPS 转发性能亦不受任何波及。
  * **防级联故障 (Cascading Failure):** 重写 `ErrorHandler`，在后端微服务全线崩溃时，网关坚如磐石，优雅返回 JSON 格式的 `502 Bad Gateway`，拒绝无效 TCP 拨号阻塞。

## 🤝 贡献与支持 (Contributing)

欢迎提交 Pull Request 或发起 Issue 探讨网络编程、大模型安全 Agent 落地、高并发网关架构设计等技术细节！

## 📄 开源协议 (License)

本项目采用 [MIT License](https://opensource.org/licenses/MIT) 协议开源。