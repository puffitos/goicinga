package util

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// NewTLSConfig creates a new TLS config. If the cert path is empty,
// the system pool is returned with insecureSkipVerify set to true.
func NewTLSConfig(certPath string) *tls.Config {
	pool := loadCertPool(certPath)
	return &tls.Config{
		RootCAs:            pool,
		InsecureSkipVerify: certPath == "", //nolint:gosec
	}
}

// loadCertPool loads a certificate pool from a file. Panics on error.
func loadCertPool(path string) *x509.CertPool {
	if path == "" {
		pool, err := x509.SystemCertPool()
		if err != nil {
			panic(fmt.Sprintf("failed to load system certificate pool: %v", err))
		}
		return pool
	}
	certPool := x509.NewCertPool()
	pem, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("failed to read cert file: %v", err))
	}
	if !certPool.AppendCertsFromPEM(pem) {
		panic("failed to append certificate to pool")
	}
	return certPool
}
