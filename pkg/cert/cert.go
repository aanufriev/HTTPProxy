package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"time"
)

const (
	certDir  = "certs" + string(os.PathSeparator)
	certFile = "cert.pem"
	keyFile  = "key.pem"
)

func GetCertificate(host string) (tls.Certificate, error) {
	cert, err := getCertificateFromDisk(host)
	if err != nil {
		return createCertificate(host)
	}

	return cert, nil
}

func getCertPathByHost(host string) string {
	return certDir + host + string(os.PathSeparator)
}

func getCertificateFromDisk(host string) (tls.Certificate, error) {
	dir := getCertPathByHost(host)

	return tls.LoadX509KeyPair(dir+certFile, dir+keyFile)
}

func createCertificate(host string) (tls.Certificate, error) {
	notBefore := time.Now()
	notAfter := notBefore.AddDate(1, 0, 0)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"http proxy"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		BasicConstraintsValid: true,
	}

	if ip := net.ParseIP(host); ip != nil {
		template.IPAddresses = append(template.IPAddresses, ip)
	} else {
		template.DNSNames = append(template.DNSNames, host)
	}

	rootCA, err := tls.LoadX509KeyPair("ca.crt", "ca.key")
	if err != nil {
		return tls.Certificate{}, err
	}

	if rootCA.Leaf, err = x509.ParseCertificate(rootCA.Certificate[0]); err != nil {
		return tls.Certificate{}, err
	}

	template.AuthorityKeyId = rootCA.Leaf.SubjectKeyId

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	template.SubjectKeyId = func(n *big.Int) []byte {
		h := sha1.New()
		h.Write(n.Bytes())
		return h.Sum(nil)
	}(key.N)

	bytes, err := x509.CreateCertificate(
		rand.Reader, &template, rootCA.Leaf, &key.PublicKey, rootCA.PrivateKey,
	)
	if err != nil {
		return tls.Certificate{}, err
	}

	dir := getCertPathByHost(host)
	if err = os.MkdirAll(dir, 0755); err != nil {
		return tls.Certificate{}, err
	}

	certOut, err := os.Create(dir + certFile)
	if err != nil {
		return tls.Certificate{}, err
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: bytes}); err != nil {
		return tls.Certificate{}, err
	}

	keyOut, err := os.OpenFile(dir+keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return tls.Certificate{}, err
	}
	defer keyOut.Close()

	if err := pem.Encode(
		keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)},
	); err != nil {
		return tls.Certificate{}, err
	}

	return tls.LoadX509KeyPair(dir+certFile, dir+keyFile)
}
