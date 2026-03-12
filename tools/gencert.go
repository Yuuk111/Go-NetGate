package main

import (
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/x509"
)

func main() {
	// 1. 创建存放证书的目录
	err := os.MkdirAll("certs", 0755)
	if err != nil {
		log.Fatalf("创建目录失败: %v", err)
	}

	fmt.Println("开始生成国密 SM2 双证书...")

	// 2. 生成一个模拟的 Root CA (根证书)
	caKey, _ := sm2.GenerateKey(rand.Reader)
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10年有效期
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	// 【修复点 1】去掉了 rand.Reader
	caBytes, _ := x509.CreateCertificate(caTemplate, caTemplate, &caKey.PublicKey, caKey)
	savePEM("certs/CA.crt", "CERTIFICATE", caBytes)

	// 3. 生成 SS (签名证书 Sign Certificate)
	ssKey, _ := sm2.GenerateKey(rand.Reader)
	ssTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	// 【修复点 2】去掉了 rand.Reader
	ssBytes, _ := x509.CreateCertificate(ssTemplate, caTemplate, &ssKey.PublicKey, caKey)
	savePEM("certs/SS.crt", "CERTIFICATE", ssBytes)
	savePrivateKey("certs/SS.key", ssKey)
	fmt.Println("✅ 签名证书 (SS.crt, SS.key) 生成完毕")

	// 4. 生成 SE (加密证书 Encrypt Certificate)
	seKey, _ := sm2.GenerateKey(rand.Reader)
	seTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	// 【修复点 3】去掉了 rand.Reader
	seBytes, _ := x509.CreateCertificate(seTemplate, caTemplate, &seKey.PublicKey, caKey)
	savePEM("certs/SE.crt", "CERTIFICATE", seBytes)
	savePrivateKey("certs/SE.key", seKey)
	fmt.Println("✅ 加密证书 (SE.crt, SE.key) 生成完毕")
}

// 辅助函数：保存 PEM 格式文件
func savePEM(filename, pemType string, bytes []byte) {
	file, _ := os.Create(filename)
	defer file.Close()
	pem.Encode(file, &pem.Block{Type: pemType, Bytes: bytes})
}

// 辅助函数：保存 SM2 私钥
func savePrivateKey(filename string, key *sm2.PrivateKey) {
	file, _ := os.Create(filename)
	defer file.Close()
	bytes, _ := x509.WritePrivateKeyToPem(key, nil)
	file.Write(bytes)
}