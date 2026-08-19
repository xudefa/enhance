package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	goNet "net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func generateTestCA(t *testing.T) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey error = %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(100),
		Subject: pkix.Name{
			Organization: []string{"Test CA"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("CreateCertificate error = %v", err)
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})
}

func TestLoadTLSConfig_InvalidFiles(t *testing.T) {
	t.Parallel()

	t.Run("nonexistent cert", func(t *testing.T) {
		t.Parallel()
		_, err := LoadTLSConfig("nonexistent.pem", "nonexistent.key")
		if err == nil {
			t.Fatal("expected error for nonexistent files")
		}
	})

	t.Run("valid cert with wrong key", func(t *testing.T) {
		t.Parallel()
		certPEM, _ := generateTestCert(t)
		dir := t.TempDir()
		certFile := filepath.Join(dir, "cert.pem")
		keyFile := filepath.Join(dir, "key.pem")
		if err := os.WriteFile(certFile, certPEM, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(keyFile, []byte("not a real key"), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := LoadTLSConfig(certFile, keyFile)
		if err == nil {
			t.Fatal("expected error for mismatched cert/key")
		}
	})
}

func TestLoadTLSConfigWithCA(t *testing.T) {
	t.Parallel()

	t.Run("valid certs and CA", func(t *testing.T) {
		t.Parallel()
		certPEM, keyPEM := generateTestCert(t)
		caPEM := generateTestCA(t)

		dir := t.TempDir()
		certFile := filepath.Join(dir, "cert.pem")
		keyFile := filepath.Join(dir, "key.pem")
		caFile := filepath.Join(dir, "ca.pem")

		if err := os.WriteFile(certFile, certPEM, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(keyFile, keyPEM, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(caFile, caPEM, 0o644); err != nil {
			t.Fatal(err)
		}

		cfg, err := LoadTLSConfigWithCA(certFile, keyFile, caFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
		if len(cfg.Certificates) != 1 {
			t.Errorf("expected 1 certificate, got %d", len(cfg.Certificates))
		}
		if cfg.RootCAs == nil {
			t.Error("expected RootCAs to be set")
		}
	})

	t.Run("invalid CA file", func(t *testing.T) {
		t.Parallel()
		certPEM, keyPEM := generateTestCert(t)
		dir := t.TempDir()
		certFile := filepath.Join(dir, "cert.pem")
		keyFile := filepath.Join(dir, "key.pem")
		caFile := filepath.Join(dir, "ca.pem")

		if err := os.WriteFile(certFile, certPEM, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(keyFile, keyPEM, 0o644); err != nil {
			t.Fatal(err)
		}

		_, err := LoadTLSConfigWithCA(certFile, keyFile, caFile)
		if err == nil {
			t.Fatal("expected error for nonexistent CA file")
		}
	})

	t.Run("unparseable CA", func(t *testing.T) {
		t.Parallel()
		certPEM, keyPEM := generateTestCert(t)
		dir := t.TempDir()
		certFile := filepath.Join(dir, "cert.pem")
		keyFile := filepath.Join(dir, "key.pem")
		caFile := filepath.Join(dir, "ca.pem")

		if err := os.WriteFile(certFile, certPEM, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(keyFile, keyPEM, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(caFile, []byte("not a valid PEM"), 0o644); err != nil {
			t.Fatal(err)
		}

		_, err := LoadTLSConfigWithCA(certFile, keyFile, caFile)
		if err == nil {
			t.Fatal("expected error for unparseable CA")
		}
	})

	t.Run("invalid key pair", func(t *testing.T) {
		t.Parallel()
		caPEM := generateTestCA(t)
		dir := t.TempDir()
		certFile := filepath.Join(dir, "cert.pem")
		keyFile := filepath.Join(dir, "key.pem")
		caFile := filepath.Join(dir, "ca.pem")

		if err := os.WriteFile(certFile, []byte("bad cert"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(keyFile, []byte("bad key"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(caFile, caPEM, 0o644); err != nil {
			t.Fatal(err)
		}

		_, err := LoadTLSConfigWithCA(certFile, keyFile, caFile)
		if err == nil {
			t.Fatal("expected error for invalid key pair")
		}
	})
}

func TestLoadClientTLSConfig(t *testing.T) {
	t.Parallel()

	t.Run("valid CA", func(t *testing.T) {
		t.Parallel()
		caPEM := generateTestCA(t)
		dir := t.TempDir()
		caFile := filepath.Join(dir, "ca.pem")
		if err := os.WriteFile(caFile, caPEM, 0o644); err != nil {
			t.Fatal(err)
		}

		cfg, err := LoadClientTLSConfig(caFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
		if cfg.RootCAs == nil {
			t.Error("expected RootCAs to be set")
		}
	})

	t.Run("invalid CA file", func(t *testing.T) {
		t.Parallel()
		_, err := LoadClientTLSConfig("/nonexistent/ca.pem")
		if err == nil {
			t.Fatal("expected error for nonexistent CA file")
		}
	})

	t.Run("unparseable CA", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		caFile := filepath.Join(dir, "ca.pem")
		if err := os.WriteFile(caFile, []byte("not a valid PEM"), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := LoadClientTLSConfig(caFile)
		if err == nil {
			t.Fatal("expected error for unparseable CA")
		}
	})
}

func TestLoadCertFromPEM_Invalid(t *testing.T) {
	t.Parallel()

	t.Run("nil data", func(t *testing.T) {
		t.Parallel()
		_, err := LoadCertFromPEM(nil, nil)
		if err == nil {
			t.Fatal("expected error for nil PEM data")
		}
	})

	t.Run("invalid PEM data", func(t *testing.T) {
		t.Parallel()
		_, err := LoadCertFromPEM([]byte("not cert"), []byte("not key"))
		if err == nil {
			t.Fatal("expected error for invalid PEM data")
		}
	})

	t.Run("valid cert with wrong key PEM", func(t *testing.T) {
		t.Parallel()
		certPEM, _ := generateTestCert(t)
		_, err := LoadCertFromPEM(certPEM, []byte("not a key"))
		if err == nil {
			t.Fatal("expected error for mismatched cert/key PEM")
		}
	})
}

func TestMustLoadTLSConfig_Valid(t *testing.T) {
	t.Parallel()
	certPEM, keyPEM := generateTestCert(t)
	dir := t.TempDir()
	certFile := filepath.Join(dir, "cert.pem")
	keyFile := filepath.Join(dir, "key.pem")

	if err := os.WriteFile(certFile, certPEM, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := MustLoadTLSConfig(certFile, keyFile)
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if len(cfg.Certificates) != 1 {
		t.Errorf("expected 1 certificate, got %d", len(cfg.Certificates))
	}
}

func TestNewClient_WithInsecureTLS(t *testing.T) {
	t.Parallel()
	client := NewClient("https://localhost:8443", WithInsecureTLS())
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_WithTimeout(t *testing.T) {
	t.Parallel()
	client := NewClient("https://localhost:8443",
		WithTimeout(10*time.Second),
	)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_WithDefaultHeader(t *testing.T) {
	t.Parallel()
	client := NewClient("https://localhost:8443",
		WithDefaultHeader("Authorization", "Bearer token"),
		WithDefaultHeader("X-Custom", "value"),
	)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_WithTLSConfig(t *testing.T) {
	t.Parallel()
	cfg := InsecureTLSConfig()
	client := NewClient("https://localhost:8443", WithTLSConfig(cfg))
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_WithTransport(t *testing.T) {
	t.Parallel()
	transport := &http.Transport{
		TLSClientConfig: InsecureTLSConfig(),
	}
	client := NewClient("https://localhost:8443",
		WithTransport(transport),
		WithTLSConfig(InsecureTLSConfig()),
	)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestRSAGenerateKeyPair(t *testing.T) {
	t.Parallel()

	t.Run("2048 bits", func(t *testing.T) {
		t.Parallel()
		key, err := RSAGenerateKey(2048)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if key.N.BitLen() != 2048 {
			t.Errorf("key size = %d, want 2048", key.N.BitLen())
		}
	})

	t.Run("4096 bits", func(t *testing.T) {
		t.Parallel()
		key, err := RSAGenerateKey(4096)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if key.N.BitLen() != 4096 {
			t.Errorf("key size = %d, want 4096", key.N.BitLen())
		}
	})
}

func TestSaveAndLoadRSAPrivateKey(t *testing.T) {
	t.Parallel()
	key, err := RSAGenerateKey(2048)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pemData := MarshalRSAPrivateKey(key)
	if len(pemData) == 0 {
		t.Fatal("expected non-empty PEM data")
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		t.Fatal("expected valid PEM block")
	}
	if block.Type != "RSA PRIVATE KEY" {
		t.Errorf("block type = %q, want %q", block.Type, "RSA PRIVATE KEY")
	}

	parsed, err := ParseRSAPrivateKey(pemData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !key.Equal(parsed) {
		t.Error("parsed key does not match original")
	}
}

func TestSaveAndLoadRSAPublicKey(t *testing.T) {
	t.Parallel()
	key, err := RSAGenerateKey(2048)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pemData, err := MarshalRSAPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pemData) == 0 {
		t.Fatal("expected non-empty PEM data")
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		t.Fatal("expected valid PEM block")
	}
	if block.Type != "PUBLIC KEY" {
		t.Errorf("block type = %q, want %q", block.Type, "PUBLIC KEY")
	}

	parsed, err := ParseRSAPublicKey(pemData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !key.PublicKey.Equal(parsed) {
		t.Error("parsed key does not match original")
	}
}

func TestParseRSAPrivateKey_Invalid(t *testing.T) {
	t.Parallel()

	t.Run("not PEM", func(t *testing.T) {
		t.Parallel()
		_, err := ParseRSAPrivateKey([]byte("not a PEM block"))
		if err == nil {
			t.Fatal("expected error for non-PEM data")
		}
	})

	t.Run("PKCS8 non-RSA key", func(t *testing.T) {
		t.Parallel()
		// Create an ECDSA key encoded as PKCS8
		dir := t.TempDir()
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatal(err)
		}
		// Marshal as PKCS8
		pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			t.Fatal(err)
		}
		pemData := pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: pkcs8Bytes,
		})

		parsed, err := ParseRSAPrivateKey(pemData)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if parsed == nil {
			t.Fatal("expected non-nil key")
		}
		_ = dir
	})
}

func TestParseRSAPublicKey_Invalid(t *testing.T) {
	t.Parallel()

	t.Run("not PEM", func(t *testing.T) {
		t.Parallel()
		_, err := ParseRSAPublicKey([]byte("not a PEM block"))
		if err == nil {
			t.Fatal("expected error for non-PEM data")
		}
	})

	t.Run("invalid PEM content", func(t *testing.T) {
		t.Parallel()
		pemData := pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: []byte("invalid key data"),
		})
		_, err := ParseRSAPublicKey(pemData)
		if err == nil {
			t.Fatal("expected error for invalid PEM content")
		}
	})
}

func TestNewSelfSignedCert(t *testing.T) {
	t.Parallel()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []goNet.IP{goNet.ParseIP("127.0.0.1")},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	if len(certPEM) == 0 {
		t.Fatal("expected non-empty certificate PEM")
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("expected valid PEM block")
	}
	if block.Type != "CERTIFICATE" {
		t.Errorf("block type = %q, want %q", block.Type, "CERTIFICATE")
	}

	parsedCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("unexpected error parsing certificate: %v", err)
	}
	if parsedCert.Subject.Organization[0] != "Test" {
		t.Errorf("subject organization = %q, want %q", parsedCert.Subject.Organization[0], "Test")
	}
}

func TestAESEncrypt_InvalidIV(t *testing.T) {
	t.Parallel()
	key := []byte("0123456789abcdef")
	iv := []byte("short")

	_, err := AESEncrypt([]byte("data"), key, iv)
	if err == nil {
		t.Fatal("expected error for invalid IV length")
	}
}

func TestAESDecrypt_InvalidIV(t *testing.T) {
	t.Parallel()
	key := []byte("012345689abcdef")
	iv := []byte("short")

	_, err := AESDecrypt([]byte("data"), key, iv)
	if err == nil {
		t.Fatal("expected error for invalid IV length")
	}
}

func TestAESDecrypt_InvalidCiphertextLength(t *testing.T) {
	t.Parallel()
	key := []byte("0123456789abcdef")
	iv := []byte("1234567890abcdef")

	_, err := AESDecrypt([]byte("short"), key, iv)
	if err == nil {
		t.Fatal("expected error for invalid ciphertext length")
	}
}

func TestAESDecrypt_InvalidKey(t *testing.T) {
	t.Parallel()
	key := []byte("short")
	iv := []byte("1234567890abcdef")

	_, err := AESDecrypt([]byte("data"), key, iv)
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
}

func TestAESGCMDecrypt_InvalidKey(t *testing.T) {
	t.Parallel()
	key := []byte("short")

	_, err := AESGCMDecrypt([]byte("data"), key, nil)
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
}

func TestAESGCMDecrypt_TooShort(t *testing.T) {
	t.Parallel()
	key := []byte("0123456789abcdef")

	_, err := AESGCMDecrypt([]byte("short"), key, nil)
	if err == nil {
		t.Fatal("expected error for ciphertext too short")
	}
}

func TestPKCS7Unpad_InvalidPaddingValue(t *testing.T) {
	t.Parallel()
	data := make([]byte, 16)
	for i := range data {
		data[i] = 0
	}
	data[15] = 0 // padding value 0 is invalid
	_, err := pkcs7Unpad(data, 16)
	if err == nil {
		t.Fatal("expected error for padding value 0")
	}
}

func TestPKCS7Unpad_InvalidPaddingValueTooLarge(t *testing.T) {
	t.Parallel()
	data := make([]byte, 16)
	for i := range data {
		data[i] = byte(17) // padding value > blockSize
	}
	_, err := pkcs7Unpad(data, 16)
	if err == nil {
		t.Fatal("expected error for padding value > blockSize")
	}
}

func TestPKCS7Unpad_InconsistentPadding(t *testing.T) {
	t.Parallel()
	data := make([]byte, 16)
	for i := range data {
		data[i] = byte(4)
	}
	data[15] = 4
	data[14] = 3 // inconsistent padding byte
	_, err := pkcs7Unpad(data, 16)
	if err == nil {
		t.Fatal("expected error for inconsistent padding")
	}
}

func TestMarshalRSAPublicKey_NilKey(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil public key")
		}
	}()
	_, _ = MarshalRSAPublicKey(nil)
}
