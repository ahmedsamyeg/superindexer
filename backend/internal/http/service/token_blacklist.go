package service

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

// TokenBlacklistService handles JWT token revocation.
type TokenBlacklistService struct {
	db *bun.DB
}

// NewTokenBlacklistService creates a new TokenBlacklistService.
func NewTokenBlacklistService(db *bun.DB) *TokenBlacklistService {
	return &TokenBlacklistService{db: db}
}

// TokenBlacklist represents a revoked JWT token.
type TokenBlacklist struct {
	ID        int64     `bun:"id,pk,autoincrement" json:"id"`
	TokenHash string    `bun:"token_hash,notnull" json:"-"`
	UserID    int64     `bun:"user_id,notnull" json:"user_id"`
	ExpiresAt time.Time `bun:"expires_at,notnull" json:"expires_at"`
	CreatedAt time.Time `bun:"created_at,notnull,default:now()" json:"created_at"`
}

// TableName returns the table name for TokenBlacklist.
func (TokenBlacklist) TableName() string {
	return "token_blacklist"
}

// BlacklistToken adds a token to the blacklist.
func (s *TokenBlacklistService) BlacklistToken(ctx context.Context, tokenHash string, userID int64, expiresAt time.Time) error {
	_, err := s.db.NewInsert().Model(&TokenBlacklist{
		TokenHash: tokenHash,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}).Exec(ctx)
	if err != nil {
		return fmt.Errorf("blacklist token: %w", err)
	}
	return nil
}

// IsTokenBlacklisted checks if a token is in the blacklist.
func (s *TokenBlacklistService) IsTokenBlacklisted(ctx context.Context, tokenHash string) (bool, error) {
	var count int
	err := s.db.NewSelect().
		Table("token_blacklist").
		Column("count(*)").
		Where("token_hash = ?", tokenHash).
		Where("expires_at > ?", time.Now()).
		Scan(ctx, &count)
	if err != nil {
		return false, fmt.Errorf("check token blacklist: %w", err)
	}
	return count > 0, nil
}

// CleanupExpiredTokens removes expired tokens from the blacklist.
func (s *TokenBlacklistService) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	result, err := s.db.NewDelete().
		Table("token_blacklist").
		Where("expires_at < ?", time.Now().Add(-24*time.Hour)). // Keep last 24h of expired tokens
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("cleanup expired tokens: %w", err)
	}
	rows, _ := result.RowsAffected()
	return rows, nil
}

// TokenBlacklistRepository defines the interface for token blacklist operations.
// This allows for testing with mocks.
type TokenBlacklistRepository interface {
	BlacklistToken(ctx context.Context, tokenHash string, userID int64, expiresAt time.Time) error
	IsTokenBlacklisted(ctx context.Context, tokenHash string) (bool, error)
	CleanupExpiredTokens(ctx context.Context) (int64, error)
}

// Ensure TokenBlacklistService implements TokenBlacklistRepository.
var _ TokenBlacklistRepository = (*TokenBlacklistService)(nil)
