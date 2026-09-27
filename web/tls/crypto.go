package tls

import (
	"crypto/rsa"

	"github.com/xudefa/enhance/web/server"
)

// AESEncrypt 使用 AES-CBC 模式加密数据。
// 委托给 server.AESEncrypt 实现。
func AESEncrypt(plaintext, key, iv []byte) ([]byte, error) {
	return server.AESEncrypt(plaintext, key, iv)
}

// AESDecrypt 使用 AES-CBC 模式解密数据。
// 委托给 server.AESDecrypt 实现。
func AESDecrypt(ciphertext, key, iv []byte) ([]byte, error) {
	return server.AESDecrypt(ciphertext, key, iv)
}

// AESGCMEncrypt 使用 AES-GCM 模式加密数据（推荐，自带认证）。
// 委托给 server.AESGCMEncrypt 实现。
func AESGCMEncrypt(plaintext, key, additionalData []byte) ([]byte, error) {
	return server.AESGCMEncrypt(plaintext, key, additionalData)
}

// AESGCMDecrypt 使用 AES-GCM 模式解密数据。
// 委托给 server.AESGCMDecrypt 实现。
func AESGCMDecrypt(ciphertext, key, additionalData []byte) ([]byte, error) {
	return server.AESGCMDecrypt(ciphertext, key, additionalData)
}

// RSAGenerateKey 生成 RSA 密钥对。
// 委托给 server.RSAGenerateKey 实现。
func RSAGenerateKey(bits int) (*rsa.PrivateKey, error) {
	return server.RSAGenerateKey(bits)
}

// RSAEncrypt 使用 RSA 公钥加密数据（OAEP-SHA256）。
// 委托给 server.RSAEncrypt 实现。
func RSAEncrypt(publicKey *rsa.PublicKey, plaintext []byte) ([]byte, error) {
	return server.RSAEncrypt(publicKey, plaintext)
}

// RSADecrypt 使用 RSA 私钥解密数据（OAEP-SHA256）。
// 委托给 server.RSADecrypt 实现。
func RSADecrypt(privateKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	return server.RSADecrypt(privateKey, ciphertext)
}

// RSASign 使用 RSA 私钥签名数据（PKCS1v15-SHA256）。
// 委托给 server.RSASign 实现。
func RSASign(privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	return server.RSASign(privateKey, data)
}

// RSAVerify 使用 RSA 公钥验证签名。
// 委托给 server.RSAVerify 实现。
func RSAVerify(publicKey *rsa.PublicKey, data, signature []byte) error {
	return server.RSAVerify(publicKey, data, signature)
}

// MarshalRSAPrivateKey 将 RSA 私钥编码为 PEM 格式。
// 委托给 server.MarshalRSAPrivateKey 实现。
func MarshalRSAPrivateKey(privateKey *rsa.PrivateKey) []byte {
	return server.MarshalRSAPrivateKey(privateKey)
}

// MarshalRSAPublicKey 将 RSA 公钥编码为 PEM 格式。
// 委托给 server.MarshalRSAPublicKey 实现。
func MarshalRSAPublicKey(publicKey *rsa.PublicKey) ([]byte, error) {
	return server.MarshalRSAPublicKey(publicKey)
}

// ParseRSAPrivateKey 从 PEM 数据解析 RSA 私钥。
// 委托给 server.ParseRSAPrivateKey 实现。
func ParseRSAPrivateKey(pemData []byte) (*rsa.PrivateKey, error) {
	return server.ParseRSAPrivateKey(pemData)
}

// ParseRSAPublicKey 从 PEM 数据解析 RSA 公钥。
// 委托给 server.ParseRSAPublicKey 实现。
func ParseRSAPublicKey(pemData []byte) (*rsa.PublicKey, error) {
	return server.ParseRSAPublicKey(pemData)
}
