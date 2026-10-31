package models

import (
	"time"

	"github.com/uptrace/bun"
)

// SpotnetSource represents a configured Spotnet source.
type SpotnetSource struct {
	bun.BaseModel `bun:"table:spotnet_sources,alias:src"`

	ID            int64      `bun:"id,pk,autoincrement" json:"id"`
	Name          string     `bun:"name,notnull" json:"name"`
	SourceType    string     `bun:"source_type,notnull" json:"source_type"`
	Configuration string     `bun:"configuration,type:jsonb" json:"configuration,omitempty"`
	Enabled       bool       `bun:"enabled,notnull,default:true" json:"enabled"`
	LastSyncAt    *time.Time `bun:"last_sync_at" json:"last_sync_at,omitempty"`
	CreatedAt     time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt     time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// SpotnetPost represents a post from a Spotnet source.
type SpotnetPost struct {
	bun.BaseModel `bun:"table:spotnet_posts,alias:sp"`

	ID                 int64      `bun:"id,pk,autoincrement" json:"id"`
	ExternalID         string     `bun:"external_id,notnull" json:"external_id"`
	Title              string     `bun:"title,notnull" json:"title"`
	Description        string     `bun:"description" json:"description"`
	Author             string     `bun:"author" json:"author"`
	CategoryID         *int64     `bun:"category_id" json:"category_id,omitempty"`
	CategoryName       string     `bun:"-" json:"-"` // Populated via join
	PostedAt           time.Time  `bun:"posted_at,notnull" json:"posted_at"`
	SourceID           int64      `bun:"source_id,notnull,index" json:"source_id"`
	SourceName         string     `bun:"-" json:"-"` // Populated via join
	RawMetadata        string     `bun:"raw_metadata,type:jsonb" json:"raw_metadata,omitempty"`
	NormalizedMetadata string     `bun:"normalized_metadata,type:jsonb" json:"normalized_metadata,omitempty"`
	SearchVector       string     `bun:"search_vector,type:tsvector" json:"-"`
	CreatedAt          time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt          time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// Category represents a content category.
type Category struct {
	bun.BaseModel `bun:"table:categories,alias:c"`

	ID        int64     `bun:"id,pk,autoincrement" json:"id"`
	Name      string    `bun:"name,notnull,unique" json:"name"`
	Slug      string    `bun:"slug,notnull,unique" json:"slug"`
	ParentID  *int64    `bun:"parent_id" json:"parent_id,omitempty"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// IndexingJob represents a background indexing job.
type IndexingJob struct {
	bun.BaseModel `bun:"table:indexing_jobs,alias:ij"`

	ID              int64      `bun:"id,pk,autoincrement" json:"id"`
	JobType         string     `bun:"job_type,notnull" json:"job_type"`
	SourceID        *int64     `bun:"source_id" json:"source_id,omitempty"`
	NewsgroupID     *int64     `bun:"newsgroup_id" json:"newsgroup_id,omitempty"`
	Status          string     `bun:"status,notnull,default:'pending'" json:"status"`
	Priority        int        `bun:"priority,notnull,default:0" json:"priority"`
	Progress        float64    `bun:"progress,notnull,default:0" json:"progress"`
	RecordsProcessed int64     `bun:"records_processed,notnull,default:0" json:"records_processed"`
	RecordsFailed   int64      `bun:"records_failed,notnull,default:0" json:"records_failed"`
	ErrorMessage    string     `bun:"error_message" json:"error_message,omitempty"`
	StartedAt       *time.Time `bun:"started_at" json:"started_at,omitempty"`
	CompletedAt     *time.Time `bun:"completed_at" json:"completed_at,omitempty"`
	CreatedAt       time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}

// IndexingCheckpoint represents a processing checkpoint.
type IndexingCheckpoint struct {
	bun.BaseModel `bun:"table:indexing_checkpoints,alias:ic"`

	ID                   int64      `bun:"id,pk,autoincrement" json:"id"`
	SourceType           string     `bun:"source_type,notnull" json:"source_type"`
	SourceID             int64      `bun:"source_id,notnull" json:"source_id"`
	NewsgroupID          *int64     `bun:"newsgroup_id" json:"newsgroup_id,omitempty"`
	LastProcessedPosition string   `bun:"last_processed_position,notnull" json:"last_processed_position"`
	LastProcessedAt      *time.Time `bun:"last_processed_at" json:"last_processed_at,omitempty"`
	UpdatedAt            time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// UserSearchHistory represents a user's search history entry.
type UserSearchHistory struct {
	bun.BaseModel `bun:"table:user_search_history,alias:ush"`

	ID          int64      `bun:"id,pk,autoincrement" json:"id"`
	UserID      int64      `bun:"user_id,notnull,index" json:"user_id"`
	SearchQuery string     `bun:"search_query,notnull" json:"search_query"`
	Filters     string     `bun:"filters,type:jsonb" json:"filters,omitempty"`
	CreatedAt   time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}

// Bookmark represents a user bookmark.
type Bookmark struct {
	bun.BaseModel `bun:"table:bookmarks,alias:b"`

	ID         int64     `bun:"id,pk,autoincrement" json:"id"`
	UserID     int64     `bun:"user_id,notnull" json:"user_id"`
	ItemID     int64     `bun:"item_id,notnull" json:"item_id"`
	ItemType   string    `bun:"item_type,notnull" json:"item_type"`
	CreatedAt  time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}

// SystemSetting represents a system configuration setting.
type SystemSetting struct {
	bun.BaseModel `bun:"table:system_settings,alias:ss"`

	ID          int64     `bun:"id,pk,autoincrement" json:"id"`
	SettingKey  string    `bun:"setting_key,notnull,unique" json:"setting_key"`
	SettingValue string   `bun:"setting_value,type:jsonb,notnull" json:"setting_value"`
	UpdatedAt   time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}
