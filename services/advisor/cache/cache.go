package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type PosPrint struct {
	Qty            int64 `json:"qty"`
	CostBasisCents int64 `json:"cost_basis_cents"`
	AvgPriceCents  int64 `json:"avg_price_cents"`
}

type Fingerprint struct {
	CashCents   int64               `json:"cash_cents"`
	EquityCents int64               `json:"equity_cents"`
	Positions   map[string]PosPrint `json:"positions"`
}

type Entry struct {
	Snapshot    string      `json:"snapshot"`
	Analysis    string      `json:"analysis"`
	Fingerprint Fingerprint `json:"fingerprint"`
	GeneratedAt time.Time   `json:"generated_at"`
}

type Cache struct {
	rdb *redis.Client
	log *zap.Logger
}

func Init(ctx context.Context, log *zap.Logger) *Cache {
	addr := os.Getenv("CACHE_ADDR")
	if addr == "" {
		log.Warn("advisor_cache_disabled", zap.String("reason", "CACHE_ADDR not set"))
		return &Cache{log: log}
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("CACHE_PASSWORD"),
		DB:       1,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Warn("advisor_cache_unreachable", zap.Error(err))
		return &Cache{log: log}
	}

	log.Info("advisor_cache_connected", zap.String("addr", addr))
	return &Cache{rdb: rdb, log: log}
}

func (c *Cache) Enabled() bool {
	return c.rdb != nil
}

func key(userID string) string {
	return fmt.Sprintf("advisor:analysis:%s", userID)
}

func (c *Cache) Get(ctx context.Context, userID string) (*Entry, error) {
	if c.rdb == nil {
		return nil, nil
	}
	raw, err := c.rdb.Get(ctx, key(userID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var entry Entry
	if err := json.Unmarshal(raw, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func (c *Cache) Set(ctx context.Context, userID string, entry *Entry) error {
	if c.rdb == nil {
		return nil
	}
	payload, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key(userID), payload, 7*24*time.Hour).Err()
}
