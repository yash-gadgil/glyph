package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/pkg/telemetry"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	db "github.com/yash-gadgil/glyph/services/user/db/gen"
	"github.com/yash-gadgil/glyph/services/user/strategyengine"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const maxStrategyConfigBytes = 64 * 1024

const minPositionSizeCents = 100

func deploymentStatus(status int16) string {
	if status == 0 {
		return "running"
	}
	return "stopped"
}

func deploymentToProto(id, strategyID, userID uuid.UUID, symbol string, sizeCents int64, status int16, inPosition bool, entryCents, qty int64, name string, createdAt, updatedAt time.Time) *userpb.Deployment {
	return &userpb.Deployment{
		Id:                id.String(),
		StrategyId:        strategyID.String(),
		UserId:            userID.String(),
		Symbol:            symbol,
		PositionSizeCents: sizeCents,
		Status:            deploymentStatus(status),
		InPosition:        inPosition,
		EntryPriceCents:   entryCents,
		Qty:               qty,
		StrategyName:      name,
		CreatedAt:         createdAt.Format(time.RFC3339),
		UpdatedAt:         updatedAt.Format(time.RFC3339),
	}
}

type StrategyHandler struct {
	userpb.UnimplementedStrategyServiceServer
	db   *sql.DB
	q    *db.Queries
	mrkt mrktpb.MrktdataServiceClient
	log  *zap.Logger
}

func NewStrategyHandler(sdb *sql.DB, mrkt mrktpb.MrktdataServiceClient, log *zap.Logger) *StrategyHandler {
	return &StrategyHandler{db: sdb, q: db.New(sdb), mrkt: mrkt, log: log}
}

func validateStrategyInput(name, configJSON string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return status.Errorf(codes.InvalidArgument, "strategy name is required")
	}
	if len(name) > 80 {
		return status.Errorf(codes.InvalidArgument, "strategy name must be 80 characters or fewer")
	}
	if len(configJSON) > maxStrategyConfigBytes {
		return status.Errorf(codes.InvalidArgument, "strategy config too large")
	}
	if !json.Valid([]byte(configJSON)) {
		return status.Errorf(codes.InvalidArgument, "strategy config must be valid JSON")
	}
	return nil
}

func strategyToProto(s db.Strategy) *userpb.Strategy {
	return &userpb.Strategy{
		Id:         s.ID.String(),
		UserId:     s.UserID.String(),
		Name:       s.Name,
		ConfigJson: string(s.Config),
		CreatedAt:  s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  s.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *StrategyHandler) GetStrategies(ctx context.Context, req *userpb.UserSpecifier) (*userpb.StrategiesResponse, error) {
	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	rows, err := s.q.GetStrategiesForUser(ctx, userUUID)
	if err != nil {
		s.log.Error("strategies_fetch_failed", logger.Action("get_strategies"), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to load strategies")
	}

	resp := &userpb.StrategiesResponse{}
	for _, row := range rows {
		resp.Strategies = append(resp.Strategies, strategyToProto(row))
	}
	return resp, nil
}

func (s *StrategyHandler) CreateStrategy(ctx context.Context, req *userpb.CreateStrategyRequest) (*userpb.Strategy, error) {
	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}
	if err := validateStrategyInput(req.Name, req.ConfigJson); err != nil {
		return nil, err
	}

	row, err := s.q.CreateStrategy(ctx, db.CreateStrategyParams{
		UserID: userUUID,
		Name:   strings.TrimSpace(req.Name),
		Config: json.RawMessage(req.ConfigJson),
	})
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return nil, status.Errorf(codes.AlreadyExists, "a strategy named %q already exists", req.Name)
		}
		s.log.Error("strategy_create_failed", logger.Action("create_strategy"), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create strategy")
	}

	return strategyToProto(row), nil
}

func (s *StrategyHandler) UpdateStrategy(ctx context.Context, req *userpb.UpdateStrategyRequest) (*userpb.Strategy, error) {
	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}
	strategyUUID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid strategy ID")
	}
	if err := validateStrategyInput(req.Name, req.ConfigJson); err != nil {
		return nil, err
	}

	row, err := s.q.UpdateStrategy(ctx, db.UpdateStrategyParams{
		ID:     strategyUUID,
		UserID: userUUID,
		Name:   strings.TrimSpace(req.Name),
		Config: json.RawMessage(req.ConfigJson),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "strategy not found")
		}
		s.log.Error("strategy_update_failed", logger.Action("update_strategy"), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update strategy")
	}

	return strategyToProto(row), nil
}

