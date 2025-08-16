package user

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenServiceImpl struct {
	signingKey     []byte
	expiryDuration time.Duration
	issuer         string
}

func NewTokenService(signingKey []byte, expiryDuration time.Duration, issuer string) TokenService {
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

func (s *TokenServiceImpl) GenerateToken(userID uuid.UUID) (string, error) {
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
		return "", err
	}

	return signedToken, nil
}

func (s *TokenServiceImpl) ValidateToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.signingKey, nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return uuid.Nil, jwt.ErrInvalidKey
	}

	return claims.UserID, nil
}
