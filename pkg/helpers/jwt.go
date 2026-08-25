package helpers

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	SigningMethod = jwt.SigningMethodHS256
)

type JwtManager struct {
	secretKey []byte
}

func NewJwtManager(secretKey string) *JwtManager {
	return &JwtManager{
		secretKey: []byte(secretKey),
	}
}

type TokenPair struct {
	AccessToken  *jwt.Token
	RefreshToken *jwt.Token
}

type CustomClaims struct {
	TokenType string
	Role      string
	jwt.RegisteredClaims
}

func (m *JwtManager) Parse(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != SigningMethod {
			return nil, errors.New("failed to parse token")
		}
		return m.secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, err
	}
	return claims, nil
}

func (m *JwtManager) CreateTokenPair(userID uuid.UUID, role string) (*TokenPair, error) {
	//create access
	now := time.Now()
	jti, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	accessToken := jwt.NewWithClaims(SigningMethod, CustomClaims{
		"access",
		role,
		jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        jti.String(),
		},
	})

	signingAccessToken, err := accessToken.SignedString(m.secretKey)
	if err != nil {
		return nil, err
	}

	accessToken.Raw = signingAccessToken

	// refresh

	jti2, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	refreshToken := jwt.NewWithClaims(SigningMethod, CustomClaims{
		"refresh",
		"",
		jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * 60 * 30 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        jti2.String(),
		},
	})

	signingRefreshToken, err := refreshToken.SignedString(m.secretKey)
	if err != nil {
		return nil, err
	}
	refreshToken.Raw = signingRefreshToken

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
