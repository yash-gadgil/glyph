package handlers

import (
	"context"
	"os"
	"sync"
	"time"

	alpaca "github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	"google.golang.org/grpc"
)

const sipProbeInterval = 10 * time.Minute

type cachedSymbol struct {
	Symbol      string
	CompanyName string
}

type MrktdataHandler struct {
	mrktpb.UnimplementedMrktdataServiceServer

	hub        *Hub
	serviceCtx context.Context
	stocksApi  *marketdata.Client

	alpacaClient *alpaca.Client

	symbolsMu    sync.RWMutex
	symbolsCache []cachedSymbol

	feedMu       sync.Mutex
	histFeed     string
	lastSipProbe time.Time
}

func NewMrktdataHandler(sCtx context.Context) *MrktdataHandler {
	apiKey := os.Getenv("APCA_API_KEY_ID")
	apiSecret := os.Getenv("APCA_API_SECRET_KEY")
	baseURL := os.Getenv("APCA_API_BASE_URL")
	if baseURL == "" {
		baseURL = "https://paper-api.alpaca.markets"
	}

	return &MrktdataHandler{
		hub:        NewHub(sCtx),
		serviceCtx: sCtx,
		stocksApi: marketdata.NewClient(marketdata.ClientOpts{
			APIKey:    apiKey,
			APISecret: apiSecret,
		}),
		alpacaClient: alpaca.NewClient(alpaca.ClientOpts{
			APIKey:    apiKey,
			APISecret: apiSecret,
			BaseURL:   baseURL,
		}),
	}
}

func (h *MrktdataHandler) Hub() *Hub {
	return h.hub
}

func Register(grpcServer *grpc.Server, h *MrktdataHandler) {
	mrktpb.RegisterMrktdataServiceServer(grpcServer, h)
}
