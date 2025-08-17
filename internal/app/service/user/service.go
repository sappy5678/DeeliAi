package user

import (
	"context"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/sappy5678/DeeliAi/internal/adapter/repository/postgres"
	"github.com/sappy5678/DeeliAi/internal/domain/common"
	"github.com/sappy5678/DeeliAi/internal/domain/user"
)

type userService struct {
	TokenService
	userRepo postgres.UserRepository
}

func NewUserService(ctx context.Context, userRepo postgres.UserRepository, authService TokenService) Service {
	return &userService{
		userRepo:     userRepo,
		TokenService: authService,
	}
}

func (s *userService) SignUp(ctx context.Context, email string, username string, password string) (*user.User, string, common.Error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", common.NewError(common.ErrorCodeInternalProcess, err)
	}

	newUser := &user.User{
		Email:        email,
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	createdUser, cerr := s.userRepo.CreateUser(ctx, newUser)
	if cerr != nil {
		return nil, "", cerr
	}

	token, err := s.TokenService.GenerateToken(ctx, createdUser.ID)
	if err != nil {
		return nil, "", common.NewError(common.ErrorCodeInternalProcess, err)
	}

	return createdUser, token, nil
}

func (s *userService) Login(ctx context.Context, email string, password string) (*user.User, string, common.Error) {
	foundUser, cerr := s.userRepo.GetUserByEmail(ctx, email)
	if cerr != nil {
		return nil, "", cerr
	}

	err := bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password))
	if err != nil {
		return nil, "", common.NewError(common.ErrorCodeAuthNotAuthenticated, err, common.WithMsg("invalid password")) // Changed here
	}

	token, err := s.TokenService.GenerateToken(ctx, foundUser.ID)
	if err != nil {
		return nil, "", common.NewError(common.ErrorCodeInternalProcess, err)
	}

	return foundUser, token, nil
}

func (s *userService) GetUser(ctx context.Context, userID uuid.UUID) (*user.User, common.Error) {
	foundUser, cerr := s.userRepo.GetUserByID(ctx, userID)
	if cerr != nil {
		return nil, cerr
	}

	return foundUser, nil
}
