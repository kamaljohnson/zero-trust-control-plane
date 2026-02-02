package security

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"strings"
)

// ErrInvalidKey is returned when PEM or key type is invalid.
var ErrInvalidKey = errors.New("invalid key")

// LoadPEM reads content from path if s does not look like inline PEM; otherwise returns s as bytes.
func LoadPEM(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, ErrInvalidKey
	}
	if strings.HasPrefix(s, "-----BEGIN") {
		return []byte(s), nil
	}
	return os.ReadFile(s)
}

// ParsePrivateKey parses a PEM-encoded private key (RSA or ECDSA). s may be inline PEM or a file path.
func ParsePrivateKey(s string) (crypto.Signer, error) {
	pemBytes, err := LoadPEM(s)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, ErrInvalidKey
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		signer, ok := key.(crypto.Signer)
		if !ok {
			return nil, ErrInvalidKey
		}
		return signer, nil
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(block.Bytes)
	default:
		return nil, ErrInvalidKey
	}
}

// ParsePublicKey parses a PEM-encoded public key (RSA or ECDSA). s may be inline PEM or a file path.
func ParsePublicKey(s string) (crypto.PublicKey, error) {
	pemBytes, err := LoadPEM(s)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, ErrInvalidKey
	}
	switch block.Type {
	case "RSA PUBLIC KEY":
		return x509.ParsePKCS1PublicKey(block.Bytes)
	case "PUBLIC KEY":
		return x509.ParsePKIXPublicKey(block.Bytes)
	default:
		return nil, ErrInvalidKey
	}
}

// KeyAlg returns "RS256" for RSA and "ES256" for ECDSA P-256; empty otherwise.
func KeyAlg(pub crypto.PublicKey) string {
	switch pub.(type) {
	case *rsa.PublicKey:
		return "RS256"
	case *ecdsa.PublicKey:
		return "ES256"
	default:
		return ""
	}
}
