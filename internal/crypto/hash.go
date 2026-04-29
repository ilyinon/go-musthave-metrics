package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// HashSHA256 calculates an HMAC-SHA256 hash of the given body using the provided key.
func HashSHA256(body []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}
