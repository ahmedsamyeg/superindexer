package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/superindexer/indexer/internal/auth"
	"github.com/superindexer/indexer/internal/database/repositories"
	"github.com/uptrace/bun"
)

// RefreshTokenService handles refresh token operations.
type RefreshTokenService struct {
	db            *bun.DB
	userRepo      repositories.UserRepository
	jwtSecret     string
	jwtIssuer     string
	jwtExpiry     int64
	refreshExpiry time.Duration
}

// NewRefreshTokenService creates a new RefreshTokenService.
func NewRefreshTokenService(db *bun.DB, userRepo repositories.UserRepository, jwtSecret string, jwtExpiry time.Duration) *RefreshTokenService {
	return &RefreshTokenService{
		db:            db,
		userRepo:      userRepo,
		jwtSecret:     jwtSecret,
		jwtIssuer:     "superindexer",
		jwtExpiry:     int64(jwtExpiry.Seconds()),
		refreshExpiry: 7 * 24 * time.Hour, // 7 days
	}
}

// RefreshTokenPair holds both the access token and refresh token.
type RefreshTokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

// GenerateRefreshToken creates a new refresh token for the user.
func (s *RefreshTokenService) GenerateRefreshToken(ctx context.Context, userID int64, role, username string) (*RefreshTokenPair, error) {
	// Generate access token
	accessToken, err := auth.GenerateToken(userID, role, username, auth.JWTConfig{
		Secret: []byte(s.jwtSecret),
		Issuer: s.jwtIssuer,
		Expiry: s.jwtExpiry,
	})
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.createRefreshToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("create refresh token: %w", err)
	}

	return &RefreshTokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(s.jwtExpiry) * time.Second).Unix(),
	}, nil
}

// createRefreshToken creates and stores a refresh token in the database.
func (s *RefreshTokenService) createRefreshToken(ctx context.Context, userID int64) (string, error) {
	// Generate a secure random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Store in database
	_, err := s.db.NewInsert().
		Model(&RefreshToken{
			UserID:    userID,
			Token:     token,
			ExpiresAt: time.Now().Add(s.refreshExpiry),
			IsRevoked: false,
		}).
		Exec(ctx)
	if err != nil {
		return "", fmt.Errorf("store refresh token: %w", err)
	}

	return token, nil
}

// RefreshAccessKey validates a refresh token and returns a new access token.
func (s *RefreshTokenService) RefreshAccessKey(ctx context.Context, refreshToken string) (*RefreshTokenPair, error) {
	// Find the refresh token in the database
	var tokenRecord RefreshToken
	err := s.db.NewSelect().
		Model(&tokenRecord).
		Where("token = ?", refreshToken).
		Where("is_revoked = false").
		Where("expires_at > ?", time.Now()).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Get user info
	user, err := s.userRepo.GetByID(ctx, tokenRecord.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("account is deactivated")
	}

	// Revoke the old refresh token (use it once)
	_, err = s.db.NewUpdate().
		Model(&RefreshToken{}).
		Set("is_revoked = true").
		Where("id = ?", tokenRecord.ID).
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("revoke old refresh token: %w", err)
	}

	// Generate new token pair
	return s.GenerateRefreshToken(ctx, user.ID, user.Role, user.Username)
}

// RevokeRefreshToken revokes a specific refresh token.
func (s *RefreshTokenService) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	_, err := s.db.NewUpdate().
		Model(&RefreshToken{}).
		Set("is_revoked = true").
		Where("token = ?", refreshToken).
		Where("is_revoked = false").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

// RevokeAllUserTokens revokes all refresh tokens for a user (e.g., on password change).
func (s *RefreshTokenService) RevokeAllUserTokens(ctx context.Context, userID int64) error {
	_, err := s.db.NewUpdate().
		Model(&RefreshToken{}).
		Set("is_revoked = true").
		Where("user_id = ?", userID).
		Where("is_revoked = false").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("revoke all user tokens: %w", err)
	}
	return nil
}

// CleanupExpiredTokens removes expired refresh tokens.
func (s *RefreshTokenService) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	result, err := s.db.NewDelete().
		Model(&RefreshToken{}).
		Where("expires_at < ?", time.Now().Add(-24*time.Hour)).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("cleanup expired refresh tokens: %w", err)
	}
	rows, _ := result.RowsAffected()
	return rows, nil
}

// RefreshToken represents a stored refresh token in the database.
type RefreshToken struct {
	ID        int64     `bun:"id,pk,autoincrement" json:"id"`
	UserID    int64     `bun:"user_id,notnull" json:"user_id"`
	Token     string    `bun:"token,notnull" json:"-"`
	ExpiresAt time.Time `bun:"expires_at,notnull" json:"expires_at"`
	IsRevoked bool      `bun:"is_revoked,notnull,default:false" json:"is_revoked"`
	CreatedAt time.Time `bun:"created_at,notnull,default:now()" json:"created_at"`
}

// TableName returns the table name for RefreshToken.
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
