package controller

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const (
	EncodingDigits       = "digits"
	EncodingAlphanumeric = "alphanumeric"
	EncodingSymbols      = "symbols"

	charsetDigits  = "0123456789"
	charsetLetters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsetSymbols = "!@#$%^&*()-_=+[]{}|;:,.<>?/"
)

func generatePassword(length int, encoding string) (string, error) {

	var charset string

	switch strings.ToLower(encoding) {
	case EncodingDigits:
		charset = charsetDigits
	case EncodingAlphanumeric:
		charset = charsetDigits + charsetLetters
	case EncodingSymbols:
		charset = charsetDigits + charsetLetters + charsetSymbols
	default:
		charset = charsetDigits + charsetLetters
	}

	result := make([]byte, length)
	charsetLength := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", err
		}

		result[i] = charset[num.Int64()]
	}

	return string(result), nil

}

func needsUpdate(claim map[string]string, secret map[string][]byte) bool {
	if len(claim) != len(secret) {
		return true
	}

	for k, v := range claim {
		secretValue, exists := secret[k]
		if !exists || string(secretValue) != v {
			return true
		}
	}

	return false
}
