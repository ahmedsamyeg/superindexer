package groups

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/database/repositories"
	"github.com/superindexer/indexer/internal/usenet/connection"
	"github.com/superindexer/indexer/internal/usenet/client"
)

// NewsgroupManager handles newsgroup discovery and management
type NewsgroupManager struct {
	db      repositories.NewsgroupRepository
	connMgr *connection.ConnectionManager
}

// NewNewsgroupManager creates a new newsgroup manager
func NewNewsgroupManager(db repositories.NewsgroupRepository, connMgr *connection.ConnectionManager) *NewsgroupManager {
	return &NewsgroupManager{
		db:      db,
		connMgr: connMgr,
	}
}

// DiscoverNewsgroups lists all available newsgroups on an NNTP server
func (nm *NewsgroupManager) DiscoverNewsgroups(ctx context.Context, server *models.NNTPServer) ([]*models.Newsgroup, error) {
	// Get a connection to the server
	conn, err := nm.connMgr.GetConnection(ctx, server)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection to server %s: %w", server.Hostname, err)
	}
	defer nm.connMgr.ReleaseConnection(server.ID, conn)

	// Use the NNTP client's List method to get newsgroups
	newsgroupInfos, err := conn.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list newsgroups: %w", err)
	}

	newsgroups := make([]*models.Newsgroup, 0, len(newsgroupInfos))
	for _, ni := range newsgroupInfos {
		newsgroup := &models.Newsgroup{
			Name:                   ni.Name,
			Description:            ni.Description,
			Enabled:                true,
			FirstArticleNumber:     ni.FirstArticle,
			LastArticleNumber:      ni.LastArticle,
			LastIndexedArticleNum:  nil,
			LastIndexedAt:          nil,
			LastCheckAt:            nil,
			IndexingPriority:       0,
			CreatedAt:              time.Now(),
			UpdatedAt:              time.Now(),
		}
		newsgroups = append(newsgroups, newsgroup)
	}

	return newsgroups, nil
}

// GetNewsgroupInfo retrieves detailed information about a specific newsgroup
func (nm *NewsgroupManager) GetNewsgroupInfo(ctx context.Context, server *models.NNTPServer, newsgroupName string) (*models.Newsgroup, error) {
	// Get a connection to the server
	conn, err := nm.connMgr.GetConnection(ctx, server)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection to server %s: %w", server.Hostname, err)
	}
	defer nm.connMgr.ReleaseConnection(server.ID, conn)

	// Use GROUP command to get newsgroup info
	status, response, _, err := conn.SendCommand(ctx, "GROUP", newsgroupName)
	if err != nil {
		return nil, fmt.Errorf("failed to send GROUP command: %w", err)
	}

	if status >= 400 {
		return nil, fmt.Errorf("server GROUP command failed: %s", response)
	}

	// Parse the response to get article numbers
	// Format: "211 <newsgroup_name> article_count first_article last_article"
	parts := strings.Fields(response)
	if len(parts) < 4 {
		return nil, fmt.Errorf("unexpected GROUP response format: %s", response)
	}

	// Parse the article numbers
	firstArticle, err := parseArticleNumber(parts[2])
	if err != nil {
		return nil, fmt.Errorf("failed to parse first article number: %w", err)
	}

	lastArticle, err := parseArticleNumber(parts[3])
	if err != nil {
		return nil, fmt.Errorf("failed to parse last article number: %w", err)
	}

	newsgroup := &models.Newsgroup{
		Name:                   newsgroupName,
		Description:            "",
		Enabled:                true,
		FirstArticleNumber:     firstArticle,
		LastArticleNumber:      lastArticle,
		LastIndexedArticleNum:  nil,
		LastIndexedAt:          nil,
		LastCheckAt:            nil,
		IndexingPriority:       0,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	return newsgroup, nil
}

// GetNewsgroupActiveInfo gets newsgroup info using LIST ACTIVE command
func (nm *NewsgroupManager) GetNewsgroupActiveInfo(ctx context.Context, server *models.NNTPServer) ([]*models.Newsgroup, error) {
	conn, err := nm.connMgr.GetConnection(ctx, server)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection to server %s: %w", server.Hostname, err)
	}
	defer nm.connMgr.ReleaseConnection(server.ID, conn)

	newsgroupInfos, err := conn.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list active newsgroups: %w", err)
	}

	newsgroups := make([]*models.Newsgroup, 0, len(newsgroupInfos))
	for _, ni := range newsgroupInfos {
		newsgroup := &models.Newsgroup{
			Name:                   ni.Name,
			Description:            ni.Description,
			Enabled:                true,
			FirstArticleNumber:     ni.FirstArticle,
			LastArticleNumber:      ni.LastArticle,
			LastIndexedArticleNum:  nil,
			LastIndexedAt:          nil,
			LastCheckAt:            nil,
			IndexingPriority:       0,
			CreatedAt:              time.Now(),
			UpdatedAt:              time.Now(),
		}
		newsgroups = append(newsgroups, newsgroup)
	}

	return newsgroups, nil
}

// GetNewsgroupArticleRange gets the article range for a specific newsgroup using STAT command
func (nm *NewsgroupManager) GetNewsgroupArticleRange(ctx context.Context, server *models.NNTPServer, newsgroupName string) (int64, int64, error) {
	conn, err := nm.connMgr.GetConnection(ctx, server)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get connection to server %s: %w", server.Hostname, err)
	}
	defer nm.connMgr.ReleaseConnection(server.ID, conn)

	// Use GROUP command to get article range
	status, response, _, err := conn.SendCommand(ctx, "GROUP", newsgroupName)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to send GROUP command: %w", err)
	}

	if status >= 400 {
		return 0, 0, fmt.Errorf("server GROUP command failed: %s", response)
	}

	parts := strings.Fields(response)
	if len(parts) < 4 {
		return 0, 0, fmt.Errorf("unexpected GROUP response format: %s", response)
	}

	firstArticle, err := parseArticleNumber(parts[2])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse first article number: %w", err)
	}

	lastArticle, err := parseArticleNumber(parts[3])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse last article number: %w", err)
	}

	return firstArticle, lastArticle, nil
}

// parseArticleCount converts article count string to integer
func parseArticleCount(countStr string) (int64, error) {
	count, err := parseArticleNumber(countStr)
	if err != nil {
		return 0, fmt.Errorf("invalid article count: %w", err)
	}
	return count, nil
}

// parseArticleNumber converts article number string to integer
func parseArticleNumber(numStr string) (int64, error) {
	numStr = strings.Trim(numStr, " \t")
	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse article number: %w", err)
	}
	return num, nil
}

// ConvertNewsgroupInfoToModel converts NNTP client NewsgroupInfo to database model
func ConvertNewsgroupInfoToModel(ni client.NewsgroupInfo) *models.Newsgroup {
	return &models.Newsgroup{
		Name:                   ni.Name,
		Description:            ni.Description,
		Enabled:                true,
		FirstArticleNumber:     ni.FirstArticle,
		LastArticleNumber:      ni.LastArticle,
		LastIndexedArticleNum:  nil,
		LastIndexedAt:          nil,
		LastCheckAt:            nil,
		IndexingPriority:       0,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}
}
