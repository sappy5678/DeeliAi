package article

import (
	"context"

	"github.com/google/uuid"

	"github.com/sappy5678/DeeliAi/internal/domain/article"
	"github.com/sappy5678/DeeliAi/internal/domain/common"
)

// ArticleRepository defines the interface for interacting with article and user_article data.
type ArticleRepository interface {
	CreateArticle(ctx context.Context, url string) (*article.Article, common.Error)
	CreateUserArticle(ctx context.Context, userID uuid.UUID, articleID uuid.UUID) (*article.UserArticle, common.Error)
	GetUserArticle(ctx context.Context, userID uuid.UUID, articleID uuid.UUID) (*article.UserArticle, common.Error)
	ListArticles(ctx context.Context, userID uuid.UUID, afterID uuid.UUID, limit int) ([]*article.Article, common.Error)
	DeleteUserArticle(ctx context.Context, userID uuid.UUID, articleID uuid.UUID) common.Error
	UpdateUserArticleRate(ctx context.Context, userID uuid.UUID, articleID uuid.UUID, rate int16) common.Error
	DeleteUserArticleRate(ctx context.Context, userID uuid.UUID, articleID uuid.UUID) common.Error
	CreateMetadataFetchRetry(ctx context.Context, articleID uuid.UUID, url string) common.Error
	GetPendingMetadataFetchRetries(ctx context.Context) ([]*article.MetadataFetchRetry, common.Error)
	UpdateMetadataFetchRetryStatus(ctx context.Context, retryID int64, status int16, errorMessage string) common.Error
	IncrementMetadataFetchRetryCount(ctx context.Context, retryID int64) common.Error
	GetArticleByID(ctx context.Context, articleID uuid.UUID) (*article.Article, common.Error)
	UpdateArticle(ctx context.Context, art *article.Article) common.Error
	GetTopRatedArticlesExcludingUser(ctx context.Context, excludeUserID uuid.UUID, limit int) (article.RecommendationArticles, common.Error)
	RefreshMaterializedView(ctx context.Context) common.Error
}

// ArticleService defines the interface for article-related business logic.
type ArticleService interface {
	CreateArticle(ctx context.Context, userID uuid.UUID, url string) (*article.Article, common.Error)
	ListArticles(ctx context.Context, userID uuid.UUID, afterID uuid.UUID, limit int) ([]*article.Article, common.Error)
	DeleteArticle(ctx context.Context, userID uuid.UUID, articleID uuid.UUID) common.Error
	RateArticle(ctx context.Context, userID uuid.UUID, articleID uuid.UUID, rate int16) common.Error
	GetArticleRating(ctx context.Context, userID uuid.UUID, articleID uuid.UUID) (*article.UserArticle, common.Error)
	DeleteArticleRating(ctx context.Context, userID uuid.UUID, articleID uuid.UUID) common.Error
	GetRecommendations(ctx context.Context, userID uuid.UUID, limit int) (article.RecommendationArticles, common.Error)
}
