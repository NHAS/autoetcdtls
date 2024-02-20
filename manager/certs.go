package manager

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

const (
	CACertFileName = "ca_cert.pem"
	CAKeyFileName  = "ca_key.pem"

	PeerCertFileName = "peer_cert.pem"
	PeerKeyFileName  = "peer_key.pem"
)

func generateCA(dirPath string) (*x509.Certificate, *rsa.PrivateKey, error) {
	// set up our CA certificate
	n, err := rand.Int(rand.Reader, big.NewInt(4000))
	if err != nil {
		return nil, nil, err
	}

	ca := &x509.Certificate{
		SerialNumber: n,
		Subject: pkix.Name{
			CommonName: "Wag Cluster CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(30, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// create our private and public key
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	err = writeCert(dirPath, CACertFileName, CAKeyFileName, caBytes, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	return ca, caPrivKey, nil
}

func generatePeer(path, domain string, caCert *x509.Certificate, caKey *rsa.PrivateKey) error {
	n, err := rand.Int(rand.Reader, big.NewInt(4000))
	if err != nil {
		return err
	}

	// set up our server certificate
	cert := &x509.Certificate{
		SerialNumber: n,
		Subject: pkix.Name{
			CommonName: "Wag",
		},
		DNSNames:     []string{domain},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(30, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, caCert, &certPrivKey.PublicKey, caKey)
	if err != nil {
		return err
	}

	return writeCert(path, PeerCertFileName, PeerKeyFileName, certBytes, certPrivKey)
}

func writeCert(path, certName, keyName string, certBytes []byte, certPrivKey *rsa.PrivateKey) error {

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})

	err := os.MkdirAll(path, 0700)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(path, certName), certPEM.Bytes(), 0600)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(path, keyName), certPrivKeyPEM.Bytes(), 0600)
	if err != nil {
		return err
	}

	return nil
}

func loadCert(path, cert string) (*x509.Certificate, error) {

	certBytes, err := os.ReadFile(filepath.Join(path, cert))
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(certBytes)
	if block == nil {
		return nil, errors.New("failed to parse pem block for certificate")
	}

	certData, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	return certData, nil
}

func loadKey(path, key string) (*rsa.PrivateKey, error) {
	keyBytes, err := os.ReadFile(filepath.Join(path, key))
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, errors.New("failed to parse pem block for certificate")
	}

	rsaPriv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return rsaPriv, nil

}

func createOrLoadCerts(path, domain string) (err error) {

	if _, err := os.Stat(filepath.Join(path, PeerKeyFileName)); err == nil {

		_, err := loadCert(path, PeerCertFileName)
		if err != nil {
			return err
		}
		return nil
	}

	var (
		caCert *x509.Certificate
		caKey  *rsa.PrivateKey
	)
	if _, err := os.Stat(filepath.Join(path, CAKeyFileName)); errors.Is(err, os.ErrNotExist) {

		caCert, caKey, err = generateCA(path)
		if err != nil {
			return err
		}

	} else if err == nil {

		caCert, err = loadCert(path, CACertFileName)
		if err != nil {
			return err
		}

		caKey, err = loadKey(path, CAKeyFileName)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	return generatePeer(path, domain, caCert, caKey)
}
