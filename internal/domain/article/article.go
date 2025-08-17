package article

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Article represents an article in the system.
type Article struct {
	ID            uuid.UUID
	URL           string
	Title         string
	Description   string
	ImageURL      string
	Metadata      json.RawMessage
	CreatedAt     time.Time
	UpdatedAt     time.Time
	AverageRating float64
}

type RetryStatus int8

const (
	RetryStatusPending RetryStatus = iota // 0
	RetryStatusSuccess                    // 1
	RetryStatusFailed                     // 2
)

// MetadataFetchRetry represents a retry attempt for metadata fetching.
type MetadataFetchRetry struct {
	ID            int64
	ArticleID     uuid.UUID
	URL           string
	RetryCount    int16
	LastAttemptAt *time.Time
	NextAttemptAt *time.Time
	Status        RetryStatus
	ErrorMessage  string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
