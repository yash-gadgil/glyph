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

type BacktestSummary struct {
	TotalReturnPct float64 `json:"total_return_pct"`
	MaxDrawdownPct float64 `json:"max_drawdown_pct"`
	Sharpe         float64 `json:"sharpe"`
	WinRate        float64 `json:"win_rate"`
	ProfitFactor   float64 `json:"profit_factor"`
	NumTrades      int32   `json:"num_trades"`
}

type StratJob struct {
	State      string           `json:"state"`
	Name       string           `json:"name"`
	ConfigJSON string           `json:"config_json"`
	Rationale  string           `json:"rationale"`
	Backtest   *BacktestSummary `json:"backtest,omitempty"`
	Error      string           `json:"error"`
	StartedAt  time.Time        `json:"started_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

func stratJobKey(userID string) string {
	return fmt.Sprintf("advisor:stratjob:%s", userID)
}

func stratLockKey(userID string) string {
	return fmt.Sprintf("advisor:stratjoblock:%s", userID)
}

func (c *Cache) GetJob(ctx context.Context, userID string) (*StratJob, error) {
	if c.rdb == nil {
		return nil, nil
	}
	raw, err := c.rdb.Get(ctx, stratJobKey(userID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var job StratJob
	if err := json.Unmarshal(raw, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (c *Cache) SetJob(ctx context.Context, userID string, job *StratJob) error {
	if c.rdb == nil {
		return nil
	}
	job.UpdatedAt = time.Now().UTC()
	payload, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, stratJobKey(userID), payload, 24*time.Hour).Err()
}

func (c *Cache) AcquireJobLock(ctx context.Context, userID string, ttl time.Duration) (bool, error) {
	if c.rdb == nil {
		return true, nil
	}
	return c.rdb.SetNX(ctx, stratLockKey(userID), "1", ttl).Result()
}

func (c *Cache) ReleaseJobLock(ctx context.Context, userID string) {
	if c.rdb == nil {
		return
	}
	c.rdb.Del(ctx, stratLockKey(userID))
}

type ChatTurn struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatSession struct {
	Turns     []ChatTurn `json:"turns"`
	InFlight  bool       `json:"in_flight"`
	Partial   string     `json:"partial"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func chatKey(userID string) string {
	return fmt.Sprintf("advisor:chat:%s", userID)
}

func chatLockKey(userID string) string {
	return fmt.Sprintf("advisor:chatlock:%s", userID)
}

func (c *Cache) AcquireChatLock(ctx context.Context, userID string, ttl time.Duration) (bool, error) {
	if c.rdb == nil {
		return true, nil
	}
	return c.rdb.SetNX(ctx, chatLockKey(userID), "1", ttl).Result()
}

func (c *Cache) ReleaseChatLock(ctx context.Context, userID string) {
	if c.rdb == nil {
		return
	}
	c.rdb.Del(ctx, chatLockKey(userID))
}

func (c *Cache) GetChatSession(ctx context.Context, userID string) (*ChatSession, error) {
	if c.rdb == nil {
		return nil, nil
	}
	raw, err := c.rdb.Get(ctx, chatKey(userID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var session ChatSession
	if err := json.Unmarshal(raw, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (c *Cache) SetChatSession(ctx context.Context, userID string, session *ChatSession) error {
	if c.rdb == nil {
		return nil
	}
	session.UpdatedAt = time.Now().UTC()
	payload, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, chatKey(userID), payload, 45*time.Minute).Err()
}

func (c *Cache) DeleteChatSession(ctx context.Context, userID string) {
	if c.rdb == nil {
		return
	}
	c.rdb.Del(ctx, chatKey(userID))
}
