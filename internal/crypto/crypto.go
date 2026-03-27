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

	"golang.org/x/crypto/argon2"
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

// Argon2id parameters for key derivation.
const (
	argon2Time    = 1
	argon2Memory  = 64 * 1024 // 64 MB
	argon2Threads = 4
	argon2KeyLen  = 32
	argon2SaltLen = 16
)

// DeriveKey derives a 32-byte AES-256 key from a password using SHA-256.
// Returns the key as a 64-char hex string compatible with Encrypt/Decrypt.
//
// Deprecated: Use DeriveKeyArgon2 for new exports. Kept for backward-compatible import of old config backups.
func DeriveKey(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// DeriveKeyArgon2 derives a 32-byte AES-256 key from a password using Argon2id.
// Returns the hex-encoded key and the random salt used.
func DeriveKeyArgon2(password string) (hexKey string, salt []byte, err error) {
	salt = make([]byte, argon2SaltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", nil, fmt.Errorf("generating salt: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
	return hex.EncodeToString(key), salt, nil
}

// DeriveKeyWithSalt derives a 32-byte AES-256 key from a password and a known salt using Argon2id.
// Returns the hex-encoded key.
func DeriveKeyWithSalt(password string, salt []byte) string {
	key := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
	return hex.EncodeToString(key)
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
