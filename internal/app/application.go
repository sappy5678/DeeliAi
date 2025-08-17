package app

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/sappy5678/DeeliAi/internal/adapter/repository/postgres"

	"github.com/sappy5678/DeeliAi/internal/app/service/article"
	"github.com/sappy5678/DeeliAi/internal/app/service/user"
)

type Application struct {
	Params         ApplicationParams
	ArticleService article.ArticleService
	UserService    user.Service
}

type ApplicationParams struct {
	// General configuration
	Env string

	// Database parameters
	DatabaseDSN string

	// Token parameter
	TokenSigningKey     []byte
	TokenExpiryDuration time.Duration
	TokenIssuer         string
}

func MustNewApplication(ctx context.Context, wg *sync.WaitGroup, params ApplicationParams) *Application {
	app, err := NewApplication(ctx, wg, params)
	if err != nil {
		log.Panicf("fail to new application, err: %s", err.Error())
	}
	return app
}

func NewApplication(ctx context.Context, wg *sync.WaitGroup, params ApplicationParams) (*Application, error) {
	// Create repositories
	db := sqlx.MustOpen("postgres", params.DatabaseDSN)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	pgRepo := postgres.NewPostgresRepository(ctx, db)

	// service initialization
	tokenService := user.NewTokenService(ctx, params.TokenSigningKey, params.TokenExpiryDuration, params.TokenIssuer)

	// Create application
	app := &Application{
		Params:         params,
		ArticleService: article.NewArticleService(ctx, pgRepo),
		UserService:    user.NewUserService(ctx, pgRepo, tokenService),
	}

	return app, nil
}
