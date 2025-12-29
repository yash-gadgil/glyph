package types

import "time"

type FillEvent struct {
	TradeID        string    `json:"trade_id"`
	Symbol         string    `json:"symbol"`
	OrderID        string    `json:"order_id"`
	CounterOrderID string    `json:"counter_order_id"`
	UserID         string    `json:"user_id"`
	Side           int16     `json:"side"`
	Qty            int64     `json:"qty"`
	PriceCents     int64     `json:"price_cents"`
	Liquidity      string    `json:"liquidity"`
	ExecutedAt     time.Time `json:"executed_at"`
}

type DoneEvent struct {
	OrderID     string `json:"order_id"`
	UserID      string `json:"user_id"`
	Reason      string `json:"reason"`
	UnfilledQty int64  `json:"unfilled_qty"`
}
