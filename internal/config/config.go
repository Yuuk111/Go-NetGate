package config

import (
	"flag"
)

type Appconfig struct {
	ListenPort     string
	TargetURL      string
	BackendServers []string
}

func LoadConfig() *Appconfig {
	cfg := new(Appconfig)
	flag.StringVar(&cfg.ListenPort, "port", "8443", "网关监听端口")
	flag.StringVar(&cfg.TargetURL, "t", "http://localhost:9001", "后端目标服务器地址")
	flag.Parse()
	cfg.BackendServers = []string{"http://localhost:9001", "http://localhost:9002", "http://localhost:9003"}
	return cfg
}
