package service

import (
	"context"
	"time"
	"waste-space/internal/dto"
	"waste-space/internal/model"
	"waste-space/internal/storage/cache"
	"waste-space/internal/storage/repository"
	"waste-space/pkg/auth"
	apperrors "waste-space/pkg/errors"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

const refreshTokenTTL = 7 * 24 * time.Hour

type UserService interface {
	Register(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error)
	Logout(ctx context.Context, userID string, accessToken string) error
	GetByID(ctx context.Context, userID string) (*dto.UserResponse, error)
}

type userService struct {
	userRepo     repository.UserRepository
	tokenService auth.TokenService
	tokenCache   cache.TokenCache
}

func NewUserService(userRepo repository.UserRepository, tokenService auth.TokenService, tokenCache cache.TokenCache) UserService {
	return &userService{
		userRepo:     userRepo,
		tokenService: tokenService,
		tokenCache:   tokenCache,
	}
}

func (s *userService) Register(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error) {
	user, err := model.NewUserFromDTO(req)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *userService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperrors.Unauthorized("invalid email or password")
	}

	if !user.IsActive {
		return nil, apperrors.Forbidden("user account is inactive")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperrors.Unauthorized("invalid email or password")
	}

	tokenPair, err := s.tokenService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, apperrors.Internal("failed to generate tokens", err)
	}

	if err := s.tokenCache.SetRefreshToken(ctx, user.ID, tokenPair.RefreshToken, refreshTokenTTL); err != nil {
		return nil, apperrors.Internal("failed to cache refresh token", err)
	}

	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &dto.LoginResponse{
		User:         response,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

func (s *userService) RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error) {
	claims, err := s.tokenService.ValidateToken(req.RefreshToken)
	if err != nil {
		claims, err = s.tokenService.ValidateToken(req.RefreshToken)
		if err != nil {
			return nil, apperrors.Unauthorized("invalid refresh token")
		}
	}

	cachedToken, err := s.tokenCache.GetRefreshToken(ctx, claims.UserID)
	if err != nil {
		if err == redis.Nil {
			return nil, apperrors.Unauthorized("refresh token expired or revoked")
		}
		return nil, apperrors.Internal("failed to get cached token", err)
	}

	if cachedToken != req.RefreshToken {
		return nil, apperrors.Unauthorized("invalid refresh token")
	}

	accessToken, err := s.tokenService.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &dto.RefreshTokenResponse{
		AccessToken: accessToken,
	}, nil
}

func (s *userService) Logout(ctx context.Context, userID string, accessToken string) error {
	return nil
}

func (s *userService) GetByID(ctx context.Context, userID string) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}