func (s *StrategyHandler) DeleteStrategy(ctx context.Context, req *userpb.StrategySpecifier) (*emptypb.Empty, error) {
	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}
	strategyUUID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid strategy ID")
	}

	if err := s.q.DeleteStrategy(ctx, db.DeleteStrategyParams{
		ID:     strategyUUID,
		UserID: userUUID,
	}); err != nil {
		s.log.Error("strategy_delete_failed", logger.Action("delete_strategy"), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to delete strategy")
	}

	return &emptypb.Empty{}, nil
}

func (s *StrategyHandler) DeployStrategy(ctx context.Context, req *userpb.DeployStrategyRequest) (*userpb.Deployment, error) {
	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}
	strategyUUID, err := uuid.Parse(req.StrategyId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid strategy ID")
	}
	symbol := strings.ToUpper(strings.TrimSpace(req.Symbol))
	if symbol == "" || len(symbol) > 10 {
		return nil, status.Errorf(codes.InvalidArgument, "a valid symbol is required")
	}
	if req.PositionSizeCents < minPositionSizeCents {
		return nil, status.Errorf(codes.InvalidArgument, "position size must be at least $1")
	}

	strat, err := s.q.GetStrategy(ctx, db.GetStrategyParams{ID: strategyUUID, UserID: userUUID})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "strategy not found")
	}
	if _, err := strategyengine.ParseConfig(strat.Config); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "strategy has no runnable entry rules, edit it first")
	}

	existing, err := s.q.GetLatestDeploymentForSymbol(ctx, db.GetLatestDeploymentForSymbolParams{
		UserID:     userUUID,
		StrategyID: strategyUUID,
		Symbol:     symbol,
	})
	if err == nil {
		row := existing
		if existing.Status != 0 {
			row, err = s.q.ReactivateDeployment(ctx, db.ReactivateDeploymentParams{
				ID:                existing.ID,
				PositionSizeCents: req.PositionSizeCents,
			})
			if err != nil {
				s.log.Error("deployment_reactivate_failed", logger.Action("deploy_strategy"), zap.Error(err))
				return nil, status.Errorf(codes.Internal, "failed to deploy strategy")
			}
			telemetry.StrategyDeploymentsTotal.Inc()
		}
		return deploymentToProto(row.ID, row.StrategyID, row.UserID, row.Symbol, row.PositionSizeCents,
			row.Status, row.InPosition, row.EntryPriceCents, row.Qty, strat.Name, row.CreatedAt, row.UpdatedAt), nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		s.log.Error("deployment_lookup_failed", logger.Action("deploy_strategy"), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to deploy strategy")
	}

	row, err := s.q.CreateDeployment(ctx, db.CreateDeploymentParams{
		UserID:            userUUID,
		StrategyID:        strategyUUID,
		Symbol:            symbol,
		PositionSizeCents: req.PositionSizeCents,
	})
	if err != nil {
		s.log.Error("deployment_create_failed", logger.Action("deploy_strategy"), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to deploy strategy")
	}

	telemetry.StrategyDeploymentsTotal.Inc()

	return deploymentToProto(row.ID, row.StrategyID, row.UserID, row.Symbol, row.PositionSizeCents,
		row.Status, row.InPosition, row.EntryPriceCents, row.Qty, strat.Name, row.CreatedAt, row.UpdatedAt), nil
}

func (s *StrategyHandler) StopDeployment(ctx context.Context, req *userpb.DeploymentSpecifier) (*userpb.Deployment, error) {
	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}
	depUUID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid deployment ID")
	}

	row, err := s.q.StopDeployment(ctx, db.StopDeploymentParams{ID: depUUID, UserID: userUUID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "deployment not found")
		}
		s.log.Error("deployment_stop_failed", logger.Action("stop_deployment"), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to stop deployment")
	}

	return deploymentToProto(row.ID, row.StrategyID, row.UserID, row.Symbol, row.PositionSizeCents,
		row.Status, row.InPosition, row.EntryPriceCents, row.Qty, "", row.CreatedAt, row.UpdatedAt), nil
}

func (s *StrategyHandler) DeleteDeployment(ctx context.Context, req *userpb.DeploymentSpecifier) (*emptypb.Empty, error) {
	userUUID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}
	depUUID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid deployment ID")
	}

	rows, err := s.q.DeleteDeployment(ctx, db.DeleteDeploymentParams{ID: depUUID, UserID: userUUID})
	if err != nil {
		s.log.Error("deployment_delete_failed", logger.Action("delete_deployment"), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to delete deployment")
	}
	if rows == 0 {
		return nil, status.Errorf(codes.NotFound, "deployment not found or still running, stop it first")
	}

	return &emptypb.Empty{}, nil
}
