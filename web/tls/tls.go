package tls

import (
	"crypto/tls"

	"github.com/xudefa/enhance/web/server"
)

// LoadTLSConfig 从证书文件和密钥文件加载 TLS 配置。
// 委托给 server.LoadTLSConfig 实现。
func LoadTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	return server.LoadTLSConfig(certFile, keyFile)
}

// LoadTLSConfigWithCA 加载 TLS 配置并启用 CA 证书验证（用于客户端验证服务端）。
// 委托给 server.LoadTLSConfigWithCA 实现。
func LoadTLSConfigWithCA(certFile, keyFile, caFile string) (*tls.Config, error) {
	return server.LoadTLSConfigWithCA(certFile, keyFile, caFile)
}

// LoadClientTLSConfig 加载仅客户端验证用的 TLS 配置（不需要服务端证书）。
// 委托给 server.LoadClientTLSConfig 实现。
func LoadClientTLSConfig(caFile string) (*tls.Config, error) {
	return server.LoadClientTLSConfig(caFile)
}

// LoadCertFromPEM 从 PEM 编码的字节数据解析 TLS 证书对。
// 委托给 server.LoadCertFromPEM 实现。
func LoadCertFromPEM(certPEM, keyPEM []byte) (*tls.Config, error) {
	return server.LoadCertFromPEM(certPEM, keyPEM)
}

// InsecureTLSConfig 创建跳过证书验证的 TLS 配置（仅用于开发测试）。
// 委托给 server.InsecureTLSConfig 实现。
func InsecureTLSConfig() *tls.Config {
	return server.InsecureTLSConfig()
}

// MustLoadTLSConfig 加载 TLS 配置，失败时 panic。
// 委托给 server.MustLoadTLSConfig 实现。
func MustLoadTLSConfig(certFile, keyFile string) *tls.Config {
	return server.MustLoadTLSConfig(certFile, keyFile)
}
