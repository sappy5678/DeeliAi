package user

import (
	"context"

	"github.com/google/uuid"

	"github.com/sappy5678/DeeliAi/internal/domain/common"
	"github.com/sappy5678/DeeliAi/internal/domain/user"
)

type Service interface {
	TokenService
	SignUp(ctx context.Context, email string, username string, password string) (*user.User, string, common.Error)
	Login(ctx context.Context, email string, password string) (*user.User, string, common.Error)
	GetUser(ctx context.Context, userID uuid.UUID) (*user.User, common.Error)
}

type TokenService interface {
	GenerateToken(ctx context.Context, userID uuid.UUID) (string, common.Error)
	ValidateToken(ctx context.Context, token string) (uuid.UUID, common.Error)
}

//go:generate mockgen -destination automock/user_repository.go -package=automock . UserRepository
type UserRepository interface {
	CreateUser(ctx context.Context, user *user.User) (*user.User, common.Error)
	GetUserByEmail(ctx context.Context, email string) (*user.User, common.Error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, common.Error)
}
