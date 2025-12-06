package core

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
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

func generateSelfSignedCert(certPath, keyPath string, logger *zerolog.Logger) error {
	// Check if both files already exist
	if _, err := os.Stat(certPath); !os.IsNotExist(err) {
		if _, err := os.Stat(keyPath); !os.IsNotExist(err) {
			return nil
		}
	}

	logger.Info().Msg("app: Generating new self-signed TLS certificate and key")

	if err := os.MkdirAll(filepath.Dir(certPath), 0755); err != nil {
		return err
	}

	// generate private key with ECDSA (P256)
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Seanime Self-Signed"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	// Create certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	if err != nil {
		return err
	}

	// Save certificate
	certOut, err := os.Create(certPath)
	if err != nil {
		return err
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		_ = certOut.Close()
		return err
	}
	if err := certOut.Close(); err != nil {
		return err
	}
	logger.Debug().Msgf("app: Wrote TLS certificate to %s", certPath)

	// Save private key
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	// Marshal ECDSA key to SEC 1, ASN.1 DER form
	privKeyBytes, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		_ = keyOut.Close()
		return err
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privKeyBytes}); err != nil {
		_ = keyOut.Close()
		return err
	}
	if err := keyOut.Close(); err != nil {
		return err
	}
	logger.Debug().Msgf("app: Wrote TLS private key to %s", keyPath)

	return nil
}
