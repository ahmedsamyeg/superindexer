package service

import (
	"context"
	"fmt"
	"time"

	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/database/repositories"
	"github.com/superindexer/indexer/internal/spotnet/client"
	"github.com/superindexer/indexer/internal/spotnet/metadata"
	"github.com/superindexer/indexer/internal/spotnet/parser"
)

// SpotnetService handles Spotnet source and post operations
type SpotnetService struct {
	sourceRepo repositories.SpotnetSourceRepository
	postRepo   repositories.SpotnetPostRepository
	client     *client.SpotnetClient
	parser     *parser.PostParser
	metadata   *metadata.MetadataProcessor
}

// NewSpotnetService creates a new Spotnet service
func NewSpotnetService(
	sourceRepo repositories.SpotnetSourceRepository,
	postRepo repositories.SpotnetPostRepository,
	client *client.SpotnetClient,
	parser *parser.PostParser,
	metadata *metadata.MetadataProcessor,
) *SpotnetService {
	return &SpotnetService{
		sourceRepo: sourceRepo,
		postRepo:   postRepo,
		client:     client,
		parser:     parser,
		metadata:   metadata,
	}
}

// CreateSource creates a new Spotnet source
func (s *SpotnetService) CreateSource(ctx context.Context, name, sourceType, config string) (*models.SpotnetSource, error) {
	source := &models.SpotnetSource{
		Name:          name,
		SourceType:    sourceType,
		Configuration: config,
		Enabled:       true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := s.sourceRepo.Create(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("failed to create source: %w", err)
	}

	return source, nil
}

// GetSource gets a Spotnet source by ID
func (s *SpotnetService) GetSource(ctx context.Context, id int64) (*models.SpotnetSource, error) {
	source, err := s.sourceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get source: %w", err)
	}

	return source, nil
}

// ListSources lists all Spotnet sources
func (s *SpotnetService) ListSources(ctx context.Context) ([]*models.SpotnetSource, error) {
	sources, err := s.sourceRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list sources: %w", err)
	}

	return sources, nil
}

// UpdateSource updates a Spotnet source
func (s *SpotnetService) UpdateSource(ctx context.Context, id int64, name, sourceType, config string) (*models.SpotnetSource, error) {
	source, err := s.sourceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get source: %w", err)
	}

	source.Name = name
	source.SourceType = sourceType
	source.Configuration = config
	source.UpdatedAt = time.Now()

	err = s.sourceRepo.Update(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("failed to update source: %w", err)
	}

	return source, nil
}

// DeleteSource deletes a Spotnet source
func (s *SpotnetService) DeleteSource(ctx context.Context, id int64) error {
	err := s.sourceRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete source: %w", err)
	}

	return nil
}

// TestSourceConnection tests if we can connect to a Spotnet source
func (s *SpotnetService) TestSourceConnection(ctx context.Context, source *models.SpotnetSource) error {
	return s.client.TestConnection(ctx, source)
}

// SyncSource synchronizes posts from a Spotnet source
func (s *SpotnetService) SyncSource(ctx context.Context, source *models.SpotnetSource) (int64, error) {
	// Get posts from the source
	posts, err := s.client.GetPosts(ctx, source)
	if err != nil {
		return 0, fmt.Errorf("failed to get posts from source: %w", err)
	}

	// Parse and process posts
	parsedPosts := make([]*models.SpotnetPost, 0, len(posts))
	
	for _, post := range posts {
		parsedPost, err := s.parser.ParseSpotnetPost(ctx, post, source)
		if err != nil {
			// Log error but continue with other posts
			continue
		}
		
		// Process metadata for the post
		err = s.metadata.ProcessPostMetadata(ctx, parsedPost)
		if err != nil {
			// Log error but continue with other posts
			continue
		}
		
		parsedPosts = append(parsedPosts, parsedPost)
	}
	
	// Create posts in database (bulk insert)
	if len(parsedPosts) > 0 {
		err = s.postRepo.CreateBulk(ctx, parsedPosts)
		if err != nil {
			return 0, fmt.Errorf("failed to create posts: %w", err)
		}
	}
	
	// Update last sync time
	now := time.Now()
	err = s.sourceRepo.UpdateLastSync(ctx, source.ID, now)
	if err != nil {
		return 0, fmt.Errorf("failed to update last sync: %w", err)
	}
	
	return int64(len(parsedPosts)), nil
}

// ValidateSourceConfig validates a Spotnet source configuration
func (s *SpotnetService) ValidateSourceConfig(ctx context.Context, config *client.SpotnetSourceConfig) error {
	return s.client.ValidateSourceConfig(config)
}