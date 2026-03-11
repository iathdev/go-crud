package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"learning-go/internal/auth/application/port"
	sharederror "learning-go/internal/shared/error"
	"time"

	"github.com/redis/go-redis/v9"
)

type RefreshTokenRepository struct {
	client *redis.Client
}

func NewRefreshTokenRepository(client *redis.Client) port.RefreshTokenStorePort {
	return &RefreshTokenRepository{client: client}
}

func (repo *RefreshTokenRepository) Save(ctx context.Context, userID string, token string, expiry time.Duration) error {
	hash := hashToken(token)
	pipe := repo.client.Pipeline()
	pipe.Set(ctx, tokenKey(hash), userID, expiry)
	pipe.SAdd(ctx, userTokensKey(userID), hash)
	pipe.Expire(ctx, userTokensKey(userID), expiry)
	_, err := pipe.Exec(ctx)
	return err
}

func (repo *RefreshTokenRepository) Find(ctx context.Context, token string) (string, error) {
	hash := hashToken(token)
	userID, err := repo.client.Get(ctx, tokenKey(hash)).Result()
	if errors.Is(err, redis.Nil) {
		return "", sharederror.ErrNotFound
	}
	return userID, err
}

func (repo *RefreshTokenRepository) Delete(ctx context.Context, token string) error {
	hash := hashToken(token)
	userID, err := repo.client.Get(ctx, tokenKey(hash)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}
	pipe := repo.client.Pipeline()
	pipe.Del(ctx, tokenKey(hash))
	if userID != "" {
		pipe.SRem(ctx, userTokensKey(userID), hash)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (repo *RefreshTokenRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	hashes, err := repo.client.SMembers(ctx, userTokensKey(userID)).Result()
	if err != nil {
		return err
	}
	if len(hashes) == 0 {
		return nil
	}
	keys := make([]string, 0, len(hashes)+1)
	for _, h := range hashes {
		keys = append(keys, tokenKey(h))
	}
	keys = append(keys, userTokensKey(userID))
	return repo.client.Del(ctx, keys...).Err()
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func tokenKey(hash string) string {
	return "refresh_token:" + hash
}

func userTokensKey(userID string) string {
	return "user_tokens:" + userID
}
