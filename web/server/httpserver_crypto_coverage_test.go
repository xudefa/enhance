package server

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestHttpServerOptions_WithLogger_WithContext(t *testing.T) {
	t.Parallel()
	srv := NewHTTPServer(
		WithHost(":0"),
		WithContext(context.Background()),
	)
	if srv == nil {
		t.Fatal("NewHTTPServer() returned nil")
	}
	srv.SetContext(context.Background())
}

func TestHttpServer_logCtx_WithContext(t *testing.T) {
	t.Parallel()
	srv := NewHTTPServer()
	ctx := srv.logCtx()
	if ctx == nil {
		t.Error("logCtx() should not return nil")
	}
}

func TestHttpServer_logCtx_WithBackgroundContext(t *testing.T) {
	t.Parallel()
	srv := NewHTTPServer(WithContext(context.Background()))
	ctx := srv.logCtx()
	if ctx == nil {
		t.Error("logCtx() should not return nil")
	}
}

func TestHttpServer_wireHandler(t *testing.T) {
	t.Parallel()
	srv := NewHTTPServer()
	if err := srv.wireHandler(); err != nil {
		t.Errorf("wireHandler() error = %v", err)
	}
}

func TestHttpServer_wireHandlerLocked_WithRouter(t *testing.T) {
	t.Parallel()
	srv := NewHTTPServer()
	router := NewRouter()
	srv.SetHandler(router)
	if err := srv.wireHandler(); err != nil {
		t.Errorf("wireHandler() error = %v", err)
	}
}

func TestHttpServer_Start_DefaultServeMux(t *testing.T) {
	t.Parallel()
	srv := NewHTTPServer(WithHost("127.0.0.1:0"))

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Stop(ctx); err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	select {
	case startErr := <-errCh:
		if startErr != nil && startErr != http.ErrServerClosed {
			t.Errorf("Start() error = %v", startErr)
		}
	case <-time.After(3 * time.Second):
		t.Error("Start() did not return")
	}
}

func TestNewTLSClient_WithTransport(t *testing.T) {
	t.Parallel()
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := NewTLSClient("https://localhost:8443",
		WithTLSTransport(transport),
	)
	if client == nil {
		t.Fatal("NewTLSClient() returned nil")
	}
	if client.httpClient.Transport != transport {
		t.Error("custom transport should be used")
	}
}

func TestLoadTLSConfigWithCA(t *testing.T) {
	t.Parallel()
	_, err := LoadTLSConfigWithCA("nonexistent.pem", "nonexistent.key", "nonexistent-ca.pem")
	if err == nil {
		t.Error("expected error for nonexistent files")
	}
}

func TestLoadClientTLSConfig(t *testing.T) {
	t.Parallel()
	_, err := LoadClientTLSConfig("nonexistent-ca.pem")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadTLSConfigWithCA_InvalidCAFile(t *testing.T) {
	t.Parallel()
	certPEM, keyPEM := generateTestCertForCoverage(t)
	tmpDir := t.TempDir()
	certPath := tmpDir + "/server.crt"
	keyPath := tmpDir + "/server.key"
	caPath := tmpDir + "/ca.crt"

	if err := writeFile(certPath, certPEM); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(keyPath, keyPEM); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(caPath, []byte("not a valid cert")); err != nil {
		t.Fatal(err)
	}

	_, err := LoadTLSConfigWithCA(certPath, keyPath, caPath)
	if err == nil {
		t.Error("expected error for invalid CA")
	}
}

func TestLoadClientTLSConfig_InvalidCAFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	caPath := tmpDir + "/invalid-ca.crt"
	if err := writeFile(caPath, []byte("garbage")); err != nil {
		t.Fatal(err)
	}
	_, err := LoadClientTLSConfig(caPath)
	if err == nil {
		t.Error("expected error for invalid CA")
	}
}

