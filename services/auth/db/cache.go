package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	Rdb *redis.Client
}

func InitCache(ctx context.Context) *Cache {
	addr := os.Getenv("CACHE_ADDR")
	password := os.Getenv("CACHE_PASSWORD")

	if addr == "" {
		log.Println("CACHE_ADDR not set, skipping Redis connection") // LOG
		return &Cache{
			Rdb: nil,
		}
	}

	Rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	if err := Rdb.Ping(ctx).Err(); err != nil {
		log.Printf("Failed to connect to Redis: %v (continuing without Redis)", err) // LOG
		return &Cache{
			Rdb: nil,
		}
	}

	log.Println("Connected to Redis") // LOG
	return &Cache{
		Rdb: Rdb,
	}
}

func (c *Cache) StorePendingRegistration(ctx context.Context, email, passwordHash string, ttl time.Duration) error {
	if c.Rdb == nil {
		return fmt.Errorf("redis not available")
	}

	key := fmt.Sprintf("pending_registration:%s", email)
	return c.Rdb.Set(ctx, key, passwordHash, ttl).Err()
}

func (c *Cache) GetPendingRegistration(ctx context.Context, email string) (string, error) {
	if c.Rdb == nil {
		return "", fmt.Errorf("redis not available")
	}

	key := fmt.Sprintf("pending_registration:%s", email)
	passwordHash, err := c.Rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("registration not found or expired")
	}
	return passwordHash, err
}

func (c *Cache) DeletePendingRegistration(ctx context.Context, email string) error {
	if c.Rdb == nil {
		return nil
	}

	key := fmt.Sprintf("pending_registration:%s", email)
	return c.Rdb.Del(ctx, key).Err()
}
