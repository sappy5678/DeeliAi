package user

import (
	"context"

	"github.com/google/uuid"

	"github.com/sappy5678/DeeliAi/internal/domain/common"
	"github.com/sappy5678/DeeliAi/internal/domain/user"
)

type Service interface {
	AuthService
	SignUp(ctx context.Context, email string, username string, password string) (*user.User, string, common.Error)
	Login(ctx context.Context, email string, password string) (*user.User, string, common.Error)
	GetUser(ctx context.Context, userID uuid.UUID) (*user.User, common.Error)
}

type AuthService interface {
	TokenService
	ValidateUserToken(ctx context.Context, token string) (uuid.UUID, common.Error)
}

type TokenService interface {
	GenerateToken(userID uuid.UUID) (string, error)
	ValidateToken(token string) (uuid.UUID, error)
}

//go:generate mockgen -destination automock/user_repository.go -package=automock . UserRepository
type UserRepository interface {
	CreateUser(ctx context.Context, user *user.User) (*user.User, common.Error)
	GetUserByEmail(ctx context.Context, email string) (*user.User, common.Error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, common.Error)
}
