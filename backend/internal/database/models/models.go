package models

import (
	"time"

	"github.com/uptrace/bun"
)

// BaseModel provides common fields for all models.
type BaseModel struct {
	ID        int64     `bun:"id,pk,autoincrement" json:"id"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// SoftDelete provides soft delete functionality.
type SoftDelete struct {
	DeletedAt time.Time `bun:"deleted_at,soft_delete" json:"deleted_at,omitempty"`
}

// PaginatedResult holds paginated query results.
type PaginatedResult struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}

// NewPaginatedResult creates a new paginated result.
func NewPaginatedResult(data interface{}, total, page, perPage int64) *PaginatedResult {
	totalPages := total / perPage
	if total%perPage > 0 {
		totalPages++
	}

	return &PaginatedResult{
		Data:       data,
		Total:      total,
		Page:       int(page),
		PerPage:    int(perPage),
		TotalPages: int(totalPages),
	}
}

// User represents a user in the system.
type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID                  int64      `bun:"id,pk,autoincrement" json:"id"`
	Username            string     `bun:"username,notnull,unique" json:"username"`
	Email               string     `bun:"email,notnull,unique" json:"email"`
	PasswordHash        string     `bun:"password_hash,notnull" json:"-"`
	Role                string     `bun:"role,notnull,default:'user'" json:"role"`
	IsActive            bool       `bun:"is_active,notnull,default:true" json:"is_active"`
	PasswordResetToken  *string    `bun:"password_reset_token" json:"-"`
	PasswordResetExpiry *time.Time `bun:"password_reset_expiry" json:"-"`
	CreatedAt           time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt           time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// NNTPServer represents a configured NNTP server.
type NNTPServer struct {
	bun.BaseModel `bun:"table:nntp_servers,alias:s"`

	ID               int64     `bun:"id,pk,autoincrement" json:"id"`
	Name             string    `bun:"name,notnull" json:"name"`
	Hostname         string    `bun:"hostname,notnull" json:"hostname"`
	Port             int       `bun:"port,notnull" json:"port"`
	Username         string    `bun:"username" json:"-"`
	EncryptedPassword string   `bun:"encrypted_password" json:"-"`
	TLSEnabled       bool      `bun:"tls_enabled,notnull,default:false" json:"tls_enabled"`
	MaxConnections   int       `bun:"max_connections,notnull,default:10" json:"max_connections"`
	Priority         int       `bun:"priority,notnull,default:0" json:"priority"`
	Enabled          bool      `bun:"enabled,notnull,default:true" json:"enabled"`
	LastConnectedAt  *time.Time `bun:"last_connected_at" json:"last_connected_at,omitempty"`
	ConnectionStatus string    `bun:"connection_status,notnull,default:'unknown'" json:"connection_status"`
	Capabilities     string    `bun:"capabilities" json:"capabilities,omitempty"`
	CreatedAt        time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt        time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// Newsgroup represents a Usenet newsgroup.
type Newsgroup struct {
	bun.BaseModel `bun:"table:newsgroups,alias:g"`

	ID                     int64      `bun:"id,pk,autoincrement" json:"id"`
	Name                   string     `bun:"name,notnull,unique" json:"name"`
	Description            string     `bun:"description" json:"description"`
	Enabled                bool       `bun:"enabled,notnull,default:true" json:"enabled"`
	FirstArticleNumber     int64      `bun:"first_article_number" json:"first_article_number"`
	LastArticleNumber      int64      `bun:"last_article_number" json:"last_article_number"`
	LastIndexedArticleNum  *int64     `bun:"last_indexed_article_number" json:"last_indexed_article_number,omitempty"`
	LastIndexedAt          *time.Time `bun:"last_indexed_at" json:"last_indexed_at,omitempty"`
	LastCheckAt            *time.Time `bun:"last_check_at" json:"last_check_at,omitempty"`
	IndexingPriority       int        `bun:"indexing_priority,notnull,default:0" json:"indexing_priority"`
	CreatedAt              time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt              time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// Article represents an indexed Usenet article.
type Article struct {
	bun.BaseModel `bun:"table:articles,alias:a"`

	ID               int64      `bun:"id,pk,autoincrement" json:"id"`
	MessageID        string     `bun:"message_id,notnull,unique" json:"message_id"`
	NewsgroupID      int64      `bun:"newsgroup_id,notnull,index" json:"newsgroup_id"`
	NewsgroupName    string     `bun:"-" json:"-"` // Populated via join
	Subject          string     `bun:"subject,notnull" json:"subject"`
	Author           string     `bun:"author" json:"author"`
	PostedAt         time.Time  `bun:"posted_at,notnull" json:"posted_at"`
	ArticleSize      int64      `bun:"article_size" json:"article_size"`
	RawMetadata      string     `bun:"raw_metadata,type:jsonb" json:"raw_metadata,omitempty"`
	NormalizedMetadata string   `bun:"normalized_metadata,type:jsonb" json:"normalized_metadata,omitempty"`
	SearchVector     string     `bun:"search_vector,type:tsvector" json:"-"`
	CreatedAt        time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}

// SearchQuery holds parameters for a search operation.
type SearchQuery struct {
	Query       string
	Category    string
	Newsgroup   string
	SourceType  string
	Author      string
	MinSize     *int64
	MaxSize     *int64
	MinDate     *time.Time
	MaxDate     *time.Time
	FileType    string
	ContentType string
	ProcessingStatus string
	SortBy      string
	SortOrder   string
	Page        int
	PerPage     int
}

// Pagination holds pagination parameters.
type Pagination struct {
	Page    int
	PerPage int
}

// NewPagination creates a new pagination with defaults.
func NewPagination(page, perPage int) Pagination {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	return Pagination{Page: page, PerPage: perPage}
}

// Offset returns the SQL offset.
func (p Pagination) Offset() int {
	return (p.Page - 1) * p.PerPage
}
