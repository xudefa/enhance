// Package tls 提供 TLS 证书管理和加密工具，用于 enhance 框架。
//
// 该模块提供 TLS 配置、HTTPS 客户端、AES/RSA 加密解密等安全通信和加密功能。
// 适用于需要安全通信和数据加密的场景。
//
// # 架构设计
//
//   - TLSConfig: TLS 配置管理器，负责证书和密钥管理
//   - HTTPSServer: HTTPS 服务器，支持 TLS 通信
//   - AESCipher: AES 加密解密器
//   - RSACipher: RSA 加密解密器
//   - CertificateManager: 证书管理器
//
// # 核心功能
//
//   - TLS 配置: 支持 TLS 1.2/1.3 协议配置
//   - 证书管理: 支持证书加载、验证、更新
//   - AES 加密: 支持 AES-128/256 对称加密
//   - RSA 加密: 支持 RSA 非对称加密
//   - HTTPS 客户端: 支持 HTTPS 请求，验证服务器证书
//
// # 使用方式
//
// TLS 配置：
//
//	cfg := tls.NewTLSConfig()
//	cfg.SetCertificate("/path/to/cert.pem", "/path/to/key.pem")
//	cfg.SetMinVersion(tls.VersionTLS12)
//
// AES 加密：
//
//	cipher := tls.NewAESCipher(key)
//	encrypted, err := cipher.Encrypt(plaintext)
//	decrypted, err := cipher.Decrypt(encrypted)
//
// RSA 加密：
//
//	cipher := tls.NewRSACipher(publicKey)
//	encrypted, err := cipher.Encrypt(plaintext)
//
// # 支持的加密算法
//
//   - AES-128-CBC: AES 128 位 CBC 模式
//   - AES-256-CBC: AES 256 位 CBC 模式
//   - AES-128-GCM: AES 128 位 GCM 模式
//   - AES-256-GCM: AES 256 位 GCM 模式
//   - RSA-2048: RSA 2048 位密钥
//   - RSA-4096: RSA 4096 位密钥
package tls
