package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yash-gadgil/glyph/pkg/telemetry"
	db "github.com/yash-gadgil/glyph/services/user/db/gen"
	"github.com/yash-gadgil/glyph/services/user/handlers"
	"github.com/yash-gadgil/glyph/services/user/types"
	"go.uber.org/zap"
)

const (
	ExchangeName    = "order.events"
	DLXName         = "order.events.dlx"
	FillsQueueName  = "user-svc.fills"
	DoneQueueName   = "user-svc.done"
	FillRoutingKey  = "order.fill"
	DoneRoutingKey  = "order.done"
	dlqSuffix       = ".dlq"
	consumerTagBase = "user-svc-settlement"
)

type Settler struct {
	db  *sql.DB
	q   *db.Queries
	log *zap.Logger
}

func NewSettler(sdb *sql.DB, log *zap.Logger) *Settler {
	return &Settler{db: sdb, q: db.New(sdb), log: log}
}

func (s *Settler) SettleFill(ctx context.Context, ev types.FillEvent) error {
	tradeID, err := uuid.Parse(ev.TradeID)
	if err != nil {
		return fmt.Errorf("bad trade_id %q: %w", ev.TradeID, err)
	}
	orderID, err := uuid.Parse(ev.OrderID)
	if err != nil {
		return fmt.Errorf("bad order_id %q: %w", ev.OrderID, err)
	}
	userID, err := uuid.Parse(ev.UserID)
	if err != nil {
		return fmt.Errorf("bad user_id %q: %w", ev.UserID, err)
	}
	if ev.Qty <= 0 || ev.PriceCents < 0 {
		return fmt.Errorf("invalid fill payload: qty=%d price=%d", ev.Qty, ev.PriceCents)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback()

	qtx := s.q.WithTx(tx)

	var pos handlers.PositionState
	row, err := qtx.GetPositionForUpdate(ctx, db.GetPositionForUpdateParams{
		UserID: userID,
		Symbol: ev.Symbol,
	})
	switch {
	case err == nil:
		pos = handlers.PositionState{Qty: row.Qty, CostBasis: row.CostBasis, RealizedPnl: row.RealizedPnl}
	case errors.Is(err, sql.ErrNoRows):
	default:
		return fmt.Errorf("load position: %w", err)
	}

	outcome, err := handlers.ApplyFill(pos, handlers.SettledFill{
		Side:       ev.Side,
		Qty:        ev.Qty,
		PriceCents: ev.PriceCents,
	})
	if err != nil {
		return fmt.Errorf("apply fill: %w", err)
	}

	inserted, err := qtx.InsertSettlement(ctx, db.InsertSettlementParams{
		TradeID:          tradeID,
		OrderID:          orderID,
		UserID:           userID,
		Symbol:           ev.Symbol,
		Side:             ev.Side,
		Qty:              ev.Qty,
		PriceCents:       ev.PriceCents,
		CashDeltaCents:   outcome.CashDeltaCents,
		RealizedPnlCents: outcome.RealizedCents,
	})
	if err != nil {
		return fmt.Errorf("insert settlement: %w", err)
	}
	if inserted == 0 {
		s.log.Info("fill_already_settled",
			zap.String("trade_id", ev.TradeID),
			zap.String("order_id", ev.OrderID),
		)
		return tx.Commit()
	}

	if err := qtx.SetPosition(ctx, db.SetPositionParams{
		UserID:      userID,
		Symbol:      ev.Symbol,
		Qty:         outcome.Position.Qty,
		RealizedPnl: outcome.Position.RealizedPnl,
		CostBasis:   outcome.Position.CostBasis,
	}); err != nil {
		return fmt.Errorf("set position: %w", err)
	}

	var holdRelease int64
	res, err := qtx.GetReservationForUpdate(ctx, orderID)
	switch {
	case err == nil:
		if ev.Side == handlers.SideBuy {
			holdRelease = ev.Qty * res.CentsPerShare
		}
		if _, err := qtx.ReduceReservation(ctx, db.ReduceReservationParams{
			OrderID:      orderID,
			RemainingQty: ev.Qty,
		}); err != nil {
			return fmt.Errorf("reduce reservation: %w", err)
		}
	case errors.Is(err, sql.ErrNoRows):
	default:
		return fmt.Errorf("load reservation: %w", err)
	}

	if err := qtx.ApplyCashDelta(ctx, db.ApplyCashDeltaParams{
		UserID:       userID,
		CashBalance:  outcome.CashDeltaCents,
		ReservedCash: holdRelease,
	}); err != nil {
		return fmt.Errorf("apply cash delta: %w", err)
	}

	if ev.Side == handlers.SideSell {
		if err := qtx.ReleaseShares(ctx, db.ReleaseSharesParams{
			UserID:      userID,
			Symbol:      ev.Symbol,
			ReservedQty: ev.Qty,
		}); err != nil {
			return fmt.Errorf("release shares: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	telemetry.FillsAppliedTotal.WithLabelValues(telemetry.SideLabel(int32(ev.Side))).Inc()
	if !ev.ExecutedAt.IsZero() {
		telemetry.FillApplyLagSeconds.Observe(time.Since(ev.ExecutedAt).Seconds())
	}

	s.log.Info("fill_settled",
		zap.String("trade_id", ev.TradeID),
		zap.String("order_id", ev.OrderID),
		zap.String("symbol", ev.Symbol),
		zap.Int64("qty", ev.Qty),
		zap.Int64("price_cents", ev.PriceCents),
		zap.Int64("cash_delta_cents", outcome.CashDeltaCents),
	)
	return nil
}

func (s *Settler) HandleDone(ctx context.Context, ev types.DoneEvent) error {
	orderID, err := uuid.Parse(ev.OrderID)
	if err != nil {
		return fmt.Errorf("bad order_id %q: %w", ev.OrderID, err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback()

	qtx := s.q.WithTx(tx)

	res, err := qtx.GetReservationForUpdate(ctx, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tx.Commit()
		}
		return fmt.Errorf("load reservation: %w", err)
	}

	if res.RemainingQty > 0 {
		switch res.Side {
		case handlers.SideBuy:
			if err := qtx.ReleaseCash(ctx, db.ReleaseCashParams{
				UserID:       res.UserID,
				ReservedCash: res.RemainingQty * res.CentsPerShare,
			}); err != nil {
				return fmt.Errorf("release cash: %w", err)
			}
		case handlers.SideSell:
			if err := qtx.ReleaseShares(ctx, db.ReleaseSharesParams{
				UserID:      res.UserID,
				Symbol:      res.Symbol,
				ReservedQty: res.RemainingQty,
			}); err != nil {
				return fmt.Errorf("release shares: %w", err)
			}
		}
	}

	if err := qtx.DeleteReservation(ctx, orderID); err != nil {
		return fmt.Errorf("delete reservation: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	s.log.Info("reservation_released",
		zap.String("order_id", ev.OrderID),
		zap.String("reason", ev.Reason),
	)
	return nil
}

func StartSettlementConsumer(ctx context.Context, ch *amqp.Channel, settler *Settler, log *zap.Logger) error {
	if err := DeclareTopology(ch); err != nil {
		return err
	}

	fills, err := ch.Consume(FillsQueueName, consumerTagBase+"-fills", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume fills: %w", err)
	}
	done, err := ch.Consume(DoneQueueName, consumerTagBase+"-done", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume done: %w", err)
	}

	go consumeLoop(ctx, fills, log, func(body []byte) error {
		var ev types.FillEvent
		if err := json.Unmarshal(body, &ev); err != nil {
			return permanent(fmt.Errorf("bad fill payload: %w", err))
		}
		if err := settler.SettleFill(ctx, ev); err != nil {
			if isPermanentSettleError(err) {
				return permanent(err)
			}
			return err
		}
		return nil
	})

	go consumeLoop(ctx, done, log, func(body []byte) error {
		var ev types.DoneEvent
		if err := json.Unmarshal(body, &ev); err != nil {
			return permanent(fmt.Errorf("bad done payload: %w", err))
		}
		if err := settler.HandleDone(ctx, ev); err != nil {
			if isPermanentSettleError(err) {
				return permanent(err)
			}
			return err
		}
		return nil
	})

	log.Info("settlement_consumer_started")
	return nil
}

func DeclareTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(ExchangeName, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}
	if err := ch.ExchangeDeclare(DLXName, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare dlx: %w", err)
	}

	for queue, rk := range map[string]string{
		FillsQueueName: FillRoutingKey,
		DoneQueueName:  DoneRoutingKey,
	} {
		dlq := queue + dlqSuffix
		if _, err := ch.QueueDeclare(dlq, true, false, false, false, nil); err != nil {
			return fmt.Errorf("declare %s: %w", dlq, err)
		}
		if err := ch.QueueBind(dlq, rk, DLXName, false, nil); err != nil {
			return fmt.Errorf("bind %s: %w", dlq, err)
		}
		if _, err := ch.QueueDeclare(queue, true, false, false, false, amqp.Table{
			"x-dead-letter-exchange": DLXName,
		}); err != nil {
			return fmt.Errorf("declare %s: %w", queue, err)
		}
		if err := ch.QueueBind(queue, rk, ExchangeName, false, nil); err != nil {
			return fmt.Errorf("bind %s: %w", queue, err)
		}
	}
	return nil
}

type permanentError struct{ error }

func permanent(err error) error { return permanentError{err} }

func isPermanentSettleError(err error) bool {
	var pe permanentError
	return errors.As(err, &pe)
}

func consumeLoop(ctx context.Context, msgs <-chan amqp.Delivery, log *zap.Logger, handle func([]byte) error) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			err := handle(msg.Body)
			switch {
			case err == nil:
				_ = msg.Ack(false)
			case errors.As(err, &permanentError{}):
				log.Error("event_dead_lettered", zap.Error(err), zap.ByteString("body", msg.Body))
				_ = msg.Nack(false, false)
			default:
				log.Warn("event_requeued", zap.Error(err))
				_ = msg.Nack(false, true)
			}
		}
	}
}
