package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

func roleToMSP(role string) string {
	switch role {
	case "ICC":
		return "ICCMSP"
	case "Board1":
		return "Board1MSP"
	case "Board2":
		return "Board2MSP"
	default:
		return ""
	}
}

func loadCert(certPEM []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, errors.New("failed to parse cert")
	}
	return x509.ParseCertificate(block.Bytes)
}

func newSigner(privateKey crypto.PrivateKey) func([]byte) ([]byte, error) {
	signer := privateKey.(crypto.Signer)
	return func(digest []byte) ([]byte, error) {
		return signer.Sign(nil, digest, nil)
	}
}

func loadPrivateKey(path string) (*ecdsa.PrivateKey, error) {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey.(*ecdsa.PrivateKey), nil
}

