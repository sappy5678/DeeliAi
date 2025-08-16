package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/sappy5678/DeeliAi/internal/adapter/repository/postgres"
	"github.com/sappy5678/DeeliAi/internal/domain/common"
)

type AuthServiceImpl struct {
	userRepo postgres.UserRepository
	tokenSrv TokenService

	signingKey     []byte
	expiryDuration time.Duration
	issuer         string
}

type AuthServiceParam struct {
	UserRepo postgres.UserRepository

	SigningKey     []byte
	ExpiryDuration time.Duration
	Issuer         string
}

func NewAuthService(_ context.Context, param AuthServiceParam) AuthService { // Return AuthService interface
	return &AuthServiceImpl{
		userRepo: param.UserRepo,
		tokenSrv: NewTokenService(param.SigningKey, param.ExpiryDuration, param.Issuer),

		signingKey:     param.SigningKey,
		expiryDuration: param.ExpiryDuration,
		issuer:         param.Issuer,
	}
}

// ValidateUserToken validates the given token and returns the user ID if valid.
func (s *AuthServiceImpl) ValidateUserToken(ctx context.Context, token string) (uuid.UUID, common.Error) { // Use AuthServiceImpl
	userID, err := s.tokenSrv.ValidateToken(token)
	if err != nil {
		return uuid.Nil, common.NewError(common.ErrorCodeAuthNotAuthenticated, err)
	}

	return userID, nil
}

// logger wrap the execution context with component info
func (s *AuthServiceImpl) logger(ctx context.Context) *zerolog.Logger { // Use AuthServiceImpl
	l := zerolog.Ctx(ctx).With().Str("component", "auth-service").Logger()
	return &l
}
