package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateToken() (string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}

	tokenStr := hex.EncodeToString(token)
	return tokenStr, nil
}
