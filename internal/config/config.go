package config

import (
	"flag"
	"fmt"

	"github.com/spf13/viper"
	"golang.org/x/time/rate"
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
	Server          ServerConfig          `mapstructure:"server"`
	LoadBalance     LoadBalanceConfig     `mapstructure:"load_balance"`
	Gm              GmConfig              `mapstructure:"gmtls"`
	Tls             TlsConfig             `mapstructure:"tls"`
	Redis           RedisConfig           `mapstructure:"redis"`
	RedisRateLimit  RedisRateLimitConfig  `mapstructure:"redis_rate_limit"`
	SingleRateLimit SingleRateLimitConfig `mapstructure:"single_rate_limit"`
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

// RedisConfig 定义 Redis 连接相关的配置结构体
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// RedisRateLimitConfig 定义分布式限流相关的配置结构体
type RedisRateLimitConfig struct {
	Rate  rate.Limit `mapstructure:"rate"`  // 每秒生成的令牌数
	Burst int        `mapstructure:"burst"` // 令牌桶容量
}

// SingleRateLimitConfig 定义单机内存限流相关的配置结构体
type SingleRateLimitConfig struct {
	Rate  rate.Limit `mapstructure:"rate"`  // 每秒生成的令牌数
	Burst int        `mapstructure:"burst"` // 令牌桶容量
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

	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 100)

	viper.SetDefault("redis_rate_limit.rate", 5)
	viper.SetDefault("redis_rate_limit.burst", 10)

	viper.SetDefault("single_rate_limit.rate", 5)
	viper.SetDefault("single_rate_limit.burst", 10)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	// 解析配置到结构体
	var cfg AppFileConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置到结构体失败: %w", err)
	}

	// 检查值合法性
	if cfg.Redis.DB < 0 {
		return nil, fmt.Errorf("Redis DB 索引不能为负数")
	}

	if cfg.RedisRateLimit.Rate <= 0 || cfg.RedisRateLimit.Burst <= 0 {
		return nil, fmt.Errorf("分布式限流参数必须为正整数")
	}

	if cfg.SingleRateLimit.Rate <= 0 || cfg.SingleRateLimit.Burst <= 0 {
		return nil, fmt.Errorf("单机限流参数必须为正整数")
	}

	return &cfg, nil

}
