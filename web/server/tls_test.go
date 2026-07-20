package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	goNet "net"
	"net/http"
	"net/http/httptest"
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
		t.Fatalf("LoadCertFromPEM() failed: err=%v, cfg=%v", err, tlsCfg != nil)
	}
}

func TestInsecureTLSConfig(t *testing.T) {
	t.Parallel()
	cfg := InsecureTLSConfig()
	if cfg == nil || !cfg.InsecureSkipVerify || cfg.MinVersion != tls.VersionTLS12 {
		t.Fatalf("InsecureTLSConfig() failed: cfg=%v", cfg != nil)
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

func TestNewTLSClient(t *testing.T) {
	t.Parallel()
	certPEM, keyPEM := generateTestCert(t)

	tlsCfg, err := LoadCertFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("LoadCertFromPEM() error = %v", err)
	}

	client := NewTLSClient("https://localhost:8443",
		WithTLSConfig(tlsCfg),
		WithTLSRequestTimeout(5*time.Second),
		WithTLSDefaultHeader("User-Agent", "enhance-test"),
	)

	if client == nil {
		t.Fatal("NewTLSClient() returned nil")
	}
}

func TestNewTLSClient_Insecure(t *testing.T) {
	t.Parallel()
	client := NewTLSClient("https://example.com",
		WithInsecureTLS(),
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

	client := NewTLSClient(ts.URL,
		WithInsecureTLS(),
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
	defer ts.Close()

	client := NewTLSClient(ts.URL, WithInsecureTLS())
	ctx := context.Background()

	t.Run("GET", func(t *testing.T) {
		resp, err := client.Get(ctx, "/")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("POST", func(t *testing.T) {
		resp, err := client.Post(ctx, "/", map[string]string{"key": "value"})
		if err != nil {
			t.Fatalf("Post() error = %v", err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusCreated)
		}
	})

	t.Run("PUT", func(t *testing.T) {
		resp, err := client.Put(ctx, "/", map[string]string{"key": "value"})
		if err != nil {
			t.Fatalf("Put() error = %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("DELETE", func(t *testing.T) {
		resp, err := client.Delete(ctx, "/")
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}
		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusNoContent)
		}
	})
}

func TestLoadTLSConfig(t *testing.T) {
	t.Parallel()
	_, err := LoadTLSConfig("nonexistent.pem", "nonexistent.key")
	if err == nil {
		t.Error("expected error for nonexistent files")
	}
}
