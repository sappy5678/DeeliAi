package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/sappy5678/DeeliAi/internal/domain/common"
)

type authService struct {
	TokenService
}

func NewAuthService(_ context.Context, tokenService TokenService) AuthService { // Return AuthService interface
	return &authService{
		TokenService: tokenService,
	}
}

// ValidateUserToken validates the given token and returns the user ID if valid.
func (s *authService) ValidateUserToken(ctx context.Context, token string) (uuid.UUID, common.Error) { // Use authService
	userID, err := s.TokenService.ValidateToken(token)
	if err != nil {
		return uuid.Nil, common.NewError(common.ErrorCodeAuthNotAuthenticated, err)
	}

	return userID, nil
}

// logger wrap the execution context with component info
func (s *authService) logger(ctx context.Context) *zerolog.Logger { // Use authService
	l := zerolog.Ctx(ctx).With().Str("component", "auth-service").Logger()
	return &l
}
