package auth

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"strings"
	"time"
)

// TokenClaims defines the structure for the JWT claims.
type TokenClaims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

type Service interface {
	GetClaims(token string) (claims *TokenClaims, err error)
}

type JwtService struct {
	codePhrase string
	TokenTTL   time.Duration
}

func NewJwtService(codePhrase string) *JwtService {
	return &JwtService{codePhrase: codePhrase}
}

func (service JwtService) GetClaims(token string) (claims *TokenClaims, err error) {
	split := strings.Split(token, " ")

	if len(split) != 2 || split[0] != "Bearer" {
		err = errors.New("invalid token format")
		return
	}
	claims = &TokenClaims{}

	jwtToken, err := jwt.ParseWithClaims(split[1], claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(service.codePhrase), nil
	})
	if err != nil {
		return nil, err
	}

	// Validate the token and return the claims.
	if claims, ok := jwtToken.Claims.(*TokenClaims); ok && jwtToken.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func (service JwtService) GenerateToken(userID int) (string, error) {
	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &TokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(service.TokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ID:        uuid.New().String(),
		},
	})

	// Sign token with secret
	tokenString, err := token.SignedString([]byte(service.codePhrase))

	if err != nil {
		return "", errors.New("could not sign the token")
	}

	return tokenString, nil
}
