package internal

import (
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
	"path/filepath"
)

func ReadCert(certFilepath string) (*x509.Certificate, error) {
	bs, err := os.ReadFile(filepath.Clean(certFilepath)) // handle error
	if err != nil {

	}
	block, _ := pem.Decode(bs)
	if block == nil {
		log.Fatal("failed to parse PEM block containing the public key")
	}

	return x509.ParseCertificate(block.Bytes) // handle error
}
