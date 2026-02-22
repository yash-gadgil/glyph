package db

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Cache struct {
	Rdb *redis.Client
	log *zap.Logger
}

type PendingSignupData struct {
	Name         string `json:"name"`
	PasswordHash string `json:"password_hash"`
}

func InitCache(ctx context.Context, logger *zap.Logger) *Cache {
	addr := os.Getenv("CACHE_ADDR")
	password := os.Getenv("CACHE_PASSWORD")

	if addr == "" {
		logger.Warn("redis_disabled", zap.String("reason", "CACHE_ADDR not set"))
		return &Cache{Rdb: nil, log: logger}
	}

	Rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	if err := Rdb.Ping(ctx).Err(); err != nil {
		logger.Warn("redis_unreachable_continuing_without", zap.Error(err))
		return &Cache{Rdb: nil, log: logger}
	}

	logger.Info("redis_connected", zap.String("addr", addr))
	return &Cache{Rdb: Rdb, log: logger}
}

func (c *Cache) StorePendingSignup(ctx context.Context, email, name, passwordHash string, ttl time.Duration) error {
	if c.Rdb == nil {
		return fmt.Errorf("redis not available")
	}

	data := PendingSignupData{Name: name, PasswordHash: passwordHash}
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal pending signup data: %w", err)
	}

	key := fmt.Sprintf("pending_signup:%s", email)
	return c.Rdb.Set(ctx, key, payload, ttl).Err()
}

func (c *Cache) GetPendingSignup(ctx context.Context, email string) (*PendingSignupData, error) {
	if c.Rdb == nil {
		return nil, fmt.Errorf("redis not available")
	}

	key := fmt.Sprintf("pending_signup:%s", email)
	raw, err := c.Rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("signup not found or expired")
	}
	if err != nil {
		return nil, err
	}

	var data PendingSignupData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pending signup data: %w", err)
	}
	return &data, nil
}

func (c *Cache) DeletePendingSignup(ctx context.Context, email string) error {
	if c.Rdb == nil {
		return nil
	}

	key := fmt.Sprintf("pending_signup:%s", email)
	return c.Rdb.Del(ctx, key).Err()
}

func (c *Cache) EnqueueVerificationEmail(ctx context.Context, email, token string) error {
	if c.Rdb == nil {
		return fmt.Errorf("redis not available")
	}

	payload := fmt.Sprintf(`{"email":"%s","token":"%s"}`, email, token)
	return c.Rdb.RPush(ctx, "email_verification_queue", payload).Err()
}

func (c *Cache) DequeueVerificationEmail(ctx context.Context) (string, error) {
	if c.Rdb == nil {
		return "", fmt.Errorf("redis not available")
	}

	result, err := c.Rdb.BLPop(ctx, 5*time.Second, "email_verification_queue").Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	if len(result) < 2 {
		return "", nil
	}
	return result[1], nil
}

func (c *Cache) MoveVerificationToDLQ(ctx context.Context, payload string, errMsg string) error {
	if c.Rdb == nil {
		return nil
	}

	dlqPayload := fmt.Sprintf(`{"payload":%s,"error":"%s","timestamp":"%s"}`,
		payload, errMsg, time.Now().Format(time.RFC3339))
	return c.Rdb.RPush(ctx, "email_verification_dlq", dlqPayload).Err()
}

func (c *Cache) EnqueuePasswordResetEmail(ctx context.Context, email, token string) error {
	if c.Rdb == nil {
		return fmt.Errorf("redis not available")
	}

	payload := fmt.Sprintf(`{"email":"%s","token":"%s"}`, email, token)
	return c.Rdb.RPush(ctx, "password_reset_queue", payload).Err()
}

func (c *Cache) DequeuePasswordResetEmail(ctx context.Context) (string, error) {
	if c.Rdb == nil {
		return "", fmt.Errorf("redis not available")
	}

	result, err := c.Rdb.BLPop(ctx, 5*time.Second, "password_reset_queue").Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	if len(result) < 2 {
		return "", nil
	}
	return result[1], nil
}
