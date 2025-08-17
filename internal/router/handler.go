package router

import (
	"github.com/gin-gonic/gin"

	// Import postgres
	"github.com/sappy5678/DeeliAi/internal/app"
	// Import user service
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

	// Add user namespace
	userGroup := v1.Group("/user")
	{
		userGroup.POST("/signup", SignUp(app))
		userGroup.POST("/login", Login(app))
		userGroup.GET("/me", BearerToken.Required(), GetCurrentUser(app))
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
		articleGroup.GET("/recommendations", GetRecommendations(app))
	}
}
