package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

func TestHybridRSARoundTripSupportsLargePayload(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	payload := bytes.Repeat([]byte("metrics payload "), 1024)

	encrypted, err := EncryptHybridRSA(&privateKey.PublicKey, payload)
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := DecryptHybridRSA(privateKey, encrypted)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(decrypted, payload) {
		t.Fatal("decrypted payload mismatch")
	}
}
