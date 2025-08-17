package router

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/sappy5678/DeeliAi/internal/app"
	"github.com/sappy5678/DeeliAi/internal/app/service/article"
	"github.com/sappy5678/DeeliAi/internal/domain/common"
)

// RecommendationHandler handles recommendation-related HTTP requests.
type RecommendationHandler struct {
	recommendationService article.RecommendationService
}

// NewRecommendationHandler creates a new RecommendationHandler.
func NewRecommendationHandler(svc article.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{
		recommendationService: svc,
	}
}

func GetRecommendations(app *app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := GetCurrentUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		limitStr := c.DefaultQuery("limit", "10")
		limit, atoiErr := strconv.Atoi(limitStr)
		if atoiErr != nil || limit < 1 || limit > 50 {
			respondWithError(c, common.NewError(common.ErrorCodeParameterInvalid, atoiErr, common.WithMsg("Invalid limit parameter")))
			return
		}

		recommendations, svcErr := app.ArticleService.GetRecommendations(c.Request.Context(), userID, limit)
		if svcErr != nil {
			respondWithError(c, common.NewError(common.ErrorCodeInternalProcess, svcErr, common.WithMsg("Failed to get recommendations")))
			return
		}

		c.JSON(http.StatusOK, gin.H{"items": recommendations})
	}
}
