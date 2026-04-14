import auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func generateCitizenIDHash(citizenID string) string {
	secretKey := []byte(os.Getenv("HASH_SECRET_KEY"))
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(citizenID))
	return hex.EncodeToString(h.Sum(nil))
}