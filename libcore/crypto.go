package libcore

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
)

// Sha1 вычисляет SHA-1 хеш от данных и возвращает его в виде среза байтов.
func Sha1(data []byte) []byte {
	hash := sha1.Sum(data)
	return hash[:]
}

// Sha256Hex вычисляет SHA-256 хеш от данных и возвращает его в виде шестнадцатеричной строки.
func Sha256Hex(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
