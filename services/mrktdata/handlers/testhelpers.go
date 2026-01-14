package handlers

import (
	"context"

	alpaca "github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
)

func NewTestMrktdataHandler(stocksApi *marketdata.Client, alpacaClient *alpaca.Client) *MrktdataHandler {
	return &MrktdataHandler{
		hub:          NewHub(context.Background()),
		serviceCtx:   context.Background(),
		stocksApi:    stocksApi,
		alpacaClient: alpacaClient,
	}
}
