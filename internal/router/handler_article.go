package router

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sappy5678/DeeliAi/internal/app"
	"github.com/sappy5678/DeeliAi/internal/domain/common"
)

func CreateArticle(app *app.Application) gin.HandlerFunc {
	type Body struct {
		URL string `json:"url" binding:"required,url"`
	}

	type Response struct {
		ID  uuid.UUID `json:"id"`
		URL string    `json:"url"`
	}

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var body Body
		if err := c.ShouldBindJSON(&body); err != nil {
			respondWithError(c, common.NewError(common.ErrorCodeParameterInvalid, err, common.WithMsg("invalid parameter")))
			return
		}

		userID, err := GetCurrentUserID(c)
		if err != nil {
			respondWithError(c, err)
			return
		}

		art, err := app.ArticleService.CreateArticle(ctx, userID, body.URL)
		if err != nil {
			respondWithError(c, err)
			return
		}

		resp := Response{
			ID:  art.ID,
			URL: art.URL,
		}
		respondWithJSON(c, http.StatusCreated, resp)
	}
}

func ListArticles(app *app.Application) gin.HandlerFunc {
	type Query struct {
		After string `form:"after"`
		Limit int    `form:"limit"`
	}

	type ArticleResponse struct {
		ID          uuid.UUID `json:"id"`
		URL         string    `json:"url"`
		Title       string    `json:"title,omitempty"`
		Description string    `json:"description,omitempty"`
		ImageURL    string    `json:"image_url,omitempty"`
	}

	type Response struct {
		Articles []ArticleResponse `json:"articles"`
	}

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var query Query
		if err := c.ShouldBindQuery(&query); err != nil {
			respondWithError(c, common.NewError(common.ErrorCodeParameterInvalid, err, common.WithMsg("invalid query parameter")))
			return
		}

		var afterID uuid.UUID
		if query.After != "" {
			var parseErr error
			afterID, parseErr = uuid.Parse(query.After)
			if parseErr != nil {
				respondWithError(c, common.NewError(common.ErrorCodeParameterInvalid, parseErr, common.WithMsg("invalid after id")))
				return
			}
		}

		if query.Limit == 0 {
			query.Limit = 10 // default limit
		}

		userID, err := GetCurrentUserID(c)
		if err != nil {
			respondWithError(c, err)
			return
		}

		articles, err := app.ArticleService.ListArticles(ctx, userID, afterID, query.Limit)
		if err != nil {
			respondWithError(c, err)
			return
		}

		resp := Response{
			Articles: make([]ArticleResponse, 0, len(articles)),
		}
		for _, art := range articles {
			resp.Articles = append(resp.Articles, ArticleResponse{
				ID:          art.ID,
				URL:         art.URL,
				Title:       art.Title,
				Description: art.Description,
				ImageURL:    art.ImageURL,
			})
		}

		respondWithJSON(c, http.StatusOK, resp)
	}
}

func DeleteArticle(app *app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		articleID, err := GetParamUUID(c, "article_id")
		if err != nil {
			respondWithError(c, err)
			return
		}

		userID, err := GetCurrentUserID(c)
		if err != nil {
			respondWithError(c, err)
			return
		}

		if err := app.ArticleService.DeleteArticle(ctx, userID, articleID); err != nil {
			respondWithError(c, err)
			return
		}

		respondWithoutBody(c, http.StatusNoContent)
	}
}

func RateArticle(app *app.Application) gin.HandlerFunc {
	type Body struct {
		Rate int16 `json:"rate"`
	}

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		articleID, err := GetParamUUID(c, "article_id")
		if err != nil {
			respondWithError(c, err)
			return
		}

		var body Body
		if err := c.ShouldBindJSON(&body); err != nil {
			respondWithError(c, common.NewError(common.ErrorCodeParameterInvalid, err, common.WithMsg("invalid parameter")))
			return
		}

		userID, err := GetCurrentUserID(c)
		if err != nil {
			respondWithError(c, err)
			return
		}

		if err := app.ArticleService.RateArticle(ctx, userID, articleID, body.Rate); err != nil {
			respondWithError(c, err)
			return
		}

		respondWithoutBody(c, http.StatusNoContent)
	}
}

func GetArticleRating(app *app.Application) gin.HandlerFunc {
	type Response struct {
		Rate int16 `json:"rate"`
	}

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		articleID, err := GetParamUUID(c, "article_id")
		if err != nil {
			respondWithError(c, err)
			return
		}

		userID, err := GetCurrentUserID(c)
		if err != nil {
			respondWithError(c, err)
			return
		}

		userArticle, err := app.ArticleService.GetArticleRating(ctx, userID, articleID)
		if err != nil {
			respondWithError(c, err)
			return
		}

		respondWithJSON(c, http.StatusOK, Response{Rate: userArticle.Rate})
	}
}

func DeleteArticleRating(app *app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		articleID, err := GetParamUUID(c, "article_id")
		if err != nil {
			respondWithError(c, err)
			return
		}

		userID, err := GetCurrentUserID(c)
		if err != nil {
			respondWithError(c, err)
			return
		}

		if err := app.ArticleService.DeleteArticleRating(ctx, userID, articleID); err != nil {
			respondWithError(c, err)
			return
		}

		respondWithoutBody(c, http.StatusNoContent)
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