func TestParseRSAPrivateKey_PKCS8(t *testing.T) {
	t.Parallel()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey error = %v", err)
	}

	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("MarshalPKCS8PrivateKey error = %v", err)
	}
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: pkcs8Bytes,
	})

	parsed, err := ParseRSAPrivateKey(pemData)
	if err != nil {
		t.Fatalf("ParseRSAPrivateKey() error = %v", err)
	}
	if !privateKey.Equal(parsed) {
		t.Error("parsed key does not match original")
	}
}

func TestParseRSAPrivateKey_PKCS8_NonRSA(t *testing.T) {
	t.Parallel()
	ecdsaKey, err := generateECDSAKey()
	if err != nil {
		t.Skip("ECDSA key generation failed:", err)
	}

	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(ecdsaKey)
	if err != nil {
		t.Fatalf("MarshalPKCS8PrivateKey error = %v", err)
	}
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: pkcs8Bytes,
	})

	_, err = ParseRSAPrivateKey(pemData)
	if err == nil {
		t.Error("expected error for non-RSA PKCS8 key")
	}
}

func TestParseRSAPrivateKey_InvalidPEM(t *testing.T) {
	t.Parallel()
	_, err := ParseRSAPrivateKey([]byte("not pem data"))
	if err == nil {
		t.Error("expected error for invalid PEM")
	}
}

func TestParseRSAPrivateKey_InvalidPKCS1(t *testing.T) {
	t.Parallel()
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: []byte("invalid key data"),
	})
	_, err := ParseRSAPrivateKey(pemData)
	if err == nil {
		t.Error("expected error for invalid PKCS1 data")
	}
}

func TestParseRSAPublicKey_InvalidPEM(t *testing.T) {
	t.Parallel()
	_, err := ParseRSAPublicKey([]byte("not pem data"))
	if err == nil {
		t.Error("expected error for invalid PEM")
	}
}

func TestParseRSAPublicKey_NonRSA(t *testing.T) {
	t.Parallel()
	ecdsaKey, err := generateECDSAKey()
	if err != nil {
		t.Skip("ECDSA key generation failed:", err)
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(&ecdsaKey.PublicKey)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey error = %v", err)
	}
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	_, err = ParseRSAPublicKey(pemData)
	if err == nil {
		t.Error("expected error for non-RSA public key")
	}
}

func TestLoadCertFromPEM_InvalidCert(t *testing.T) {
	t.Parallel()
	_, err := LoadCertFromPEM([]byte("not a cert"), []byte("not a key"))
	if err == nil {
		t.Error("expected error for invalid PEM data")
	}
}

func TestMustLoadTLSConfig_ValidFiles(t *testing.T) {
	t.Parallel()
	certPEM, keyPEM := generateTestCertForCoverage(t)
	tmpDir := t.TempDir()
	certPath := tmpDir + "/server.crt"
	keyPath := tmpDir + "/server.key"

	if err := writeFile(certPath, certPEM); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(keyPath, keyPEM); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("MustLoadTLSConfig() panicked: %v", r)
		}
	}()

	cfg := MustLoadTLSConfig(certPath, keyPath)
	if cfg == nil {
		t.Fatal("MustLoadTLSConfig() returned nil")
	}
	if len(cfg.Certificates) != 1 {
		t.Errorf("Certificates count = %d, want 1", len(cfg.Certificates))
	}
}

func TestLoadTLSConfig_ValidFiles(t *testing.T) {
	t.Parallel()
	certPEM, keyPEM := generateTestCertForCoverage(t)
	tmpDir := t.TempDir()
	certPath := tmpDir + "/server.crt"
	keyPath := tmpDir + "/server.key"

	if err := writeFile(certPath, certPEM); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(keyPath, keyPEM); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadTLSConfig(certPath, keyPath)
	if err != nil {
		t.Fatalf("LoadTLSConfig() error = %v", err)
	}
	if cfg == nil || len(cfg.Certificates) != 1 {
		t.Errorf("expected 1 certificate, got %v", cfg)
	}
}

