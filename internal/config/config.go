package config

import (
	"flag"
)

type Appconfig struct {
	ListenPort       string
	TargetURL        string
	BackendServers   []string
	LoadBalancerAlgo string
}

func LoadConfig() *Appconfig {
	cfg := new(Appconfig)
	flag.StringVar(&cfg.ListenPort, "port", "8443", "网关监听端口")
	flag.StringVar(&cfg.TargetURL, "t", "http://localhost:9001", "后端目标服务器地址")
	flag.StringVar(&cfg.LoadBalancerAlgo, "algo", "RR", "负载均衡算法\n可选值: RR (轮询), IPHash (基于IP哈希)")

	flag.Parse()
	//后续改为从配置文件加载
	cfg.BackendServers = []string{"http://localhost:9001", "http://localhost:9002", "http://localhost:9003"}
	return cfg
}
