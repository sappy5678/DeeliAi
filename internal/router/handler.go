package router

import (
	"github.com/gin-gonic/gin"

	"github.com/sappy5678/DeeliAi/internal/adapter/repository/postgres" // Import postgres
	"github.com/sappy5678/DeeliAi/internal/app"
	"github.com/sappy5678/DeeliAi/internal/app/service/user" // Import user service
)

func RegisterHandlers(router *gin.Engine, app *app.Application) {
	registerAPIHandlers(router, app)
}

func registerAPIHandlers(router *gin.Engine, app *app.Application) {
	// Build middlewares
	BearerToken := NewAuthMiddlewareBearer(app)

	// We mount all handlers under /api path
	r := router.Group("/api")
	v1 := r.Group("/v1")

	// Add health-check
	v1.GET("/health", handlerHealthCheck())

	// Initialize UserHandler
	userHandler := NewUserHandler(user.NewUserService(
		app.UserRepository.(postgres.UserRepository), // Explicitly cast
		app.TokenService.(user.TokenService),
	))

	// Add user namespace
	userGroup := v1.Group("/user")
	{
		userGroup.POST("/signup", userHandler.SignUp)
		userGroup.POST("/login", userHandler.Login)
		userGroup.GET("/me", BearerToken.Required(), userHandler.GetCurrentUser)
	}

	// Add barter namespace
	barterGroup := v1.Group("/barter", BearerToken.Required())
	{
		barterGroup.POST("/goods", PostGood(app))
		barterGroup.GET("/goods", ListMyGoods(app))
		barterGroup.GET("/goods/traders", ListOthersGoods(app))
		barterGroup.DELETE("/goods/:good_id", RemoveMyGood(app))
		barterGroup.POST("/goods/exchange", ExchangeGoods(app))
	}

	// Add articles namespace
	articleGroup := v1.Group("/articles", BearerToken.Required()) // Use BearerToken.Required()
	{
		articleGroup.POST("", CreateArticle(app))
		articleGroup.GET("", ListArticles(app))
		articleGroup.DELETE("/:article_id", DeleteArticle(app))
		articleGroup.PUT("/:article_id/rate", RateArticle(app))
		articleGroup.GET("/:article_id/rate", GetArticleRating(app))
		articleGroup.DELETE("/:article_id/rate", DeleteArticleRating(app))
	}
}
