package article

import (
	"context"

	"github.com/google/uuid"

	"github.com/sappy5678/DeeliAi/internal/domain/article"
	"github.com/sappy5678/DeeliAi/internal/domain/common"
)

type articleService struct {
	articleRepo ArticleRepository
	RecommendationService
	metadataWorker *MetadataWorker
}

func NewArticleService(articleRepo ArticleRepository) ArticleService {
	service := &articleService{
		articleRepo: articleRepo,
	}

	recommendationService := NewRecommendationService(articleRepo)
	service.RecommendationService = recommendationService

	worker := NewMetadataWorker(service)

	service.metadataWorker = worker

	return service
}

func (s *articleService) CreateArticle(ctx context.Context, userID uuid.UUID, url string) (*article.Article, common.Error) {
	art, err := s.articleRepo.CreateArticle(ctx, url)
	if err != nil {
		return nil, err
	}

	_, err = s.articleRepo.CreateUserArticle(ctx, userID, art.ID)
	if err != nil {
		return nil, err
	}

	err = s.articleRepo.CreateMetadataFetchRetry(ctx, art.ID, art.URL)
	if err != nil {
		return nil, err
	}

	return art, nil
}

func (s *articleService) ListArticles(ctx context.Context, userID uuid.UUID, afterID uuid.UUID, limit int) ([]*article.Article, common.Error) {
	return s.articleRepo.ListArticles(ctx, userID, afterID, limit)
}

func (s *articleService) DeleteArticle(ctx context.Context, userID uuid.UUID, articleID uuid.UUID) common.Error {
	return s.articleRepo.DeleteUserArticle(ctx, userID, articleID)
}

func (s *articleService) RateArticle(ctx context.Context, userID uuid.UUID, articleID uuid.UUID, rate int16) common.Error {
	userArticle, err := s.articleRepo.GetUserArticle(ctx, userID, articleID)
	if err != nil {
		return err
	}

	if err := userArticle.Rating(rate); err != nil {
		return common.NewError(common.ErrorCodeParameterInvalid, err)
	}

	return s.articleRepo.UpdateUserArticleRate(ctx, userID, articleID, userArticle.Rate)
}

func (s *articleService) GetArticleRating(ctx context.Context, userID uuid.UUID, articleID uuid.UUID) (*article.UserArticle, common.Error) {
	return s.articleRepo.GetUserArticle(ctx, userID, articleID)
}

func (s *articleService) DeleteArticleRating(ctx context.Context, userID uuid.UUID, articleID uuid.UUID) common.Error {
	_, err := s.articleRepo.GetUserArticle(ctx, userID, articleID)
	if err != nil {
		return err
	}
	return s.articleRepo.DeleteUserArticleRate(ctx, userID, articleID)
}
