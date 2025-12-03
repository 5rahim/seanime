package core

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

// generateSelfSignedCert checks for the existence of cert/key files and creates them if they are missing.
func generateSelfSignedCert(certPath, keyPath string, logger *zerolog.Logger) error {
	// Check if both files already exist.
	if _, err := os.Stat(certPath); !os.IsNotExist(err) {
		if _, err := os.Stat(keyPath); !os.IsNotExist(err) {
			logger.Debug().Msg("app: Found existing TLS certificate and key.")
			return nil
		}
	}

	logger.Info().Msg("app: Generating new self-signed TLS certificate and key...")

	// Ensure the directory for the certs exists.
	if err := os.MkdirAll(filepath.Dir(certPath), 0755); err != nil {
		return err
	}

	// Generate a new private key
	privKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	// Create a certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Organization: []string{"Seanime Self-Signed"},
			CommonName:   "localhost",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 0, 0),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	// Create the certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	if err != nil {
		return err
	}

	// --- Save the certificate to cert.pem ---
	certOut, err := os.Create(certPath)
	if err != nil {
		return err
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return err
	}
	if err := certOut.Close(); err != nil {
		return err
	}
	logger.Debug().Msgf("app: Wrote TLS certificate to %s", certPath)

	// --- Save the private key to key.pem ---
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return err
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privKeyBytes}); err != nil {
		return err
	}
	if err := keyOut.Close(); err != nil {
		return err
	}
	logger.Debug().Msgf("app: Wrote TLS private key to %s", keyPath)

	return nil
}
