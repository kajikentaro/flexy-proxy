package configs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kajikentaro/flexy-proxy/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// https://github.com/kajikentaro/flexy-proxy/issues/7
func TestAlwaysMitmWithRegex(t *testing.T) {
	proxyConfig, err := parseRawConfig(
		&models.RawConfig{
			Routes:     []models.RouteConf{{Url: "https://example\\.test", Regex: true}, {Url: "https://foo.test"}},
			AlwaysMitm: true,
		},
		".",
	)
	require.NoError(t, err)

	assert.Empty(t, proxyConfig.HttpsHostNames)
}

func createTestCert(t *testing.T, certPath, certKeyPath string) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "localhost"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	require.NoError(t, err)
	certOut := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyOut := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	require.NoError(t, os.WriteFile(certPath, certOut, 0644))
	require.NoError(t, os.WriteFile(certKeyPath, keyOut, 0644))
}

func TestLoadCertificateWithRelativePath(t *testing.T) {
	certName := "cert.pem"
	certKeyName := "key.pem"
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, certName)
	certKeyPath := filepath.Join(tmpDir, certKeyName)

	createTestCert(t, certPath, certKeyPath)

	rawConfig := models.RawConfig{
		Certificate:    certName,
		CertificateKey: certKeyName,
	}
	proxyConfig, err := parseRawConfig(&rawConfig, tmpDir /* directory which has the cert files */)
	require.NoError(t, err)
	assert.NotNil(t, proxyConfig.Certificate)
}

func TestLoadCertificateWithAbsolutePath(t *testing.T) {
	certName := "cert.pem"
	certKeyName := "key.pem"
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, certName)
	certKeyPath := filepath.Join(tmpDir, certKeyName)

	createTestCert(t, certPath, certKeyPath)

	rawConfig := models.RawConfig{
		Certificate:    certPath,
		CertificateKey: certKeyPath,
	}
	proxyConfig, err := parseRawConfig(&rawConfig, "." /* directory which doesn't have files */)
	require.NoError(t, err)
	assert.NotNil(t, proxyConfig.Certificate)
}
