package handlers

import (
	"context"
	"time"

	"github.com/yash-gadgil/glyph/pkg/logger"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	ordrpb "github.com/yash-gadgil/glyph/services/gen/golang/order"
	obpb "github.com/yash-gadgil/glyph/services/gen/golang/order_book"
	"go.uber.org/zap"
)

const (
	pricePollInterval = 30 * time.Second
	sweepCheckEvery   = time.Minute
)

func (h *OrderHandler) RunPriceFeed(ctx context.Context, mrktClient mrktpb.MrktdataServiceClient) {
	if mrktClient == nil {
		h.log.Warn("price_feed_disabled_no_mrktdata_client")
		return
	}

	ticker := time.NewTicker(pricePollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !IsMarketOpen(time.Now()) {
				continue
			}
			h.pollAndInject(ctx, mrktClient)
		}
	}
}

func (h *OrderHandler) pollAndInject(ctx context.Context, mrktClient mrktpb.MrktdataServiceClient) {
	callCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	symbols, err := h.q.GetOpenOrderSymbols(callCtx)
	if err != nil {
		h.log.Warn("price_feed_open_symbols_failed", zap.Error(err))
		return
	}
	if len(symbols) == 0 {
		return
	}

	prices, err := mrktClient.GetLatestPrices(callCtx, &mrktpb.LatestPricesRequest{Symbols: symbols})
	if err != nil {
		h.log.Warn("price_feed_latest_prices_failed", zap.Error(err))
		return
	}

	for _, p := range prices.Prices {
		if p.PriceCents <= 0 {
			continue
		}
		if _, err := h.ob.InjectPrice(callCtx, &obpb.InjectPriceRequest{
			Symbol:     p.Symbol,
			PriceCents: p.PriceCents,
		}); err != nil {
			h.log.Warn("price_feed_inject_failed", logger.KV("symbol", p.Symbol), zap.Error(err))
		}
	}
}

func (h *OrderHandler) RunDayOrderSweeper(ctx context.Context) {
	ticker := time.NewTicker(sweepCheckEvery)
	defer ticker.Stop()

	var lastSweepDay string

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			if tradingAlwaysOpen() || !inCloseSweepWindow(now) {
				continue
			}
			day := now.In(nyLocation).Format("2006-01-02")
			if day == lastSweepDay {
				continue
			}
			lastSweepDay = day
			h.sweepDayOrders(ctx)
		}
	}
}

func (h *OrderHandler) sweepDayOrders(ctx context.Context) {
	callCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	orders, err := h.q.GetOpenDayOrders(callCtx)
	if err != nil {
		h.log.Warn("day_sweeper_query_failed", zap.Error(err))
		return
	}

	for _, order := range orders {
		if _, err := h.CancelOrder(callCtx, &ordrpb.CancelOrderRequest{
			OrderId: order.ID.String(),
			UserId:  order.UserID.String(),
		}); err != nil {
			h.log.Warn("day_sweeper_cancel_failed", logger.KV("order_id", order.ID.String()), zap.Error(err))
		}
	}
	if len(orders) > 0 {
		h.log.Info("day_orders_expired", zap.Int("count", len(orders)))
	}
}
