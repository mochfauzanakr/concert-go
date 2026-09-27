package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
)

func deriveAESKey(secret string) ([]byte, error) {
	key, err := hex.DecodeString(secret)
	if err != nil {
		return nil, errors.New("aes secret must be valid hex string")
	}
	if len(key) != 32 {
		return nil, errors.New("aes secret must be 32 bytes (64 hex chars)")
	}
	return key, nil
}

// EncryptAES encrypts plaintext using AES-256-GCM. Returns base64-encoded ciphertext.
func EncryptAES(secret, plainText string) (string, error) {
	key, err := deriveAESKey(secret)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// nonce + ciphertext + tag
	sealed := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return hex.EncodeToString(sealed), nil
}

// DecryptAES decrypts hex-encoded AES-256-GCM ciphertext back to plaintext.
func DecryptAES(secret, cipherHex string) (string, error) {
	key, err := deriveAESKey(secret)
	if err != nil {
		return "", err
	}

	data, err := hex.DecodeString(cipherHex)
	if err != nil {
		return "", errors.New("invalid ciphertext encoding")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("decryption failed: invalid key or tampered data")
	}

	return string(plaintext), nil
}
