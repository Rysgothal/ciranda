package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

func NewToken() (plain string, hash [32]byte, err error) {
	raw := make([]byte, 32)

	if _, err := rand.Read(raw); err != nil {
		return "", [32]byte{}, fmt.Errorf("generate token entropy: %w", err)
	}

	plain = base64.RawURLEncoding.EncodeToString(raw)
	return plain, sha256.Sum256([]byte(plain)), nil
}

func HashToken(plain string) [32]byte {
	return sha256.Sum256([]byte(plain))
}
