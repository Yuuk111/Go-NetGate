package config

import (
	"flag"
	"fmt"

	"github.com/spf13/viper"
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

// AppFileConfig 定义了从配置文件加载的结构体
type AppFileConfig struct {
	Server      ServerConfig      `mapstructure:"server"`
	LoadBalance LoadBalanceConfig `mapstructure:"load_balance"`
	Gm          GmConfig          `mapstructure:"gmtls"`
	Tls         TlsConfig         `mapstructure:"tls"`
}

// ServerConfig 定义服务器相关的配置结构体
type ServerConfig struct {
	ListenPort string `mapstructure:"port"`
	TLSMode    string `mapstructure:"tls_mode"` // tls模式
}

// LoadBalanceConfig 定义负载均衡相关的配置结构体
type LoadBalanceConfig struct {
	Algorithm string   `mapstructure:"algorithm"`
	Backends  []string `mapstructure:"backends"`
}

// GmConfig 定义国密证书相关的配置结构体
type GmConfig struct {
	SignCertFile string `mapstructure:"SignCertFile"`
	SignKeyFile  string `mapstructure:"SignKeyFile"`
	EncCertFile  string `mapstructure:"EncCertFile"`
	EncKeyFile   string `mapstructure:"EncKeyFile"`
}

// TlsConfig 定义标准TLS证书相关的配置结构体
type TlsConfig struct {
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

func LoadFileConfig() (*AppFileConfig, error) {
	// 这里可以实现从文件加载配置的逻辑，例如使用 Viper 库
	viper.SetConfigName("config") // 配置文件名（不带扩展名）
	viper.SetConfigType("yaml")   // 配置文件类型
	viper.AddConfigPath(".")      // 配置文件路径
	// 设置默认值
	viper.SetDefault("server.port", "8443")
	viper.SetDefault("load_balance.algorithm", "RR")
	viper.SetDefault("server.tls_mode", "gmtls")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	// 解析配置到结构体
	var cfg AppFileConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置到结构体失败: %w", err)
	}
	return &cfg, nil

}
