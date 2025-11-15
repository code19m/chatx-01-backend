package hasher

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
)

// Hasher defines the interface for password hashing operations.
type Hasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

type pbkdf2Hasher struct {
	iterations int
	saltLen    int
	keyLen     int
}

// NewHasher creates a new PBKDF2-based password hasher using only standard library.
// iterations: number of iterations (default: 100000 for security)
// saltLen: length of random salt in bytes (default: 16)
// keyLen: length of derived key in bytes (default: 32)
func NewHasher(iterations, saltLen, keyLen int) Hasher {
	if iterations <= 0 {
		iterations = 100000 // OWASP recommended minimum
	}
	if saltLen <= 0 {
		saltLen = 16
	}
	if keyLen <= 0 {
		keyLen = 32
	}
	return &pbkdf2Hasher{
		iterations: iterations,
		saltLen:    saltLen,
		keyLen:     keyLen,
	}
}

// Hash generates a password hash using PBKDF2 with SHA-256.
// Returns a string in the format: iterations$salt$hash (all base64 encoded)
func (h *pbkdf2Hasher) Hash(password string) (string, error) {
	// Generate random salt
	salt := make([]byte, h.saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key using PBKDF2
	hash := pbkdf2([]byte(password), salt, h.iterations, h.keyLen)

	// Encode as base64 and format as: iterations$salt$hash
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("%d$%s$%s", h.iterations, encodedSalt, encodedHash), nil
}

// Compare compares a hashed password with a plain text password.
// Returns nil if they match, error otherwise.
func (h *pbkdf2Hasher) Compare(hashedPassword, password string) error {
	// Parse the hashed password
	parts := strings.Split(hashedPassword, "$")
	if len(parts) != 3 {
		return fmt.Errorf("invalid hash format")
	}

	// Extract iterations, salt, and hash
	var iterations int
	if _, err := fmt.Sscanf(parts[0], "%d", &iterations); err != nil {
		return fmt.Errorf("invalid iterations: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("invalid salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return fmt.Errorf("invalid hash: %w", err)
	}

	// Derive key from provided password
	actualHash := pbkdf2([]byte(password), salt, iterations, len(expectedHash))

	// Constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare(expectedHash, actualHash) != 1 {
		return fmt.Errorf("password does not match")
	}

	return nil
}

// pbkdf2 implements PBKDF2 key derivation using HMAC-SHA256.
// This is a standard library implementation without external dependencies.
func pbkdf2(password, salt []byte, iterations, keyLen int) []byte {
	hashLen := sha256.Size
	numBlocks := (keyLen + hashLen - 1) / hashLen

	var dk []byte
	for block := 1; block <= numBlocks; block++ {
		dk = append(dk, pbkdf2Block(password, salt, iterations, block)...)
	}

	return dk[:keyLen]
}

// pbkdf2Block computes one block of PBKDF2
func pbkdf2Block(password, salt []byte, iterations, blockNum int) []byte {
	// U1 = PRF(password, salt || INT_32_BE(i))
	h := hmacSHA256(password, append(salt, byte(blockNum>>24), byte(blockNum>>16), byte(blockNum>>8), byte(blockNum)))
	result := make([]byte, len(h))
	copy(result, h)

	// U2 through Uc
	for i := 2; i <= iterations; i++ {
		h = hmacSHA256(password, h)
		for j := range result {
			result[j] ^= h[j]
		}
	}

	return result
}

// hmacSHA256 computes HMAC-SHA256
func hmacSHA256(key, data []byte) []byte {
	blockSize := 64 // SHA-256 block size

	// Keys longer than blockSize are shortened by hashing them
	if len(key) > blockSize {
		h := sha256.Sum256(key)
		key = h[:]
	}

	// Keys shorter than blockSize are padded with zeros
	if len(key) < blockSize {
		padded := make([]byte, blockSize)
		copy(padded, key)
		key = padded
	}

	// Compute inner and outer padded keys
	ipad := make([]byte, blockSize)
	opad := make([]byte, blockSize)
	for i := 0; i < blockSize; i++ {
		ipad[i] = key[i] ^ 0x36
		opad[i] = key[i] ^ 0x5c
	}

	// Compute inner hash
	innerHash := sha256.New()
	innerHash.Write(ipad)
	innerHash.Write(data)
	inner := innerHash.Sum(nil)

	// Compute outer hash
	outerHash := sha256.New()
	outerHash.Write(opad)
	outerHash.Write(inner)

	return outerHash.Sum(nil)
}
