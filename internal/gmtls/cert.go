package gmtls

import (
	"github.com/tjfoc/gmsm/gmtls"
)

func LoadGMTLSConfig(signCertPath, signKeyPath, encCertPath, encKeyPath string) (*gmtls.Config, error) {
	// 加载签名证书
	sigCert, err := gmtls.LoadX509KeyPair(signCertPath, signKeyPath)
	if err != nil {
		return nil, err
	}

	// 加载加密证书
	encCert, err := gmtls.LoadX509KeyPair(encCertPath, encKeyPath)
	if err != nil {
		return nil, err
	}

	// 配置 gmtls
	// 注意：Certificates 数组中，第一个必须是签名证书，第二个是加密证书
	config := &gmtls.Config{
		Certificates: []gmtls.Certificate{sigCert, encCert},
		GMSupport:    &gmtls.GMSupport{},
	}

	return config, nil
}
