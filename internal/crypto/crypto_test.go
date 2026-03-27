package crypto

import (
	"testing"
)

func mustGenerateKey(t *testing.T) string {
	t.Helper()
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	return key
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	t.Parallel()
	key := mustGenerateKey(t)
	plaintext := "my-secret-api-key-12345"

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// Ciphertext should differ from plaintext
	if ciphertext == plaintext {
		t.Error("ciphertext should not equal plaintext")
	}

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestEncryptDecryptEmptyString(t *testing.T) {
	t.Parallel()
	key := mustGenerateKey(t)

	ciphertext, err := Encrypt("", key)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if decrypted != "" {
		t.Errorf("expected empty string, got %q", decrypted)
	}
}

func TestDecryptWrongKeyFails(t *testing.T) {
	t.Parallel()
	key1 := mustGenerateKey(t)
	key2 := mustGenerateKey(t)

	ciphertext, err := Encrypt("secret", key1)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	_, err = Decrypt(ciphertext, key2)
	if err == nil {
		t.Error("expected error when decrypting with wrong key")
	}
	if err != ErrDecryptFailed {
		t.Errorf("expected ErrDecryptFailed, got %v", err)
	}
}

func TestDecryptCorruptedDataFails(t *testing.T) {
	t.Parallel()
	key := mustGenerateKey(t)

	ciphertext, err := Encrypt("secret", key)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// Corrupt the ciphertext by flipping chars
	corrupted := ciphertext[:len(ciphertext)-4] + "ffff"

	_, err = Decrypt(corrupted, key)
	if err == nil {
		t.Error("expected error when decrypting corrupted data")
	}
}

func TestDecryptTooShortFails(t *testing.T) {
	t.Parallel()
	key := mustGenerateKey(t)

	_, err := Decrypt("abcd", key)
	if err == nil {
		t.Error("expected error for too-short ciphertext")
	}
}

func TestNonceIsUnique(t *testing.T) {
	t.Parallel()
	key := mustGenerateKey(t)
	plaintext := "same-plaintext"

	ct1, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt 1 failed: %v", err)
	}

	ct2, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt 2 failed: %v", err)
	}

	if ct1 == ct2 {
		t.Error("encrypting same plaintext twice should produce different ciphertexts (unique nonce)")
	}

	// Both should decrypt correctly
	d1, _ := Decrypt(ct1, key)
	d2, _ := Decrypt(ct2, key)
	if d1 != plaintext || d2 != plaintext {
		t.Error("both ciphertexts should decrypt to original plaintext")
	}
}

func TestInvalidKeyTooShort(t *testing.T) {
	t.Parallel()
	_, err := Encrypt("test", "abcdef")
	if err != ErrInvalidKey {
		t.Errorf("expected ErrInvalidKey, got %v", err)
	}
}

func TestInvalidKeyNotHex(t *testing.T) {
	t.Parallel()
	_, err := Encrypt("test", "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	if err != ErrInvalidKey {
		t.Errorf("expected ErrInvalidKey, got %v", err)
	}
}

func TestGenerateKeyLength(t *testing.T) {
	t.Parallel()
	key := mustGenerateKey(t)

	// 32 bytes = 64 hex chars
	if len(key) != 64 {
		t.Errorf("expected 64 hex chars, got %d", len(key))
	}

	// Should be valid for encrypt/decrypt
	ct, err := Encrypt("test", key)
	if err != nil {
		t.Fatalf("encrypt with generated key failed: %v", err)
	}
	pt, err := Decrypt(ct, key)
	if err != nil {
		t.Fatalf("decrypt with generated key failed: %v", err)
	}
	if pt != "test" {
		t.Errorf("expected 'test', got %q", pt)
	}
}

func TestGenerateKeyUnique(t *testing.T) {
	t.Parallel()
	k1 := mustGenerateKey(t)
	k2 := mustGenerateKey(t)
	if k1 == k2 {
		t.Error("two generated keys should not be equal")
	}
}

func TestLongPlaintext(t *testing.T) {
	t.Parallel()
	key := mustGenerateKey(t)

	// 10KB of data
	long := make([]byte, 10240)
	for i := range long {
		long[i] = byte(i % 256)
	}
	plaintext := string(long)

	ct, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	pt, err := Decrypt(ct, key)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if pt != plaintext {
		t.Error("round-trip failed for long plaintext")
	}
}
