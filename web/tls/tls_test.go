package tls

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/xudefa/enhance/web/server"
	"math/big"
	goNet "net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func generateTestCert(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey error = %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []goNet.IP{goNet.ParseIP("127.0.0.1")},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("CreateCertificate error = %v", err)
	}

	certPEM = pemEncode("CERTIFICATE", certDER)
	keyPEM = pemEncode("RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(privateKey))
	return certPEM, keyPEM
}

func pemEncode(typ string, derBytes []byte) []byte {
	block := &pem.Block{
		Type:  typ,
		Bytes: derBytes,
	}
	return pem.EncodeToMemory(block)
}

func TestLoadCertFromPEM(t *testing.T) {
	t.Parallel()
	certPEM, keyPEM := generateTestCert(t)

	tlsCfg, err := LoadCertFromPEM(certPEM, keyPEM)
	if err != nil || tlsCfg == nil || len(tlsCfg.Certificates) != 1 {
		t.Fatalf("expected valid tls config with 1 certificate, got err=%v, cfg=%v", err, tlsCfg != nil)
	}
}

func TestInsecureTLSConfig(t *testing.T) {
	t.Parallel()
	cfg := InsecureTLSConfig()
	if cfg == nil || !cfg.InsecureSkipVerify || cfg.MinVersion != tls.VersionTLS12 {
		t.Fatalf("expected valid insecure config, got cfg=%v", cfg != nil)
	}
}

func TestMustLoadTLSConfig_Panics(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid cert file")
		}
	}()
	_ = MustLoadTLSConfig("nonexistent.pem", "nonexistent.key")
}

func TestNewClient(t *testing.T) {
	t.Parallel()
	certPEM, keyPEM := generateTestCert(t)

	tlsCfg, err := LoadCertFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("LoadCertFromPEM() error = %v", err)
	}

	client := server.NewTLSClient("https://localhost:8443",
		server.WithTLSConfig(tlsCfg),
		server.WithTLSRequestTimeout(5*time.Second),
		server.WithTLSDefaultHeader("User-Agent", "enhance-test"),
	)

	if client == nil {
		t.Fatal("NewTLSClient() returned nil")
	}
}

func TestNewClient_Insecure(t *testing.T) {
	t.Parallel()
	client := server.NewTLSClient("https://example.com",
		server.WithInsecureTLS(),
	)
	if client == nil {
		t.Fatal("NewTLSClient() returned nil")
	}
}

func TestHttpsClientIntegration(t *testing.T) {
	t.Parallel()
	certPEM, keyPEM := generateTestCert(t)

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("X509KeyPair error = %v", err)
	}

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))

	ts.TLS = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	ts.StartTLS()
	defer ts.Close()

	client := server.NewTLSClient(ts.URL,
		server.WithInsecureTLS(),
	)

	resp, err := client.Get(context.Background(), "/")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestHttpsClientIntegration_AllMethods(t *testing.T) {
	t.Parallel()

	client, ctx := testTLSHttpsClientIntegrationSetup(t)

	t.Run("GET", func(t *testing.T) {
		testTLSHttpsClientIntegrationGet(t, client, ctx)
	})

	t.Run("POST", func(t *testing.T) {
		testTLSHttpsClientIntegrationPost(t, client, ctx)
	})

	t.Run("PUT", func(t *testing.T) {
		testTLSHttpsClientIntegrationPut(t, client, ctx)
	})

	t.Run("DELETE", func(t *testing.T) {
		testTLSHttpsClientIntegrationDelete(t, client, ctx)
	})
}

func testTLSHttpsClientIntegrationSetup(t *testing.T) (*server.NetClient, context.Context) {
	t.Helper()
	certPEM, keyPEM := generateTestCert(t)

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("X509KeyPair error = %v", err)
	}

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
		case http.MethodPut:
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	ts.TLS = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	ts.StartTLS()
	t.Cleanup(ts.Close)

	return server.NewTLSClient(ts.URL, server.WithInsecureTLS()), context.Background()
}

func testTLSHttpsClientIntegrationGet(t *testing.T, client *server.NetClient, ctx context.Context) {
	t.Helper()
	resp, err := client.Get(ctx, "/")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func testTLSHttpsClientIntegrationPost(t *testing.T, client *server.NetClient, ctx context.Context) {
	t.Helper()
	resp, err := client.Post(ctx, "/", map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
}

func testTLSHttpsClientIntegrationPut(t *testing.T, client *server.NetClient, ctx context.Context) {
	t.Helper()
	resp, err := client.Put(ctx, "/", map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func testTLSHttpsClientIntegrationDelete(t *testing.T, client *server.NetClient, ctx context.Context) {
	t.Helper()
	resp, err := client.Delete(ctx, "/")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}

func TestLoadTLSConfig(t *testing.T) {
	t.Parallel()
	_, err := LoadTLSConfig("nonexistent.pem", "nonexistent.key")
	if err == nil {
		t.Error("expected error for nonexistent files")
	}
}

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

func writeTestCertKeyFiles(t *testing.T, certPEM, keyPEM []byte) (certFile, keyFile string) {
	t.Helper()
	dir := t.TempDir()
	certFile = filepath.Join(dir, "cert.pem")
	keyFile = filepath.Join(dir, "key.pem")
	if err := os.WriteFile(certFile, certPEM, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0o644); err != nil {
		t.Fatal(err)
	}
	return certFile, keyFile
}

func testLoadTLSConfigWithCAValid(t *testing.T) {
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
}

func testLoadTLSConfigWithCAInvalidCA(t *testing.T) {
	t.Parallel()
	certPEM, keyPEM := generateTestCert(t)
	certFile, keyFile := writeTestCertKeyFiles(t, certPEM, keyPEM)
	caFile := filepath.Join(t.TempDir(), "ca.pem")

	_, err := LoadTLSConfigWithCA(certFile, keyFile, caFile)
	if err == nil {
		t.Fatal("expected error for nonexistent CA file")
	}
}

func testLoadTLSConfigWithCAUnparseable(t *testing.T) {
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
}

func testLoadTLSConfigWithCAInvalidKeyPair(t *testing.T) {
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
}

func TestLoadTLSConfigWithCA(t *testing.T) {
	t.Parallel()

	t.Run("valid certs and CA", testLoadTLSConfigWithCAValid)
	t.Run("invalid CA file", testLoadTLSConfigWithCAInvalidCA)
	t.Run("unparseable CA", testLoadTLSConfigWithCAUnparseable)
	t.Run("invalid key pair", testLoadTLSConfigWithCAInvalidKeyPair)
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
