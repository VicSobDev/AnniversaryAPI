package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2Config is a configuration for Argon2
type Argon2Config struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
}

// Argon2 is a wrapper around the argon2id algorithm
type Argon2 struct {
	config Argon2Config
}

// NewArgon2 creates a new Argon2 instance with the specified configuration
func NewArgon2(config Argon2Config) *Argon2 {

	return &Argon2{config: config}
}

// Hash hashes the specified password using the Argon2 algorithm
func (a *Argon2) Hash(password string) (string, error) {

	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, a.config.Time, a.config.Memory, a.config.Threads, a.config.KeyLen)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, a.config.Memory, a.config.Time, a.config.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash)), nil
}

// Verify compares the specified password with the specified hash
func (a *Argon2) Verify(hash, password string) (bool, error) {

	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}

	var version int
	var config Argon2Config
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, fmt.Errorf("failed to parse version: %w", err)
	}

	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &config.Memory, &config.Time, &config.Threads)
	if err != nil {
		return false, fmt.Errorf("failed to parse hash: %w", err)
	}

	// Extract the salt part and decode it
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	// Determine the key length from the hash part of the string, assuming base64 encoding
	config.KeyLen = uint32(len(parts[5]) * 3 / 4) // Approximation of base64 decoding size

	// Generate a new hash using the password and the decoded salt
	newHash := argon2.IDKey([]byte(password), salt, config.Time, config.Memory, config.Threads, config.KeyLen)

	// Compare the generated hash with the hash part of the original string
	return base64.RawStdEncoding.EncodeToString(newHash) == parts[5], nil
}
