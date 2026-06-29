package handlers

import (
	"strings"
	"time"

	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/services/advisor/cache"
	"github.com/yash-gadgil/glyph/services/advisor/types"
	advisorpb "github.com/yash-gadgil/glyph/services/gen/golang/advisor"
	strategypb "github.com/yash-gadgil/glyph/services/gen/golang/strategy"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AdvisorHandler struct {
	advisorpb.UnimplementedAdvisorServiceServer
	portfolio userpb.PortfolioServiceClient
	strategy  strategypb.StrategyServiceClient
	llm       types.Provider
	cache     *cache.Cache
	log       *zap.Logger
}

func NewAdvisorHandler(portfolio userpb.PortfolioServiceClient, strategy strategypb.StrategyServiceClient, model types.Provider, c *cache.Cache, log *zap.Logger) *AdvisorHandler {
	return &AdvisorHandler{portfolio: portfolio, strategy: strategy, llm: model, cache: c, log: log}
}

func (h *AdvisorHandler) AnalyzePortfolio(req *advisorpb.AnalyzeRequest, stream advisorpb.AdvisorService_AnalyzePortfolioServer) error {
	ctx := stream.Context()
	log := logger.WithContextFields(ctx, h.log).With(logger.Action("analyze_portfolio"))

	if req.UserId == "" {
		return status.Error(codes.InvalidArgument, "user id is required")
	}
	if h.llm == nil {
		return status.Error(codes.Unavailable, "inference service unavailable")
	}

	snapshot, fp, err := buildSnapshot(ctx, h.portfolio, req.UserId)
	if err != nil {
		log.Error("build_snapshot_failed", zap.Error(err))
		return status.Error(codes.Internal, "could not load portfolio")
	}

	if h.cache != nil {
		prev, err := h.cache.Get(ctx, req.UserId)
		if err != nil {
			log.Warn("cache_read_failed", zap.Error(err))
		}
		if !shouldRegenerate(prev, fp) {
			log.Info("serving_cached_analysis")
			return sendText(stream, prev.Analysis)
		}
	}

	var buf strings.Builder
	err = h.llm.Stream(ctx, systemPrompt, buildPrompt(snapshot), func(text string) error {
		buf.WriteString(text)
		return stream.Send(&advisorpb.AnalysisChunk{Text: text})
	})
	if err != nil {
		log.Error("inference_failed", zap.Error(err))
		return status.Error(codes.Internal, "analysis failed")
	}

	if h.cache != nil {
		entry := &cache.Entry{
			Snapshot:    snapshot,
			Analysis:    buf.String(),
			Fingerprint: fp,
			GeneratedAt: time.Now(),
		}
		if err := h.cache.Set(ctx, req.UserId, entry); err != nil {
			log.Warn("cache_write_failed", zap.Error(err))
		}
	}

	return stream.Send(&advisorpb.AnalysisChunk{Done: true})
}

func sendText(stream advisorpb.AdvisorService_AnalyzePortfolioServer, text string) error {
	if text != "" {
		if err := stream.Send(&advisorpb.AnalysisChunk{Text: text}); err != nil {
			return err
		}
	}
	return stream.Send(&advisorpb.AnalysisChunk{Done: true})
}
