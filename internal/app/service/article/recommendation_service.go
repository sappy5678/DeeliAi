package article

import (
	"context"
	"fmt"
	"time"

	gocron "github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/sappy5678/DeeliAi/internal/domain/article"
	"github.com/sappy5678/DeeliAi/internal/domain/common"
)

// RecommendationService defines the interface for article recommendation operations.
type RecommendationService interface {
	GetRecommendations(ctx context.Context, userID uuid.UUID, limit int) ([]article.Recommendation, common.Error)
}

// recommendationService implements RecommendationService.
type recommendationService struct {
	articleRepo ArticleRepository // Interface for articles table operations
	scheduler   gocron.Scheduler  // For refreshing materialized views
}

// NewRecommendationService creates a new RecommendationService.
func NewRecommendationService(articleRepo ArticleRepository) RecommendationService {
	scheduler, _ := gocron.NewScheduler(gocron.WithLocation(time.UTC))

	r := &recommendationService{
		articleRepo: articleRepo,
		scheduler:   scheduler,
	}

	r.scheduler.NewJob(
		gocron.DurationJob(1*time.Minute), // for dev, adjust to 1 day in production
		gocron.NewTask(r.refreshMaterializedView),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
		gocron.WithName("MaterializedViewRefresher"),
	)

	r.scheduler.Start()

	return r
}

// GetRecommendations implements RecommendationService.
func (s *recommendationService) GetRecommendations(ctx context.Context, userID uuid.UUID, limit int) ([]article.Recommendation, common.Error) {
	// Get top rated articles excluding user's collection in one go
	articles, err := s.articleRepo.GetTopRatedArticlesExcludingUser(ctx, userID, limit)
	if err != nil {
		return nil, common.NewError(common.ErrorCodeInternalProcess, fmt.Errorf("failed to get top rated articles: %w", err))
	}

	// Convert articles to recommendations
	recommendations := make([]article.Recommendation, len(articles))
	for i, a := range articles {
		recommendations[i] = article.Recommendation{
			Article: *a,
			Score:   a.AverageRating,
		}
	}

	return recommendations, nil
}

// refreshMaterializedView refreshes the materialized view
func (r *recommendationService) refreshMaterializedView() common.Error {
	r.logger(context.Background()).Info().Msg("refreshing materialized view")
	if err := r.articleRepo.RefreshMaterializedView(context.Background()); err != nil {
		r.logger(context.Background()).Error().Str("err", err.Error()).Msg("refreshing materialized view")
		return err
	}
	return nil
}

// logger wrap the execution context with component info
func (s *recommendationService) logger(ctx context.Context) *zerolog.Logger {
	l := zerolog.Ctx(ctx).With().Str("component", "recommendation-service").Logger()
	return &l
}
