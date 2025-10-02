package repository

import (
	"context"
	"errors"
	"time"
	"waste-space/internal/model"
	apperrors "waste-space/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*model.User, error)
	Count(ctx context.Context) (int64, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	VerifyEmail(ctx context.Context, id uuid.UUID) error
	VerifyPhone(ctx context.Context, id uuid.UUID) error
	SetActive(ctx context.Context, id uuid.UUID, isActive bool) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return apperrors.AlreadyExists("user with this email already exists")
		}
		return apperrors.Internal("failed to create user", result.Error)
	}

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("user not found")
		}
		return nil, apperrors.Internal("failed to get user", result.Error)
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("user not found")
		}
		return nil, apperrors.Internal("failed to get user", result.Error)
	}

	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		return apperrors.Internal("failed to update user", result.Error)
	}

	if result.RowsAffected == 0 {
		return apperrors.NotFound("user not found")
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.User{}, id)
	if result.Error != nil {
		return apperrors.Internal("failed to delete user", result.Error)
	}

	if result.RowsAffected == 0 {
		return apperrors.NotFound("user not found")
	}

	return nil
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*model.User, error) {
	var users []*model.User
	result := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users)
	if result.Error != nil {
		return nil, apperrors.Internal("failed to list users", result.Error)
	}

	return users, nil
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&model.User{}).Count(&count)
	if result.Error != nil {
		return 0, apperrors.Internal("failed to count users", result.Error)
	}

	return count, nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Update("last_login_at", now)
	if result.Error != nil {
		return apperrors.Internal("failed to update last login", result.Error)
	}

	if result.RowsAffected == 0 {
		return apperrors.NotFound("user not found")
	}

	return nil
}

func (r *userRepository) VerifyEmail(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Update("is_email_verified", true)
	if result.Error != nil {
		return apperrors.Internal("failed to verify email", result.Error)
	}

	if result.RowsAffected == 0 {
		return apperrors.NotFound("user not found")
	}

	return nil
}

func (r *userRepository) VerifyPhone(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Update("is_phone_verified", true)
	if result.Error != nil {
		return apperrors.Internal("failed to verify phone", result.Error)
	}

	if result.RowsAffected == 0 {
		return apperrors.NotFound("user not found")
	}

	return nil
}

func (r *userRepository) SetActive(ctx context.Context, id uuid.UUID, isActive bool) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Update("is_active", isActive)
	if result.Error != nil {
		return apperrors.Internal("failed to set active status", result.Error)
	}

	if result.RowsAffected == 0 {
		return apperrors.NotFound("user not found")
	}

	return nil
}