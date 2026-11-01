package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/superindexer/indexer/internal/database/models"
)

// PostParser handles parsing Spotnet posts into standardized format
type PostParser struct {
}

// NewPostParser creates a new post parser
func NewPostParser() *PostParser {
	return &PostParser{}
}

// ParseSpotnetPost converts raw Spotnet data to internal model
func (p *PostParser) ParseSpotnetPost(ctx context.Context, rawPost interface{}, source *models.SpotnetSource) (*models.SpotnetPost, error) {
	// Convert raw data to a standardized format
	var post *models.SpotnetPost
	
	switch v := rawPost.(type) {
	case *models.SpotnetPost:
		return v, nil
	case map[string]interface{}:
		// Parse from map structure
		post = &models.SpotnetPost{
			ExternalID:         getStringValue(v, "id"),
			Title:              getStringValue(v, "title"),
			Description:        getStringValue(v, "description"),
			Author:             getStringValue(v, "author"),
			CategoryName:       getStringValue(v, "category"),
			PostedAt:           parseTime(getStringValue(v, "posted_at")),
			RawMetadata:        string(marshalJSON(v)),
			NormalizedMetadata: "{}",
		}
		
		// Set source ID
		if source != nil {
			post.SourceID = source.ID
		}
		
	default:
		return nil, fmt.Errorf("unsupported post format: %T", v)
	}
	
	return post, nil
}

// ParseSpotnetPosts converts multiple raw posts to internal models
func (p *PostParser) ParseSpotnetPosts(ctx context.Context, rawPosts []interface{}, source *models.SpotnetSource) ([]*models.SpotnetPost, error) {
	posts := make([]*models.SpotnetPost, 0, len(rawPosts))
	
	for _, rawPost := range rawPosts {
		post, err := p.ParseSpotnetPost(ctx, rawPost, source)
		if err != nil {
			return nil, fmt.Errorf("failed to parse post: %w", err)
		}
		posts = append(posts, post)
	}
	
	return posts, nil
}

// NormalizeMetadata normalizes post metadata for consistent storage and search
func (p *PostParser) NormalizeMetadata(ctx context.Context, post *models.SpotnetPost) error {
	// Apply normalization rules
	if post.Title != "" {
		// Normalize title (trim whitespace, etc.)
		post.Title = normalizeString(post.Title)
	}
	
	if post.Description != "" {
		// Normalize description
		post.Description = normalizeString(post.Description)
	}
	
	if post.Author != "" {
		// Normalize author name
		post.Author = normalizeString(post.Author)
	}
	
	if post.CategoryName != "" {
		// Normalize category name
		post.CategoryName = normalizeString(post.CategoryName)
	}
	
	// Create normalized metadata structure
	normalized := map[string]interface{}{
		"title":       post.Title,
		"description": post.Description,
		"author":      post.Author,
		"category":    post.CategoryName,
		"posted_at":   post.PostedAt,
	}
	
	// Convert to JSON string
	normalizedJSON, err := json.Marshal(normalized)
	if err != nil {
		return fmt.Errorf("failed to marshal normalized metadata: %w", err)
	}
	
	post.NormalizedMetadata = string(normalizedJSON)
	return nil
}

// Helper functions

func getStringValue(data map[string]interface{}, key string) string {
	if val, exists := data[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func parseTime(timeStr string) time.Time {
	if timeStr == "" {
		return time.Now()
	}
	
	// Try parsing different time formats
	t, err := time.Parse(time.RFC3339, timeStr)
	if err == nil {
		return t
	}
	
	t, err = time.Parse("2006-01-02T15:04:05Z", timeStr)
	if err == nil {
		return t
	}
	
	t, err = time.Parse(time.RFC1123Z, timeStr)
	if err == nil {
		return t
	}
	
	return time.Now()
}

func normalizeString(s string) string {
	// Simple normalization - trim whitespace and remove extra spaces
	s = trimWhitespace(s)
	return s
}

func trimWhitespace(s string) string {
	// Remove leading/trailing whitespace and normalize internal spaces
	result := ""
	inSpace := false
	
	for i, char := range s {
		if char == ' ' || char == '\t' || char == '\n' {
			if !inSpace && i > 0 {
				result += " "
				inSpace = true
			}
		} else {
			result += string(char)
			inSpace = false
		}
	}
	
	return result
}

func marshalJSON(v interface{}) []byte {
	if data, ok := v.([]byte); ok {
		return data
	}
	
	data, _ := json.Marshal(v)
	return data
}