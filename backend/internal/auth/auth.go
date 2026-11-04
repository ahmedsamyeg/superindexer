package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
)

// ContextKey is the key used to store auth info in request context.
type ContextKey string

const (
	ContextKeyUserID   ContextKey = "user_id"
	ContextKeyUserRole ContextKey = "user_role"
	ContextKeyUsername ContextKey = "username"
)

// PasswordConfig holds the Argon2id parameters.
type PasswordConfig struct {
	Memory      uint32
	Iterations  uint32
	SaltLength  uint32
	KeyLength   uint32
	Parallelism uint8
}

// DefaultPasswordConfig uses recommended Argon2id parameters.
var DefaultPasswordConfig = PasswordConfig{
	Memory:      64 * 1024, // 64 MB
	Iterations:  3,
	SaltLength:  16,
	KeyLength:   32,
	Parallelism: 4,
}

// JWTConfig holds JWT signing parameters.
type JWTConfig struct {
	Secret     []byte
	Issuer     string
	Expiry     int64 // seconds
}

// TokenClaims holds the JWT claims.
type TokenClaims struct {
	UserID   int64  `json:"uid"`
	Role     string `json:"rol"`
	Username string `json:"usr"`
	jwt.RegisteredClaims
}

// GeneratePasswordHash hashes a password using Argon2id with a random salt.
// Returns the base64-encoded hash string in the format: base64(salt)||base64(hash).
func GeneratePasswordHash(password string, cfg PasswordConfig) (string, error) {
	salt := make([]byte, cfg.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, cfg.Iterations, cfg.Memory, cfg.Parallelism, cfg.KeyLength)

	return base64.RawStdEncoding.EncodeToString(salt) + "||" + base64.RawStdEncoding.EncodeToString(hash), nil
}

// VerifyPassword verifies a password against a base64-encoded Argon2id hash.
// The hash format is: base64(salt)||base64(hash).
func VerifyPassword(password, hash string, cfg PasswordConfig) error {
	parts := splitHash(hash)
	if len(parts) != 2 {
		return errors.New("invalid hash format")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return fmt.Errorf("decode salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("decode hash: %w", err)
	}

	hashedKey := argon2.IDKey([]byte(password), salt, cfg.Iterations, cfg.Memory, cfg.Parallelism, cfg.KeyLength)

	if !compareHashes(hashedKey, expectedHash) {
		return errors.New("invalid password")
	}

	return nil
}

// GenerateToken creates a JWT token for the given user.
func GenerateToken(userID int64, role, username string, cfg JWTConfig) (string, error) {
	claims := TokenClaims{
		UserID:   userID,
		Role:     role,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.Expiry) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(cfg.Secret)
}

// ValidateToken parses and validates a JWT token, returning the claims.
func ValidateToken(tokenString string, cfg JWTConfig) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return cfg.Secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// ExtractTokenFromHeader extracts the Bearer token from the Authorization header.
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}

	const prefix = "Bearer "
	if len(authHeader) < len(prefix) || authHeader[:len(prefix)] != prefix {
		return "", errors.New("invalid authorization header format")
	}

	return authHeader[len(prefix):], nil
}

// ExtractUserFromContext extracts user info from a request context.
func ExtractUserFromContext(ctx context.Context) (int64, string, string, bool) {
	userID, ok := ctx.Value(ContextKeyUserID).(int64)
	if !ok || userID == 0 {
		return 0, "", "", false
	}

	role, _ := ctx.Value(ContextKeyUserRole).(string)
	username, _ := ctx.Value(ContextKeyUsername).(string)

	return userID, role, username, true
}

// Helper functions

func splitHash(hash string) []string {
	idx := strings.Index(hash, "||")
	if idx == -1 {
		return []string{hash}
	}
	return []string{hash[:idx], hash[idx+2:]}
}

func compareHashes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	return subtleConstantTimeEqual(a, b) == 1
}

func subtleConstantTimeEqual(a, b []byte) int {
	if len(a) != len(b) {
		return 0
	}
	if bytes.Equal(a, b) {
		return 1
	}
	return 0
}
