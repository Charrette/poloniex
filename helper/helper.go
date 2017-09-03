package helper

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
)

// HmacSha512 signs provided content with provided secret with format sha521.
func HmacSha512(secret, content string) (string, error) {
	mac := hmac.New(sha512.New, []byte(secret))
	_, err := mac.Write([]byte(content))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(mac.Sum(nil)), nil
}
