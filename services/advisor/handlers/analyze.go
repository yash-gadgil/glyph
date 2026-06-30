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
	providers map[string]types.Provider
	cache     *cache.Cache
	log       *zap.Logger
}

func NewAdvisorHandler(portfolio userpb.PortfolioServiceClient, strategy strategypb.StrategyServiceClient, mrkt mrktpb.MrktdataServiceClient, model types.Provider, providers map[string]types.Provider, c *cache.Cache, log *zap.Logger) *AdvisorHandler {
	return &AdvisorHandler{portfolio: portfolio, strategy: strategy, mrkt: mrkt, llm: model, providers: providers, cache: c, log: log}
}

func (h *AdvisorHandler) provider(name string) types.Provider {
	if name != "" {
		if p, ok := h.providers[name]; ok && p != nil {
			return p
		}
	}
	return h.llm
}
