package metadata

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/superindexer/indexer/internal/database/models"
)

// MetadataProcessor handles processing and normalization of Spotnet metadata
type MetadataProcessor struct {
}

// NewMetadataProcessor creates a new metadata processor
func NewMetadataProcessor() *MetadataProcessor {
	return &MetadataProcessor{}
}

// ProcessPostMetadata normalizes and processes Spotnet post metadata
func (mp *MetadataProcessor) ProcessPostMetadata(ctx context.Context, post *models.SpotnetPost) error {
	// Normalize title
	if post.Title != "" {
		post.Title = mp.normalizeTitle(post.Title)
	}

	// Normalize description
	if post.Description != "" {
		post.Description = mp.normalizeDescription(post.Description)
	}

	// Normalize author
	if post.Author != "" {
		post.Author = mp.normalizeAuthor(post.Author)
	}

	// Normalize category
	if post.CategoryName != "" {
		post.CategoryName = mp.normalizeCategory(post.CategoryName)
	}

	// Normalize date/time
	if post.PostedAt.IsZero() {
		post.PostedAt = time.Now()
	}

	// Create search vector for PostgreSQL FTS (simplified)
	if post.Title != "" || post.Description != "" {
		searchVector := mp.createSearchVector(post)
		post.SearchVector = searchVector
	}

	return nil
}

// normalizeTitle normalizes the title field
func (mp *MetadataProcessor) normalizeTitle(title string) string {
	// Trim whitespace and normalize
	title = strings.TrimSpace(title)
	
	// Remove extra spaces (but preserve meaningful ones)
	var normalized []rune
	prevSpace := false
	
	for _, char := range title {
		if char == ' ' {
			if !prevSpace {
				normalized = append(normalized, char)
				prevSpace = true
			}
		} else {
			normalized = append(normalized, char)
			prevSpace = false
		}
	}
	
	return string(normalized)
}

// normalizeDescription normalizes the description field
func (mp *MetadataProcessor) normalizeDescription(desc string) string {
	// Basic normalization - trim and clean
	desc = strings.TrimSpace(desc)
	
	// Remove extra whitespace but preserve meaningful line breaks for descriptions
	var normalized []rune
	prevSpace := false
	
	for i, char := range desc {
		if char == ' ' || char == '\t' {
			if !prevSpace {
				normalized = append(normalized, ' ')
				prevSpace = true
			}
		} else {
			normalized = append(normalized, char)
			prevSpace = false
		}
	}
	
	return string(normalized)
}

// normalizeAuthor normalizes the author field
func (mp *MetadataProcessor) normalizeAuthor(author string) string {
	// Trim and clean author name
	author = strings.TrimSpace(author)
	
	// Remove extra spaces but preserve names with multiple parts
	author = strings.Join(strings.Fields(author), " ")
	
	return author
}

// normalizeCategory normalizes the category field
func (mp *MetadataProcessor) normalizeCategory(category string) string {
	// Basic category normalization
	category = strings.TrimSpace(category)
	
	// Convert to title case or lowercase as appropriate
	category = strings.ToLower(category)
	
	// Capitalize first letter of each word (if needed)
	words := strings.Fields(category)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.Title(word)
		}
	}
	
	return strings.Join(words, " ")
}

// createSearchVector creates a simplified search vector for PostgreSQL FTS
func (mp *MetadataProcessor) createSearchVector(post *models.SpotnetPost) string {
	// Simple vector creation for PostgreSQL search
	var parts []string
	
	if post.Title != "" {
		parts = append(parts, fmt.Sprintf("'%s'", strings.ReplaceAll(post.Title, "'", "''")))
	}
	
	if post.Description != "" {
		parts = append(parts, fmt.Sprintf("'%s'", strings.ReplaceAll(post.Description, "'", "''")))
	}
	
	if post.CategoryName != "" {
		parts = append(parts, fmt.Sprintf("'%s'", strings.ReplaceAll(post.CategoryName, "'", "''")))
	}
	
	if len(parts) == 0 {
		return ""
	}
	
	// For simplicity, just return a basic string - actual PostgreSQL FTS would be handled in database queries
	return strings.Join(parts, " ")
}

// ValidatePost validates that a post has all required fields
func (mp *MetadataProcessor) ValidatePost(ctx context.Context, post *models.SpotnetPost) error {
	if post.ExternalID == "" {
		return fmt.Errorf("external ID is required")
	}
	
	if post.Title == "" {
		return fmt.Errorf("title is required")
	}
	
	if post.SourceID == 0 {
		return fmt.Errorf("source ID is required")
	}
	
	return nil
}

// ApplyCategoryMapping applies category mapping rules to posts
func (mp *MetadataProcessor) ApplyCategoryMapping(ctx context.Context, post *models.SpotnetPost) error {
	// In a real implementation, this would:
	// 1. Check configured category mappings
	// 2. Apply automatic category assignments based on content
	// 3. Support manual overrides and aliases
	
	// For now, just ensure the category is set if not already
	if post.CategoryName == "" {
		post.CategoryName = "General"
	}
	
	return nil
}