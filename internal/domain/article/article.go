package article

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Article struct {
	ID          uuid.UUID
	URL         string
	Title       string
	Description string
	ImageURL    string
	Metadata    json.RawMessage
}

type RetryStatus int8

const (
	RetryStatusPending RetryStatus = iota // 0
	RetryStatusSuccess                    // 1
	RetryStatusFailed                     // 2
)

type MetadataFetchRetry struct {
	ID            int64
	ArticleID     uuid.UUID
	URL           string
	RetryCount    int16
	LastAttemptAt *time.Time
	NextAttemptAt *time.Time
	Status        RetryStatus
	ErrorMessage  string
}
