package jdownloader

import (
	"bytes"
	"crypto/aes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestDeriveLoginSecret(t *testing.T) {
	t.Parallel()
	secret := deriveLoginSecret("Test@Email.com", "mypassword")
	if len(secret) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(secret))
	}

	// Manually compute: SHA256(lower("Test@Email.com") + "mypassword" + "server")
	h := sha256.New()
	h.Write([]byte("test@email.com"))
	h.Write([]byte("mypassword"))
	h.Write([]byte("server"))
	expected := h.Sum(nil)

	if !bytes.Equal(secret, expected) {
		t.Fatalf("loginSecret mismatch:\ngot:  %x\nwant: %x", secret, expected)
	}
}

func TestDeriveDeviceSecret(t *testing.T) {
	t.Parallel()
	secret := deriveDeviceSecret("Test@Email.com", "mypassword")

	h := sha256.New()
	h.Write([]byte("test@email.com"))
	h.Write([]byte("mypassword"))
	h.Write([]byte("device"))
	expected := h.Sum(nil)

	if !bytes.Equal(secret, expected) {
		t.Fatalf("deviceSecret mismatch:\ngot:  %x\nwant: %x", secret, expected)
	}
}

func TestDeriveSecretEmailLowercase(t *testing.T) {
	t.Parallel()
	s1 := deriveLoginSecret("User@EXAMPLE.COM", "pass")
	s2 := deriveLoginSecret("user@example.com", "pass")
	if !bytes.Equal(s1, s2) {
		t.Fatal("email case should not affect secret derivation")
	}
}

func TestDeriveSecretDifferentDomains(t *testing.T) {
	t.Parallel()
	login := deriveLoginSecret("a@b.com", "pass")
	device := deriveDeviceSecret("a@b.com", "pass")
	if bytes.Equal(login, device) {
		t.Fatal("login and device secrets should differ")
	}
}

func TestUpdateEncryptionToken(t *testing.T) {
	t.Parallel()
	oldToken := deriveLoginSecret("test@test.com", "pass")
	sessionHex := hex.EncodeToString([]byte("fakesessiontoken"))

	token, err := updateEncryptionToken(oldToken, sessionHex)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(token) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(token))
	}

	// Verify manually
	sessionBytes, _ := hex.DecodeString(sessionHex)
	h := sha256.New()
	h.Write(oldToken)
	h.Write(sessionBytes)
	expected := h.Sum(nil)

	if !bytes.Equal(token, expected) {
		t.Fatalf("token mismatch:\ngot:  %x\nwant: %x", token, expected)
	}
}

func TestUpdateEncryptionTokenInvalidHex(t *testing.T) {
	t.Parallel()
	oldToken := make([]byte, 32)
	_, err := updateEncryptionToken(oldToken, "not-valid-hex!")
	if err == nil {
		t.Fatal("expected error for invalid hex")
	}
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	t.Parallel()
	token := deriveLoginSecret("user@test.com", "password123")
	plaintext := []byte(`{"rid":1,"apiVer":1}`)

	ciphertext, err := encrypt(plaintext, token)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}

	if bytes.Equal(ciphertext, plaintext) {
		t.Fatal("ciphertext should differ from plaintext")
	}

	decrypted, err := decrypt(ciphertext, token)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("roundtrip failed:\ngot:  %s\nwant: %s", decrypted, plaintext)
	}
}

func TestEncryptDecryptLargePayload(t *testing.T) {
	t.Parallel()
	token := deriveLoginSecret("a@b.com", "p")
	plaintext := []byte(strings.Repeat("abcdefghij", 100))

	ciphertext, err := encrypt(plaintext, token)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}

	decrypted, err := decrypt(ciphertext, token)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Fatal("roundtrip failed for large payload")
	}
}

func TestEncryptInvalidTokenLength(t *testing.T) {
	t.Parallel()
	_, err := encrypt([]byte("data"), []byte("short"))
	if err == nil {
		t.Fatal("expected error for short token")
	}
}

func TestDecryptInvalidTokenLength(t *testing.T) {
	t.Parallel()
	_, err := decrypt([]byte("0123456789abcdef"), []byte("short"))
	if err == nil {
		t.Fatal("expected error for short token")
	}
}

func TestDecryptInvalidBlockSize(t *testing.T) {
	t.Parallel()
	token := make([]byte, 32)
	_, err := decrypt([]byte("odd-length"), token)
	if err == nil {
		t.Fatal("expected error for non-block-aligned ciphertext")
	}
}

func TestSign(t *testing.T) {
	t.Parallel()
	key := []byte("test-secret-key-for-hmac")
	query := "/my/connect?email=test@test.com&rid=12345"

	result := sign(key, query)

	// Verify manually
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(query))
	expected := hex.EncodeToString(mac.Sum(nil))

	if result != expected {
		t.Fatalf("signature mismatch:\ngot:  %s\nwant: %s", result, expected)
	}
}

func TestSignDeterministic(t *testing.T) {
	t.Parallel()
	key := []byte("key")
	query := "/path?param=value"
	s1 := sign(key, query)
	s2 := sign(key, query)
	if s1 != s2 {
		t.Fatal("sign should be deterministic")
	}
}

func TestSignDifferentKeys(t *testing.T) {
	t.Parallel()
	query := "/path"
	s1 := sign([]byte("key1"), query)
	s2 := sign([]byte("key2"), query)
	if s1 == s2 {
		t.Fatal("different keys should produce different signatures")
	}
}

func TestPkcs7PadUnpad(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"one byte", []byte{0x42}},
		{"block aligned", bytes.Repeat([]byte{0x01}, aes.BlockSize)},
		{"block minus one", bytes.Repeat([]byte{0x02}, aes.BlockSize - 1)},
		{"large", bytes.Repeat([]byte{0x03}, 100)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			padded := pkcs7Pad(tt.data, aes.BlockSize)
			if len(padded)%aes.BlockSize != 0 {
				t.Fatalf("padded length %d not a multiple of block size", len(padded))
			}
			if len(padded) < len(tt.data)+1 {
				t.Fatal("padded data should be longer than input")
			}

			unpadded, err := pkcs7Unpad(padded)
			if err != nil {
				t.Fatalf("unpad error: %v", err)
			}
			if !bytes.Equal(unpadded, tt.data) {
				t.Fatalf("pad/unpad roundtrip failed:\ngot:  %x\nwant: %x", unpadded, tt.data)
			}
		})
	}
}

func TestPkcs7UnpadInvalid(t *testing.T) {
	t.Parallel()
	if _, err := pkcs7Unpad([]byte{}); err == nil {
		t.Fatal("expected error for empty data")
	}
	if _, err := pkcs7Unpad([]byte{0x00}); err == nil {
		t.Fatal("expected error for zero padding byte")
	}
	if _, err := pkcs7Unpad([]byte{0x05, 0x05}); err == nil {
		t.Fatal("expected error for padding larger than data")
	}
}
