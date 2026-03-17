package gmtls

import (
	tjgmtls "github.com/tjfoc/gmsm/gmtls"
)

// loadGMTLSConfig 加载国密双证书并返回 TLS 配置
func LoadGMTLSConfig(signCertPath, signKeyPath, encCertPath, encKeyPath string) (*tjgmtls.Config, error) {
	// 加载签名证书
	sigCert, err := tjgmtls.LoadX509KeyPair(signCertPath, signKeyPath)
	if err != nil {
		return nil, err
	}

	// 加载加密证书
	encCert, err := tjgmtls.LoadX509KeyPair(encCertPath, encKeyPath)
	if err != nil {
		return nil, err
	}

	// 配置 gmtls
	// 注意：Certificates 数组中，第一个必须是签名证书，第二个是加密证书
	config := &tjgmtls.Config{
		Certificates: []tjgmtls.Certificate{sigCert, encCert},
		GMSupport:    &tjgmtls.GMSupport{},
	}

	return config, nil
}
