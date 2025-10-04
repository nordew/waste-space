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

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const refreshTokenTTL = 7 * 24 * time.Hour

type UserService interface {
	Register(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error)
	Logout(ctx context.Context, userID string, accessToken string) error
	GetMe(ctx context.Context, userID string) (*dto.UserResponse, error)
	GetByID(ctx context.Context, userID string) (*dto.UserResponse, error)
	UpdateMe(ctx context.Context, userID string, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	UpdateEmail(ctx context.Context, userID string, req dto.UpdateEmailRequest) (*dto.UserResponse, error)
	UpdatePhone(ctx context.Context, userID string, req dto.UpdatePhoneRequest) (*dto.UserResponse, error)
	UpdatePassword(ctx context.Context, userID string, req dto.UpdatePasswordRequest) error
	DeleteMe(ctx context.Context, userID string) error
}

type userService struct {
	userRepo     repository.UserRepository
	tokenService auth.TokenService
	tokenCache   cache.TokenCache
	logger       *zap.Logger
}

func NewUserService(
	userRepo repository.UserRepository,
	tokenService auth.TokenService,
	tokenCache cache.TokenCache,
	logger *zap.Logger) UserService {
	return &userService{
		userRepo:     userRepo,
		tokenService: tokenService,
		tokenCache:   tokenCache,
		logger:       logger,
	}
}

func (s *userService) Register(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error) {
	user, err := model.NewUserFromDTO(req)
	if err != nil {
		s.logger.Error("failed to create user from DTO", zap.Error(err))
		return nil, err
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("failed to create user", zap.String("email", req.Email), zap.Error(err))
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
		s.logger.Error("failed to generate tokens", zap.String("userId", user.ID.String()), zap.Error(err))
		return nil, apperrors.Internal("failed to generate tokens", err)
	}

	if err := s.tokenCache.SetRefreshToken(ctx, user.ID, tokenPair.RefreshToken, refreshTokenTTL); err != nil {
		s.logger.Error("failed to cache refresh token", zap.String("userId", user.ID.String()), zap.Error(err))
		return nil, apperrors.Internal("failed to cache refresh token", err)
	}

	now := time.Now()
	user.LastLoginAt = &now
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to update last login", zap.String("userId", user.ID.String()), zap.Error(err))
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
		s.logger.Error("failed to get cached token", zap.String("userId", claims.UserID.String()), zap.Error(err))
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

func (s *userService) GetMe(ctx context.Context, userID string) (*dto.UserResponse, error) {
	return s.getUserByID(ctx, userID)
}

func (s *userService) GetByID(ctx context.Context, userID string) (*dto.UserResponse, error) {
	return s.getUserByID(ctx, userID)
}

func (s *userService) getUserByID(ctx context.Context, userID string) (*dto.UserResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.BadRequest("invalid user ID")
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *userService) UpdateMe(
	ctx context.Context,
	userID string,
	req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.getUserForUpdate(ctx, userID)
	if err != nil {
		return nil, err
	}

	s.applyUserUpdates(user, req)

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to update user", zap.String("userId", userID), zap.Error(err))
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *userService) UpdateEmail(
	ctx context.Context,
	userID string,
	req dto.UpdateEmailRequest) (*dto.UserResponse, error) {
	user, err := s.getUserForUpdate(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.Email = req.Email
	user.IsEmailVerified = false

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to update email", zap.String("userId", userID), zap.Error(err))
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *userService) UpdatePhone(
	ctx context.Context,
	userID string,
	req dto.UpdatePhoneRequest) (*dto.UserResponse, error) {
	user, err := s.getUserForUpdate(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.PhoneNumber = req.PhoneNumber
	user.IsPhoneVerified = false

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to update phone", zap.String("userId", userID), zap.Error(err))
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *userService) UpdatePassword(
	ctx context.Context,
	userID string,
	req dto.UpdatePasswordRequest) error {
	user, err := s.getUserForUpdate(ctx, userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return apperrors.Unauthorized("invalid current password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", zap.String("userId", userID), zap.Error(err))
		return apperrors.Internal("failed to hash password", err)
	}

	user.PasswordHash = string(hashedPassword)

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("failed to update password", zap.String("userId", userID), zap.Error(err))
		return err
	}

	return nil
}

func (s *userService) DeleteMe(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.BadRequest("invalid user ID")
	}

	return s.userRepo.Delete(ctx, id)
}

func (s *userService) getUserForUpdate(ctx context.Context, userID string) (*model.User, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.BadRequest("invalid user ID")
	}

	return s.userRepo.GetByID(ctx, id)
}

func (s *userService) applyUserUpdates(user *model.User, req dto.UpdateUserRequest) {
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.PhoneNumber != nil {
		user.PhoneNumber = *req.PhoneNumber
		user.IsPhoneVerified = false
	}
	if req.DateOfBirth != nil {
		user.DateOfBirth = *req.DateOfBirth
	}
	if req.Address != nil {
		user.Address = *req.Address
	}
	if req.City != nil {
		user.City = *req.City
	}
	if req.State != nil {
		user.State = *req.State
	}
	if req.ZipCode != nil {
		user.ZipCode = *req.ZipCode
	}
}
