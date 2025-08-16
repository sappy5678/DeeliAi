package article

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Article represents an article in the system.
type Article struct {
	ID          uuid.UUID
	URL         string
	Title       string
	Description string
	ImageURL    string
	Metadata    json.RawMessage
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
