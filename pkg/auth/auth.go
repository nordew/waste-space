package auth

import (
	"time"

	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type TokenService interface {
	GenerateTokenPair(userID uuid.UUID, email string) (*TokenPair, error)
	ValidateToken(token string) (*Claims, error)
	RefreshAccessToken(refreshToken string) (string, error)
}
