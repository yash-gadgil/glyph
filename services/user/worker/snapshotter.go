package worker

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	db "github.com/yash-gadgil/glyph/services/user/db/gen"
	"go.uber.org/zap"
)

const (
	snapshotInterval  = time.Minute
	pruneInterval     = time.Hour
	snapshotRetention = 30 * 24 * time.Hour
)

type Snapshotter struct {
	db     *sql.DB
	q      *db.Queries
	prices mrktpb.MrktdataServiceClient
	log    *zap.Logger
}

func NewSnapshotter(sdb *sql.DB, prices mrktpb.MrktdataServiceClient, log *zap.Logger) *Snapshotter {
	return &Snapshotter{db: sdb, q: db.New(sdb), prices: prices, log: log}
}

func (s *Snapshotter) Run(ctx context.Context) {
	tick := time.NewTicker(snapshotInterval)
	prune := time.NewTicker(pruneInterval)
	defer tick.Stop()
	defer prune.Stop()

	s.log.Info("snapshotter_started")

	if err := s.snapshotOnce(ctx); err != nil {
		s.log.Warn("snapshot_failed", zap.Error(err))
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if err := s.snapshotOnce(ctx); err != nil {
				s.log.Warn("snapshot_failed", zap.Error(err))
			}
		case <-prune.C:
			cutoff := time.Now().Add(-snapshotRetention)
			if n, err := s.q.PruneSnapshots(ctx, cutoff); err != nil {
				s.log.Warn("snapshot_prune_failed", zap.Error(err))
			} else if n > 0 {
				s.log.Info("snapshots_pruned", zap.Int64("rows", n))
			}
		}
	}
}

func (s *Snapshotter) snapshotOnce(ctx context.Context) error {
	accounts, err := s.q.GetAllAccounts(ctx)
	if err != nil {
		return err
	}
	if len(accounts) == 0 {
		return nil
	}

	positions, err := s.q.GetAllOpenPositions(ctx)
	if err != nil {
		return err
	}

	prices := s.latestPrices(ctx, positions)

	marketValue := make(map[uuid.UUID]int64, len(accounts))
	for _, p := range positions {
		if last, ok := prices[p.Symbol]; ok && last > 0 {
			marketValue[p.UserID] += last * p.Qty
		} else {
			marketValue[p.UserID] += p.CostBasis
		}
	}

	for _, a := range accounts {
		mv := marketValue[a.UserID]
		if err := s.q.InsertSnapshot(ctx, db.InsertSnapshotParams{
			UserID:           a.UserID,
			EquityCents:      a.CashBalance + mv,
			CashCents:        a.CashBalance,
			MarketValueCents: mv,
		}); err != nil {
			s.log.Warn("snapshot_insert_failed",
				zap.String("user_id", a.UserID.String()), zap.Error(err))
		}
	}
	return nil
}

func (s *Snapshotter) latestPrices(ctx context.Context, positions []db.GetAllOpenPositionsRow) map[string]int64 {
	out := map[string]int64{}
	if len(positions) == 0 || s.prices == nil {
		return out
	}

	seen := make(map[string]struct{}, len(positions))
	symbols := make([]string, 0, len(positions))
	for _, p := range positions {
		if _, ok := seen[p.Symbol]; !ok {
			seen[p.Symbol] = struct{}{}
			symbols = append(symbols, p.Symbol)
		}
	}

	resp, err := s.prices.GetLatestPrices(ctx, &mrktpb.LatestPricesRequest{Symbols: symbols})
	if err != nil {
		s.log.Warn("snapshot_price_fetch_failed_using_cost_basis", zap.Error(err))
		return out
	}
	for _, p := range resp.Prices {
		out[p.Symbol] = p.PriceCents
	}
	return out
}
