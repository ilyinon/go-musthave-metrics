package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
)

const aesKeySize = 32

type encryptedPayload struct {
	Key   string `json:"key"`
	Nonce string `json:"nonce"`
	Data  string `json:"data"`
}

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

// EncryptHybridRSA encrypts data with AES-GCM and encrypts the AES key with RSA-OAEP.
func EncryptHybridRSA(pub *rsa.PublicKey, data []byte) ([]byte, error) {
	aesKey := make([]byte, aesKeySize)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, aesKey, nil)
	if err != nil {
		return nil, err
	}

	payload := encryptedPayload{
		Key:   base64.StdEncoding.EncodeToString(encryptedKey),
		Nonce: base64.StdEncoding.EncodeToString(nonce),
		Data:  base64.StdEncoding.EncodeToString(gcm.Seal(nil, nonce, data, nil)),
	}

	return json.Marshal(payload)
}

// DecryptHybridRSA decrypts data encrypted by EncryptHybridRSA.
func DecryptHybridRSA(priv *rsa.PrivateKey, data []byte) ([]byte, error) {
	var payload encryptedPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if payload.Key == "" || payload.Nonce == "" || payload.Data == "" {
		return nil, errors.New("invalid encrypted payload")
	}

	encryptedKey, err := base64.StdEncoding.DecodeString(payload.Key)
	if err != nil {
		return nil, err
	}
	nonce, err := base64.StdEncoding.DecodeString(payload.Nonce)
	if err != nil {
		return nil, err
	}
	ciphertext, err := base64.StdEncoding.DecodeString(payload.Data)
	if err != nil {
		return nil, err
	}

	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, encryptedKey, nil)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(nonce) != gcm.NonceSize() {
		return nil, errors.New("invalid nonce size")
	}

	return gcm.Open(nil, nonce, ciphertext, nil)
}