func TestLoadCertFromPEM_ValidData(t *testing.T) {
	t.Parallel()
	certPEM, keyPEM := generateTestCertForCoverage(t)
	cfg, err := LoadCertFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("LoadCertFromPEM() error = %v", err)
	}
	if cfg == nil || len(cfg.Certificates) != 1 {
		t.Errorf("expected 1 certificate, got %v", cfg)
	}
}

func TestNewExponentialBackoff_Defaults(t *testing.T) {
	t.Parallel()
	eb := NewExponentialBackoff(0, 0)
	if eb.baseDelay != 100*time.Millisecond {
		t.Errorf("baseDelay = %v, want 100ms", eb.baseDelay)
	}
	if eb.maxDelay != 10*time.Second {
		t.Errorf("maxDelay = %v, want 10s", eb.maxDelay)
	}
	if len(eb.retryableStatus) != 4 {
		t.Errorf("retryableStatus count = %d, want 4", len(eb.retryableStatus))
	}
}

func TestNewFixedDelay_Defaults(t *testing.T) {
	t.Parallel()
	fd := NewFixedDelay(0)
	if fd.delay != 1*time.Second {
		t.Errorf("delay = %v, want 1s", fd.delay)
	}
	if len(fd.retryableStatus) != 4 {
		t.Errorf("retryableStatus count = %d, want 4", len(fd.retryableStatus))
	}
}

func TestNewRetryableClient_Defaults(t *testing.T) {
	t.Parallel()
	netClient := NewClient("http://localhost:8080")
	rc := NewRetryableClient(netClient)
	if rc.config.maxAttempts != 3 {
		t.Errorf("maxAttempts = %d, want 3", rc.config.maxAttempts)
	}
	if rc.config.strategy == nil {
		t.Error("strategy should not be nil")
	}
}

func TestNewRetryableClient_ZeroAttempts(t *testing.T) {
	t.Parallel()
	netClient := NewClient("http://localhost:8080")
	rc := NewRetryableClient(netClient, WithMaxAttempts(0))
	if rc.config.maxAttempts != 3 {
		t.Errorf("maxAttempts = %d, want 3 (reset from 0)", rc.config.maxAttempts)
	}
}

func TestNewRetryableClient_NilStrategy(t *testing.T) {
	t.Parallel()
	netClient := NewClient("http://localhost:8080")
	rc := NewRetryableClient(netClient, WithRetryStrategy(nil))
	if rc.config.strategy == nil {
		t.Error("strategy should not be nil (should use default)")
	}
}

func TestExponentialBackoff_MaxSafeShift(t *testing.T) {
	t.Parallel()
	eb := NewExponentialBackoff(100*time.Millisecond, 10*time.Second)
	shift := eb.maxSafeShift()
	if shift == 0 {
		t.Error("maxSafeShift() should be > 0")
	}
}

func TestExponentialBackoff_Delay_HighAttempt(t *testing.T) {
	t.Parallel()
	eb := NewExponentialBackoff(100*time.Millisecond, 1*time.Second)
	delay := eb.Delay(100)
	if delay > 1*time.Second {
		t.Errorf("delay(100) = %v, should be <= maxDelay 1s", delay)
	}
	if delay <= 0 {
		t.Errorf("delay(100) = %v, should be > 0", delay)
	}
}

func TestRandInt64_ZeroOrNegative(t *testing.T) {
	t.Parallel()
	if got := randInt64(0); got != 0 {
		t.Errorf("randInt64(0) = %d, want 0", got)
	}
	if got := randInt64(-1); got != 0 {
		t.Errorf("randInt64(-1) = %d, want 0", got)
	}
}

func generateECDSAKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func generateTestCertForCoverage(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey error = %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{Organization: []string{"Test"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("CreateCertificate error = %v", err)
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	return certPEM, keyPEM
}

func writeFile(path string, data []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}
