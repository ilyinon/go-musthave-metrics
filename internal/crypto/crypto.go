package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// ParseRSAPublicKey to agent
func ParseRSAPublicKey(data []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("invalid PEM data")
	}
	pubIfc, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub, ok := pubIfc.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}
	return pub, nil
}

// ParseRSAPrivateKey to support PKCS#1 and PKCS#8
func ParseRSAPrivateKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("invalid PEM data")
	}

	// PKCS#1
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	// PKCS#8
	keyIfc, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := keyIfc.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}
	return key, nil
}

// EncryptRSA to agent
func EncryptRSA(pub *rsa.PublicKey, data []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, pub, data)
}

// DecryptRSA to server
func DecryptRSA(priv *rsa.PrivateKey, data []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, priv, data)
}
