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

func TestEncryptDecrypt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		plaintext string
	}{
		{"basic round trip", "my-secret-api-key-12345"},
		{"empty string", ""},
		{"long plaintext (10KB)", string(makeLongPayload(10240))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			key := mustGenerateKey(t)

			ciphertext, err := Encrypt(tt.plaintext, key)
			if err != nil {
				t.Fatalf("encrypt failed: %v", err)
			}

			if tt.plaintext != "" && ciphertext == tt.plaintext {
				t.Error("ciphertext should not equal plaintext")
			}

			decrypted, err := Decrypt(ciphertext, key)
			if err != nil {
				t.Fatalf("decrypt failed: %v", err)
			}

			if decrypted != tt.plaintext {
				t.Errorf("expected %q, got %q", tt.plaintext, decrypted)
			}
		})
	}

	t.Run("nonce is unique", func(t *testing.T) {
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

		d1, _ := Decrypt(ct1, key)
		d2, _ := Decrypt(ct2, key)
		if d1 != plaintext || d2 != plaintext {
			t.Error("both ciphertexts should decrypt to original plaintext")
		}
	})
}

func TestEncryptDecryptErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setup         func(t *testing.T) (ciphertext string, key string)
		expectedError error
	}{
		{
			"wrong key",
			func(t *testing.T) (string, string) {
				t.Helper()
				key1 := mustGenerateKey(t)
				key2 := mustGenerateKey(t)
				ct, err := Encrypt("secret", key1)
				if err != nil {
					t.Fatalf("encrypt failed: %v", err)
				}
				return ct, key2
			},
			ErrDecryptFailed,
		},
		{
			"corrupted data",
			func(t *testing.T) (string, string) {
				t.Helper()
				key := mustGenerateKey(t)
				ct, err := Encrypt("secret", key)
				if err != nil {
					t.Fatalf("encrypt failed: %v", err)
				}
				corrupted := ct[:len(ct)-4] + "ffff"
				return corrupted, key
			},
			nil, // any error
		},
		{
			"too short ciphertext",
			func(t *testing.T) (string, string) {
				t.Helper()
				return "abcd", mustGenerateKey(t)
			},
			nil, // any error
		},
		{
			"key too short",
			func(t *testing.T) (string, string) {
				t.Helper()
				return "test", "abcdef"
			},
			ErrInvalidKey,
		},
		{
			"key not hex",
			func(t *testing.T) (string, string) {
				t.Helper()
				return "test", "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"
			},
			ErrInvalidKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ciphertext, key := tt.setup(t)

			// For key errors, test Encrypt; for data errors, test Decrypt
			var err error
			if tt.expectedError == ErrInvalidKey {
				_, err = Encrypt(ciphertext, key)
			} else {
				_, err = Decrypt(ciphertext, key)
			}

			if err == nil {
				t.Error("expected an error, got nil")
			}
			if tt.expectedError != nil && err != tt.expectedError {
				t.Errorf("expected %v, got %v", tt.expectedError, err)
			}
		})
	}
}

func TestGenerateKey(t *testing.T) {
	t.Parallel()

	t.Run("length and usability", func(t *testing.T) {
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
	})

	t.Run("unique", func(t *testing.T) {
		t.Parallel()
		k1 := mustGenerateKey(t)
		k2 := mustGenerateKey(t)
		if k1 == k2 {
			t.Error("two generated keys should not be equal")
		}
	})
}

func TestDeriveKey(t *testing.T) {
	t.Parallel()

	t.Run("argon2 round trip", func(t *testing.T) {
		t.Parallel()
		password := "my-secure-password"

		hexKey, salt, err := DeriveKeyArgon2(password)
		if err != nil {
			t.Fatalf("DeriveKeyArgon2 failed: %v", err)
		}
		if len(hexKey) != 64 {
			t.Errorf("expected 64 hex chars, got %d", len(hexKey))
		}
		if len(salt) != argon2SaltLen {
			t.Errorf("expected salt of %d bytes, got %d", argon2SaltLen, len(salt))
		}

		ct, err := Encrypt("secret-data", hexKey)
		if err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}
		pt, err := Decrypt(ct, hexKey)
		if err != nil {
			t.Fatalf("decrypt failed: %v", err)
		}
		if pt != "secret-data" {
			t.Errorf("expected 'secret-data', got %q", pt)
		}
	})

	t.Run("with salt deterministic", func(t *testing.T) {
		t.Parallel()
		password := "deterministic-test"
		salt := []byte("0123456789abcdef") // 16 bytes

		key1 := DeriveKeyWithSalt(password, salt)
		key2 := DeriveKeyWithSalt(password, salt)

		if key1 != key2 {
			t.Error("same password + salt should produce the same key")
		}
	})

	t.Run("with salt different salts", func(t *testing.T) {
		t.Parallel()
		password := "same-password"
		salt1 := []byte("salt-aaaaaaaaaa01")
		salt2 := []byte("salt-bbbbbbbbbb02")

		key1 := DeriveKeyWithSalt(password, salt1)
		key2 := DeriveKeyWithSalt(password, salt2)

		if key1 == key2 {
			t.Error("different salts should produce different keys")
		}
	})

	t.Run("argon2 unique salts", func(t *testing.T) {
		t.Parallel()
		password := "same-password"

		key1, salt1, err := DeriveKeyArgon2(password)
		if err != nil {
			t.Fatalf("first call failed: %v", err)
		}
		key2, salt2, err := DeriveKeyArgon2(password)
		if err != nil {
			t.Fatalf("second call failed: %v", err)
		}

		if string(salt1) == string(salt2) {
			t.Error("two calls should generate different salts")
		}
		if key1 == key2 {
			t.Error("different salts should produce different keys")
		}
	})

	t.Run("backward compat SHA-256", func(t *testing.T) {
		t.Parallel()
		password := "compat-test"

		oldKey := DeriveKey(password) //nolint:staticcheck
		ct, err := Encrypt("legacy-data", oldKey)
		if err != nil {
			t.Fatalf("encrypt with old key failed: %v", err)
		}
		pt, err := Decrypt(ct, oldKey)
		if err != nil {
			t.Fatalf("decrypt with old key failed: %v", err)
		}
		if pt != "legacy-data" {
			t.Errorf("expected 'legacy-data', got %q", pt)
		}
	})
}

func makeLongPayload(size int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	return data
}
