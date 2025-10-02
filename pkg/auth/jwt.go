package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	apperrors "waste-space/pkg/errors"
)

const (
	defaultAccessTokenExpiry  = 15 * time.Minute
	defaultRefreshTokenExpiry = 7 * 24 * time.Hour
)

type jwtService struct {
	secretKey       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type tokenClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Type   string    `json:"type"`
	jwt.RegisteredClaims
}

func NewJWTService(secretKey string) TokenService {
	return &jwtService{
		secretKey:       []byte(secretKey),
		accessTokenTTL:  defaultAccessTokenExpiry,
		refreshTokenTTL: defaultRefreshTokenExpiry,
	}
}

func NewJWTServiceWithTTL(secretKey string, accessTTL, refreshTTL time.Duration) TokenService {
	return &jwtService{
		secretKey:       []byte(secretKey),
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

func (s *jwtService) GenerateTokenPair(userID uuid.UUID, email string) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(s.accessTokenTTL)
	refreshExpiry := now.Add(s.refreshTokenTTL)

	accessToken, err := s.generateToken(userID, email, "access", accessExpiry)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateToken(userID, email, "refresh", refreshExpiry)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiry,
	}, nil
}

func (s *jwtService) ValidateToken(token string) (*Claims, error) {
	claims := &tokenClaims{}

	t, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperrors.Unauthorized("invalid token")
		}
		return s.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperrors.Unauthorized("token has expired")
		}
		return nil, apperrors.Unauthorized("invalid token")
	}

	if !t.Valid {
		return nil, apperrors.Unauthorized("invalid token")
	}

	if claims.Type != "access" {
		return nil, apperrors.Unauthorized("invalid token")
	}

	return &Claims{
		UserID: claims.UserID,
		Email:  claims.Email,
	}, nil
}

func (s *jwtService) RefreshAccessToken(refreshToken string) (string, error) {
	claims := &tokenClaims{}

	t, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperrors.Unauthorized("invalid token")
		}
		return s.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", apperrors.Unauthorized("token has expired")
		}
		return "", apperrors.Unauthorized("invalid token")
	}

	if !t.Valid || claims.Type != "refresh" {
		return "", apperrors.Unauthorized("invalid token")
	}

	accessExpiry := time.Now().Add(s.accessTokenTTL)
	return s.generateToken(claims.UserID, claims.Email, "access", accessExpiry)
}

func (s *jwtService) generateToken(userID uuid.UUID, email, tokenType string, expiresAt time.Time) (string, error) {
	claims := tokenClaims{
		UserID: userID,
		Email:  email,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}
