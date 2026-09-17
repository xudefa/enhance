package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	goNet "net"
	"testing"
	"time"
)

func testSelfSignedCertVerifyPEM(t *testing.T, certPEM []byte) {
	t.Helper()
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

	testSelfSignedCertVerifyPEM(t, certPEM)
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
	padded := make([]byte, 16)
	for i := range padded {
		padded[i] = 0
	}
	padded[15] = 0 // padding value 0 is invalid
	_, err := pkcs7Unpad(padded, 16)
	if err == nil {
		t.Fatal("expected error for padding value 0")
	}
}

func TestPKCS7Unpad_InvalidPaddingValueTooLarge(t *testing.T) {
	t.Parallel()
	padded := make([]byte, 16)
	for i := range padded {
		padded[i] = byte(17) // padding value > blockSize
	}
	_, err := pkcs7Unpad(padded, 16)
	if err == nil {
		t.Fatal("expected error for padding value > blockSize")
	}
}

func TestPKCS7Unpad_InconsistentPadding(t *testing.T) {
	t.Parallel()
	padded := make([]byte, 16)
	for i := range padded {
		padded[i] = byte(4)
	}
	padded[15] = 4
	padded[14] = 3 // inconsistent padding byte
	_, err := pkcs7Unpad(padded, 16)
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
