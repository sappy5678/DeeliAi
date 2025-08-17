package router

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/sappy5678/DeeliAi/internal/app"
	"github.com/sappy5678/DeeliAi/internal/domain/common"
)

type AuthMiddlewareBearer struct {
	app *app.Application
}

const (
	ContextKeyUserID = "userID"
)

func NewAuthMiddlewareBearer(app *app.Application) *AuthMiddlewareBearer {
	return &AuthMiddlewareBearer{
		app: app,
	}
}

func (m *AuthMiddlewareBearer) Required() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Get Bearer token
		token, err := GetAuthorizationToken(c)
		if err != nil {
			respondWithError(c, err)
			return
		}
		tokens := strings.Split(token, "Bearer ")
		if len(tokens) != 2 {
			msg := "bearer token is needed"
			respondWithError(c, common.NewError(common.ErrorCodeAuthNotAuthenticated, errors.New(msg), common.WithMsg(msg)))
			return
		}

		userID, cerr := m.app.UserService.ValidateToken(ctx, tokens[1])
		if cerr != nil {
			respondWithError(c, common.NewError(common.ErrorCodeAuthNotAuthenticated, errors.New(cerr.Error()), common.WithMsg(cerr.ClientMsg())))
			return
		}

		// Set user ID to context
		c.Set(ContextKeyUserID, userID)
		c.Next()
	}
}
