package article

import (
	"context"
	"time"

	gocron "github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/sappy5678/DeeliAi/internal/domain/article"
	"github.com/sappy5678/DeeliAi/internal/domain/common"
)

// recommendationService implements RecommendationService.
type recommendationService struct {
	articleRepo ArticleRepository // Interface for articles table operations
	scheduler   gocron.Scheduler  // For refreshing materialized views
}

// NewRecommendationService creates a new RecommendationService.
func NewRecommendationService(ctx context.Context, articleRepo ArticleRepository) RecommendationService {
	scheduler, _ := gocron.NewScheduler(gocron.WithLocation(time.UTC))

	r := &recommendationService{
		articleRepo: articleRepo,
		scheduler:   scheduler,
	}

	r.scheduler.NewJob(
		gocron.DurationJob(1*time.Minute), // for dev, adjust to 1 day in production
		gocron.NewTask(r.refreshMaterializedView, ctx),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
		gocron.WithName("MaterializedViewRefresher"),
	)

	r.scheduler.Start()

	return r
}

// GetRecommendations implements RecommendationService.
func (s *recommendationService) GetRecommendations(ctx context.Context, userID uuid.UUID, limit int) (article.RecommendationArticles, common.Error) {
	recommendations, err := s.articleRepo.GetTopRatedArticlesExcludingUser(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	return recommendations, nil
}

// refreshMaterializedView refreshes the materialized view
func (r *recommendationService) refreshMaterializedView(ctx context.Context) common.Error {
	r.logger(ctx).Info().Msg("refreshing materialized view")
	if err := r.articleRepo.RefreshMaterializedView(ctx); err != nil {
		r.logger(ctx).Error().Str("err", err.Error()).Msg("refreshing materialized view")
		return err
	}
	return nil
}

// logger wrap the execution context with component info
func (s *recommendationService) logger(ctx context.Context) *zerolog.Logger {
	l := zerolog.Ctx(ctx).With().Str("component", "recommendation-service").Logger()
	return &l
}
