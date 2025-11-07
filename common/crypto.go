package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// GenerateEncryptionKey generates a random 32-byte encryption key for AES-256
func GenerateEncryptionKey() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}

// HashEncryptionKey hashes the user's encryption key for storage
// Uses PBKDF2 with SHA256 for key derivation
func HashEncryptionKey(key string, salt string) string {
	if salt == "" {
		// Generate a random salt if not provided
		saltBytes := make([]byte, 16)
		rand.Read(saltBytes)
		salt = hex.EncodeToString(saltBytes)
	}
	
	// Use PBKDF2 with 100,000 iterations
	hash := pbkdf2.Key([]byte(key), []byte(salt), 100000, 32, sha256.New)
	return hex.EncodeToString(hash) + ":" + salt
}

// VerifyEncryptionKey verifies a user's encryption key against the stored hash
func VerifyEncryptionKey(key string, storedHash string) bool {
	// Extract salt from stored hash
	parts := splitString(storedHash, ":", 2)
	if len(parts) != 2 {
		return false
	}
	
	salt := parts[1]
	expectedHash := parts[0]
	
	// Hash the provided key with the same salt
	computedHash := pbkdf2.Key([]byte(key), []byte(salt), 100000, 32, sha256.New)
	computedHashStr := hex.EncodeToString(computedHash)
	
	return computedHashStr == expectedHash
}

// EncryptData encrypts data using AES-256-GCM
// Returns base64-encoded encrypted data and nonce
func EncryptData(plaintext string, keyHex string) (encryptedData string, nonce string, err error) {
	// Decode the hex key
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", "", errors.New("invalid encryption key")
	}

	if len(key) != 32 {
		return "", "", errors.New("encryption key must be 32 bytes for AES-256")
	}

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}

	// Generate nonce
	nonceBytes := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonceBytes); err != nil {
		return "", "", err
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nil, nonceBytes, []byte(plaintext), nil)

	// Encode to base64
	encryptedData = base64.StdEncoding.EncodeToString(ciphertext)
	nonce = base64.StdEncoding.EncodeToString(nonceBytes)

	return encryptedData, nonce, nil
}

// DecryptData decrypts data using AES-256-GCM
// Takes base64-encoded encrypted data and nonce
func DecryptData(encryptedData string, nonce string, keyHex string) (plaintext string, err error) {
	// Decode the hex key
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", errors.New("invalid encryption key")
	}

	if len(key) != 32 {
		return "", errors.New("encryption key must be 32 bytes for AES-256")
	}

	// Decode base64 data
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", errors.New("invalid encrypted data")
	}

	nonceBytes, err := base64.StdEncoding.DecodeString(nonce)
	if err != nil {
		return "", errors.New("invalid nonce")
	}

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Verify nonce size
	if len(nonceBytes) != gcm.NonceSize() {
		return "", errors.New("invalid nonce size")
	}

	// Decrypt the data
	plaintextBytes, err := gcm.Open(nil, nonceBytes, ciphertext, nil)
	if err != nil {
		return "", errors.New("decryption failed")
	}

	return string(plaintextBytes), nil
}

// splitString splits a string by delimiter with a maximum number of parts
func splitString(s string, delimiter string, n int) []string {
	result := []string{}
	current := ""
	delimLen := len(delimiter)
	
	for i := 0; i < len(s); {
		if len(result) < n-1 && i+delimLen <= len(s) && s[i:i+delimLen] == delimiter {
			result = append(result, current)
			current = ""
			i += delimLen
		} else {
			current += string(s[i])
			i++
		}
	}
	
	if current != "" || len(result) < n {
		result = append(result, current)
	}
	
	return result
}
