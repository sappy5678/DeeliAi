package user

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/sappy5678/DeeliAi/internal/domain/common"
)

type TokenServiceImpl struct {
	signingKey     []byte
	expiryDuration time.Duration
	issuer         string
}

func NewTokenService(_ context.Context, signingKey []byte, expiryDuration time.Duration, issuer string) TokenService {
	return &TokenServiceImpl{
		signingKey:     signingKey,
		expiryDuration: expiryDuration,
		issuer:         issuer,
	}
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

func (s *TokenServiceImpl) GenerateToken(ctx context.Context, userID uuid.UUID) (string, common.Error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiryDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.issuer,
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.signingKey)
	if err != nil {
		return "", common.NewError(common.ErrorCodeInternalProcess, err)
	}

	return signedToken, nil
}

func (s *TokenServiceImpl) ValidateToken(ctx context.Context, tokenString string) (uuid.UUID, common.Error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.signingKey, nil
	})
	if err != nil {
		return uuid.Nil, common.NewError(common.ErrorCodeParameterInvalid, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return uuid.Nil, common.NewError(common.ErrorCodeParameterInvalid, jwt.ErrInvalidKey)
	}

	return claims.UserID, nil
}
