package router

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/chatbotgang/go-clean-architecture-template/internal/domain/common"
)

const (
	// TODO: This is a temporary solution for development.
	// Remove this and use the real authentication middleware.
	ContextKeyUserID = "userID"
)

// TmpAuthMiddleware is a temporary middleware for development.
// It gets user ID from the "X-User-ID" header and sets it in the context.
func TmpAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetHeader("X-User-ID")
		if userIDStr == "" {
			msg := "X-User-ID header is required for this temporary auth"
			respondWithError(c, common.NewError(common.ErrorCodeAuthNotAuthenticated, errors.New(msg), common.WithMsg(msg)))
			c.Abort()
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			msg := "invalid user ID format in X-User-ID header"
			respondWithError(c, common.NewError(common.ErrorCodeParameterInvalid, err, common.WithMsg(msg)))
			c.Abort()
			return
		}

		c.Set(ContextKeyUserID, userID)
		c.Next()
	}
}

func GetCurrentUserID(c *gin.Context) (uuid.UUID, common.Error) {
	userID, ok := c.Get(ContextKeyUserID)
	if !ok {
		return uuid.Nil, common.NewError(common.ErrorCodeAuthNotAuthenticated, errors.New("user not found in context"))
	}

	return userID.(uuid.UUID), nil
}
