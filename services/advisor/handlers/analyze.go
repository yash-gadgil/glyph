package handlers

import (
	"github.com/yash-gadgil/glyph/services/advisor/cache"
	"github.com/yash-gadgil/glyph/services/advisor/types"
	advisorpb "github.com/yash-gadgil/glyph/services/gen/golang/advisor"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	strategypb "github.com/yash-gadgil/glyph/services/gen/golang/strategy"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"go.uber.org/zap"
)

type AdvisorHandler struct {
	advisorpb.UnimplementedAdvisorServiceServer
	portfolio userpb.PortfolioServiceClient
	strategy  strategypb.StrategyServiceClient
	mrkt      mrktpb.MrktdataServiceClient
	llm       types.Provider
	cache     *cache.Cache
	log       *zap.Logger
}

func NewAdvisorHandler(portfolio userpb.PortfolioServiceClient, strategy strategypb.StrategyServiceClient, mrkt mrktpb.MrktdataServiceClient, model types.Provider, c *cache.Cache, log *zap.Logger) *AdvisorHandler {
	return &AdvisorHandler{portfolio: portfolio, strategy: strategy, mrkt: mrkt, llm: model, cache: c, log: log}
}
