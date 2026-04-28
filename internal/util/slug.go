package util

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	slugLen  = 8
	slugChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func GenerateSlug() (string, error) {
	b := make([]byte, slugLen)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(slugChars))))
		if err != nil {
			return "", fmt.Errorf("generate slug: %w", err)
		}
		b[i] = slugChars[n.Int64()]
	}
	return string(b), nil
}
