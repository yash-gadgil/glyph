package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yash-gadgil/glyph/pkg/logger"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	db "github.com/yash-gadgil/glyph/services/user/db/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxStrategyConfigBytes = 64 * 1024

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
