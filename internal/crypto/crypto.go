package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// ErrInvalidKey is returned when the encryption key is not 32 bytes (AES-256).
var ErrInvalidKey = errors.New("encryption key must be 32 bytes (64 hex chars)")

// ErrDecryptFailed is returned when decryption fails (wrong key or corrupted data).
var ErrDecryptFailed = errors.New("decryption failed: invalid key or corrupted data")

// Encrypt encrypts plaintext using AES-256-GCM with the given hex-encoded key.
// Returns hex-encoded ciphertext (nonce + sealed data).
func Encrypt(plaintext string, hexKey string) (string, error) {
	key, err := decodeKey(hexKey)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("creating GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generating nonce: %w", err)
	}

	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(sealed), nil
}

// Decrypt decrypts hex-encoded ciphertext using AES-256-GCM with the given hex-encoded key.
func Decrypt(hexCiphertext string, hexKey string) (string, error) {
	key, err := decodeKey(hexKey)
	if err != nil {
		return "", err
	}

	ciphertext, err := hex.DecodeString(hexCiphertext)
	if err != nil {
		return "", fmt.Errorf("decoding ciphertext: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("creating GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", ErrDecryptFailed
	}

	nonce, ciphertextData := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextData, nil)
	if err != nil {
		return "", ErrDecryptFailed
	}

	return string(plaintext), nil
}

// GenerateKey generates a random 32-byte key and returns it hex-encoded.
func GenerateKey() (string, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("generating key: %w", err)
	}
	return hex.EncodeToString(key), nil
}

// DeriveKey derives a 32-byte AES-256 key from a password using SHA-256.
// Returns the key as a 64-char hex string compatible with Encrypt/Decrypt.
func DeriveKey(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func decodeKey(hexKey string) ([]byte, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, ErrInvalidKey
	}
	if len(key) != 32 {
		return nil, ErrInvalidKey
	}
	return key, nil
}
