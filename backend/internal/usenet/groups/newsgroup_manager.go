package groups

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/database/repositories"
	"github.com/superindexer/indexer/internal/usenet/client"
)

// NewsgroupManager handles newsgroup discovery and management
type NewsgroupManager struct {
	db         repositories.UserRepository // Note: this should be the correct repository type
	connMgr    *ConnectionManager         // Will need to update this reference once ConnectionManager is defined
}

// NewNewsgroupManager creates a new newsgroup manager
func NewNewsgroupManager(db repositories.UserRepository, connMgr *ConnectionManager) *NewsgroupManager {
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

	// Send GROUP command to get available newsgroups
	status, response, err := conn.SendCommand(ctx, "LIST")
	if err != nil {
		return nil, fmt.Errorf("failed to send LIST command: %w", err)
	}

	if status >= 400 {
		return nil, fmt.Errorf("server LIST command failed: %s", response)
	}

	// Parse the newsgroup list (format varies by server but typically is "newsgroup_name article_count")
	newsgroups := make([]*models.Newsgroup, 0)
	lines := strings.Split(response, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse format: "newsgroup_name article_count"
		parts := strings.Split(line, " ")
		if len(parts) >= 2 {
			newsgroupName := parts[0]
			
			// Parse article count (we'll use this for initial setup)
			articleCount, err := parseArticleCount(parts[1])
			if err != nil {
				// Skip groups with invalid article count
				continue
			}

			newsgroup := &models.Newsgroup{
				Name:                   newsgroupName,
				Description:            "",
				Enabled:                true,
				FirstArticleNumber:     1,
				LastArticleNumber:      articleCount,
				LastIndexedArticleNum:  nil,
				LastIndexedAt:          nil,
				LastCheckAt:            nil,
				IndexingPriority:       0,
				CreatedAt:              time.Now(),
				UpdatedAt:              time.Now(),
			}
			
			newsgroups = append(newsgroups, newsgroup)
		}
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

	// Use GROUP command to get newsgroup info (this will typically return article range)
	status, response, err := conn.SendCommand(ctx, "GROUP", newsgroupName)
	if err != nil {
		return nil, fmt.Errorf("failed to send GROUP command: %w", err)
	}

	if status >= 400 {
		return nil, fmt.Errorf("server GROUP command failed: %s", response)
	}

	// Parse the response to get article numbers
	// Format: "211 <newsgroup_name> article_count first_article last_article"
	parts := strings.Split(response, " ")
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

// parseArticleCount converts article count string to integer
func parseArticleCount(countStr string) (int64, error) {
	// Convert string to integer
	count, err := parseArticleNumber(countStr)
	if err != nil {
		return 0, fmt.Errorf("invalid article count: %w", err)
	}
	return count, nil
}

// parseArticleNumber converts article number string to integer
func parseArticleNumber(numStr string) (int64, error) {
	// Remove any non-numeric characters and convert to integer
	numStr = strings.Trim(numStr, " \t")
	
	// Convert string to integer
	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse article number: %w", err)
	}
	
	return num, nil
}