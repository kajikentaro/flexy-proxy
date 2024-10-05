package test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	test_utils "github.com/kajikentaro/flexy-proxy/integration_test"
	"github.com/kajikentaro/flexy-proxy/utils"
	"github.com/stretchr/testify/assert"
)

var PROXY_PORT_NUMBER = 8090
var PROXY_HTTP_ADDRESS = fmt.Sprintf(":%d", PROXY_PORT_NUMBER)
var PROXY_URL, _ = url.Parse(fmt.Sprintf("http://localhost:%d", PROXY_PORT_NUMBER))

func TestCertificateOption(t *testing.T) {
	err := createCertificates()
	assert.NoError(t, err)

	config, err := utils.ReadConfigYaml("certificate_test.yaml")
	assert.NoError(t, err, "failed to parse config")

	{
		// create a proxy server
		ctx, cancel := context.WithCancel(context.Background())
		err = test_utils.StartProxyServer(ctx, PROXY_HTTP_ADDRESS, config)
		assert.NoError(t, err, "failed to start a proxy server")
		defer cancel()
	}

	tlsConfig, err := readCertificates()
	assert.NoError(t, err)

	t.Run("should get error if we don't specify the created CA", func(t *testing.T) {
		client := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(PROXY_URL),
			},
		}
		_, err := client.Get("https://sample.test")

		var target *tls.CertificateVerificationError
		assert.ErrorAs(t, err, &target)
	})

	t.Run("should not get error if we specify the created CA", func(t *testing.T) {
		client := &http.Client{
			Transport: &http.Transport{
				Proxy:           http.ProxyURL(PROXY_URL),
				TLSClientConfig: tlsConfig,
			},
		}
		assert.False(t, tlsConfig.InsecureSkipVerify)

		res, err := client.Get("https://sample.test")
		assert.NoError(t, err)
		defer res.Body.Close()
	})
}

func readCertificates() (*tls.Config, error) {
	caCert, err := os.ReadFile("server.crt")
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}
	return tlsConfig, nil
}

// create "server.crt" and "server.key" for the testing
func createCertificates() error {
	// Step 1: Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Save the private key to a file
	keyFile, err := os.Create("server.key")
	if err != nil {
		return err
	}
	defer keyFile.Close()

	pem.Encode(keyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Step 2: Create a CSR
	subject := pkix.Name{
		Country:            []string{"JP"},
		Province:           []string{"Tokyo"},
		Locality:           []string{"Minato"},
		Organization:       []string{"Example Company"},
		OrganizationalUnit: []string{"IT Department"},
		CommonName:         "example.com",
	}

	csrTemplate := x509.CertificateRequest{
		Subject: subject,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privateKey)
	if err != nil {
		return err
	}

	// Save the CSR to a file
	csrFile, err := os.Create("server.csr")
	if err != nil {
		return err
	}
	defer csrFile.Close()

	pem.Encode(csrFile, &pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrBytes,
	})

	// Step 3: Self-sign the CSR to create a certificate
	certTemplate := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10 years
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}

	// Save the certificate to a file
	certFile, err := os.Create("server.crt")
	if err != nil {
		return err
	}
	defer certFile.Close()

	err = pem.Encode(certFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return err
	}
	return nil
}
