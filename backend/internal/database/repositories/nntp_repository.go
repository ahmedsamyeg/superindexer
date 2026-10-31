package repositories

import (
	"context"
	"fmt"

	"github.com/superindexer/indexer/internal/database/models"
	"github.com/uptrace/bun"
)

// NNTPServerRepository defines the interface for NNTP server data access.
type NNTPServerRepository interface {
	Create(ctx context.Context, server *models.NNTPServer) error
	GetByID(ctx context.Context, id int64) (*models.NNTPServer, error)
	GetAll(ctx context.Context) ([]*models.NNTPServer, error)
	GetEnabled(ctx context.Context) ([]*models.NNTPServer, error)
	Update(ctx context.Context, server *models.NNTPServer) error
	Delete(ctx context.Context, id int64) error
	UpdateStatus(ctx context.Context, id int64, status string) error
}

// nntpServerRepository implements NNTPServerRepository.
type nntpServerRepository struct {
	db *bun.DB
}

// NewNNTPServerRepository creates a new NNTPServerRepository.
func NewNNTPServerRepository(db *bun.DB) NNTPServerRepository {
	return &nntpServerRepository{db: db}
}

func (r *nntpServerRepository) Create(ctx context.Context, server *models.NNTPServer) error {
	_, err := r.db.NewInsert().Model(server).Exec(ctx)
	return err
}

func (r *nntpServerRepository) GetByID(ctx context.Context, id int64) (*models.NNTPServer, error) {
	server := &models.NNTPServer{}
	err := r.db.NewSelect().Model(server).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return server, nil
}

func (r *nntpServerRepository) GetAll(ctx context.Context) ([]*models.NNTPServer, error) {
	var servers []*models.NNTPServer
	err := r.db.NewSelect().Model(&servers).Order("priority DESC, created_at DESC").Scan(ctx)
	return servers, err
}

func (r *nntpServerRepository) GetEnabled(ctx context.Context) ([]*models.NNTPServer, error) {
	var servers []*models.NNTPServer
	err := r.db.NewSelect().Model(&servers).
		Where("enabled = true").
		Order("priority DESC, created_at DESC").
		Scan(ctx)
	return servers, err
}

func (r *nntpServerRepository) Update(ctx context.Context, server *models.NNTPServer) error {
	_, err := r.db.NewUpdate().Model(server).Where("id = ?", server.ID).Exec(ctx)
	return err
}

func (r *nntpServerRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.NewDelete().Model(&models.NNTPServer{}).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("NNTP server not found: %d", id)
	}
	return nil
}

func (r *nntpServerRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.NewUpdate().
		Model(&models.NNTPServer{}).
		Set("connection_status = ?", status).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// Ensure interface compliance
var _ NNTPServerRepository = (*nntpServerRepository)(nil)
