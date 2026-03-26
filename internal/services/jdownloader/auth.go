package jdownloader

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// deriveSecret computes SHA256(lower(email) + password + domain).
// Returns 32 bytes: first 16 = AES IV, last 16 = AES key.
func deriveSecret(email, password, domain string) []byte {
	h := sha256.New()
	h.Write([]byte(strings.ToLower(email)))
	h.Write([]byte(password))
	h.Write([]byte(domain))
	return h.Sum(nil)
}

// deriveLoginSecret returns the server-side secret for authentication.
func deriveLoginSecret(email, password string) []byte {
	return deriveSecret(email, password, "server")
}

// deriveDeviceSecret returns the device-side secret.
func deriveDeviceSecret(email, password string) []byte {
	return deriveSecret(email, password, "device")
}

// updateEncryptionToken derives a new encryption token from the old token
// and the hex-encoded session token received from the server.
func updateEncryptionToken(oldToken []byte, sessionTokenHex string) ([]byte, error) {
	sessionBytes, err := hex.DecodeString(sessionTokenHex)
	if err != nil {
		return nil, fmt.Errorf("decoding session token hex: %w", err)
	}
	h := sha256.New()
	h.Write(oldToken)
	h.Write(sessionBytes)
	return h.Sum(nil), nil
}

// encrypt encrypts data using AES-128-CBC with the given 32-byte token.
// Token split: bytes[0:16] = IV, bytes[16:32] = key.
// Returns raw ciphertext (caller handles base64 encoding if needed).
func encrypt(data, token []byte) ([]byte, error) {
	if len(token) != 32 {
		return nil, fmt.Errorf("token must be 32 bytes, got %d", len(token))
	}
	iv := token[:16]
	key := token[16:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	padded := pkcs7Pad(data, aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)
	return ciphertext, nil
}

// decrypt decrypts AES-128-CBC ciphertext with the given 32-byte token.
// Token split: bytes[0:16] = IV, bytes[16:32] = key.
func decrypt(data, token []byte) ([]byte, error) {
	if len(token) != 32 {
		return nil, fmt.Errorf("token must be 32 bytes, got %d", len(token))
	}
	iv := token[:16]
	key := token[16:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	if len(data)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext length %d is not a multiple of block size", len(data))
	}

	plaintext := make([]byte, len(data))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plaintext, data)
	return pkcs7Unpad(plaintext)
}

// sign creates an HMAC-SHA256 signature of the query string, returned as hex.
func sign(key []byte, query string) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(query))
	return hex.EncodeToString(mac.Sum(nil))
}

// pkcs7Pad pads data to a multiple of blockSize using PKCS#7.
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	pad := make([]byte, padding)
	for i := range pad {
		pad[i] = byte(padding)
	}
	return append(data, pad...)
}

// pkcs7Unpad removes PKCS#7 padding.
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	padding := int(data[len(data)-1])
	if padding == 0 || padding > aes.BlockSize || padding > len(data) {
		return nil, fmt.Errorf("invalid padding size: %d", padding)
	}
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding byte at position %d", i)
		}
	}
	return data[:len(data)-padding], nil
}
