package article

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// UserArticle represents the link between a user and an article,
// including the user's rating for the article.
type UserArticle struct {
	ID          int64
	UserID      uuid.UUID
	ArticleID   uuid.UUID
	Rate        int16
	CollectedAt time.Time
}

// Rate sets the rating for the article, ensuring it's within the valid range.
func (ua *UserArticle) Rating(rate int16) error {
	if rate < 1 || rate > 5 {
		return fmt.Errorf("rate must be between 1 and 5, but got %d", rate)
	}
	ua.Rate = rate
	return nil
}
