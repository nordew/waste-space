package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type TokenCache interface {
	SetRefreshToken(ctx context.Context, userID uuid.UUID, token string, ttl time.Duration) error
	GetRefreshToken(ctx context.Context, userID uuid.UUID) (string, error)
	DeleteRefreshToken(ctx context.Context, userID uuid.UUID) error
	BlacklistAccessToken(ctx context.Context, token string, ttl time.Duration) error
	IsAccessTokenBlacklisted(ctx context.Context, token string) (bool, error)
}

type tokenCache struct {
	client *redis.Client
}

func NewTokenCache(client *redis.Client) TokenCache {
	return &tokenCache{
		client: client,
	}
}

func (c *tokenCache) SetRefreshToken(ctx context.Context, userID uuid.UUID, token string, ttl time.Duration) error {
	key := fmt.Sprintf("refresh_token:%s", userID.String())
	return c.client.Set(ctx, key, token, ttl).Err()
}

func (c *tokenCache) GetRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	key := fmt.Sprintf("refresh_token:%s", userID.String())
	return c.client.Get(ctx, key).Result()
}

func (c *tokenCache) DeleteRefreshToken(ctx context.Context, userID uuid.UUID) error {
	key := fmt.Sprintf("refresh_token:%s", userID.String())
	return c.client.Del(ctx, key).Err()
}

func (c *tokenCache) BlacklistAccessToken(ctx context.Context, token string, ttl time.Duration) error {
	key := fmt.Sprintf("blacklist:%s", token)
	return c.client.Set(ctx, key, "1", ttl).Err()
}

func (c *tokenCache) IsAccessTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", token)
	exists, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}
