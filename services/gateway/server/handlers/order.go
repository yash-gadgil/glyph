package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/services/gateway/server/types"
	"github.com/yash-gadgil/glyph/services/gateway/server/utils"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	ordrpb "github.com/yash-gadgil/glyph/services/gen/golang/order"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (cfg *Config) LoadOrderRoutes(r chi.Router) {
	r.Use(cfg.AuthMiddleware)

	r.Get("/", cfg.GetOrders)
	r.Post("/", cfg.CreateOrder)
	r.Delete("/{id}", cfg.DeleteOrder)
}

func (cfg *Config) GetOrders(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("get_orders"),
	)

	if cfg.orderClient == nil {
		log.Error("order_client_nil", logger.Stage("init"))
		utils.ReturnErrorJSON(w, "Order service unavailable", http.StatusServiceUnavailable)
		return
	}

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		log.Error("user_id_missing", logger.Stage("context_extraction"))
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	log = log.With(logger.KV("user_id", userID))

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	limit, offset := parsePageParams(r)
	statusParam := r.URL.Query().Get("status")
	req := &ordrpb.GetOrdersRequest{
		UserId: userID,
		Limit:  limit,
		Offset: offset,
	}

	if statusParam == "" || statusParam == "all" {
		req.AllStatuses = true
	} else {
		statusVal, ok := statusStringToProto(statusParam)
		if !ok {
			utils.ReturnErrorJSON(w, "invalid status parameter", http.StatusBadRequest)
			return
		}
		req.Status = statusVal
	}

	res, err := cfg.orderClient.GetOrders(ctx, req)
	if err != nil {
		log.Error("orders_fetch_error", logger.Stage("fetch_orders"), zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to fetch orders", http.StatusInternalServerError)
		return
	}

	log.Info("fetched_orders", logger.Stage("success"))

	orders := make([]types.Order, len(res.Orders))
	for i, o := range res.Orders {
		orders[i] = protoOrderToJSON(o)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"orders": orders})
}

func (cfg *Config) CreateOrder(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("create_order"),
	)

	if cfg.orderClient == nil {
		log.Error("order_client_nil", logger.Stage("init"))
		utils.ReturnErrorJSON(w, "Order service unavailable", http.StatusServiceUnavailable)
		return
	}

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		log.Error("user_id_missing", logger.Stage("context_extraction"))
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	log = log.With(logger.KV("user_id", userID))

	var reqBody types.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Error("decode_error", logger.Stage("parse_body"), zap.Error(err))
		utils.ReturnErrorJSON(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	side, ok := sideStringToProto(reqBody.Side)
	if !ok {
		utils.ReturnErrorJSON(w, "Invalid side: must be 'buy' or 'sell'", http.StatusBadRequest)
		return
	}

	orderType, ok := orderTypeStringToProto(reqBody.OrderType)
	if !ok {
		utils.ReturnErrorJSON(w, "Invalid orderType: must be 'market', 'limit', 'stop', or 'stop_limit'", http.StatusBadRequest)
		return
	}

	tif, ok := tifStringToProto(reqBody.TimeInForce)
	if !ok {
		utils.ReturnErrorJSON(w, "Invalid timeInForce: must be 'day', 'gtc', 'ioc', or 'fok'", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	var referencePrice int64
	if cfg.mrktdataClient != nil {
		prices, err := cfg.mrktdataClient.GetLatestPrices(ctx, &mrktpb.LatestPricesRequest{
			Symbols: []string{reqBody.Symbol},
		})
		if err != nil {
			log.Warn("reference_price_fetch_failed", zap.Error(err))
		} else {
			for _, p := range prices.Prices {
				if p.Symbol == reqBody.Symbol {
					referencePrice = p.PriceCents
				}
			}
		}
	}

	res, err := cfg.orderClient.PlaceOrder(ctx, &ordrpb.PlaceOrderRequest{
		UserId:              userID,
		Symbol:              reqBody.Symbol,
		Side:                side,
		OrderType:           orderType,
		TimeInForce:         tif,
		Qty:                 reqBody.Quantity,
		Price:               reqBody.Price,
		StopPrice:           reqBody.StopPrice,
		ReferencePriceCents: referencePrice,
	})
	if err != nil {
		log.Error("place_order_error", logger.Stage("place_order"), zap.Error(err))
		code := statusFromGrpc(err, http.StatusInternalServerError)
		msg := "Failed to place order"
		if s, ok := status.FromError(err); ok && code != http.StatusInternalServerError {
			msg = s.Message()
		}
		utils.ReturnErrorJSON(w, msg, code)
		return
	}

	log.Info("order_placed",
		logger.Stage("success"),
		logger.KV("order_id", res.Order.Id),
		logger.KV("symbol", reqBody.Symbol),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"order": protoOrderToJSON(res.Order),
	})
}

func (cfg *Config) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("cancel_order"),
	)

	if cfg.orderClient == nil {
		log.Error("order_client_nil", logger.Stage("init"))
		utils.ReturnErrorJSON(w, "Order service unavailable", http.StatusServiceUnavailable)
		return
	}

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		log.Error("user_id_missing", logger.Stage("context_extraction"))
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	orderID := chi.URLParam(r, "id")
	if orderID == "" {
		utils.ReturnErrorJSON(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	log = log.With(logger.KV("order_id", orderID))

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	res, err := cfg.orderClient.CancelOrder(ctx, &ordrpb.CancelOrderRequest{
		OrderId: orderID,
		UserId:  userID,
	})
	if err != nil {
		log.Error("cancel_order_error", logger.Stage("cancel_order"), zap.Error(err))
		utils.ReturnErrorJSON(w, "Failed to cancel order", http.StatusInternalServerError)
		return
	}

	log.Info("order_cancelled", logger.Stage("success"))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": res.Success,
		"order":   protoOrderToJSON(res.Order),
	})
}

func statusFromGrpc(err error, fallback int) int {
	switch status.Code(err) {
	case codes.FailedPrecondition:
		return http.StatusUnprocessableEntity
	case codes.NotFound:
		return http.StatusNotFound
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.AlreadyExists:
		return http.StatusConflict
	default:
		return fallback
	}
}

func sideStringToProto(s string) (ordrpb.Side, bool) {
	switch s {
	case "buy":
		return ordrpb.Side_BUY, true
	case "sell":
		return ordrpb.Side_SELL, true
	default:
		return 0, false
	}
}

func orderTypeStringToProto(s string) (ordrpb.OrderType, bool) {
	switch s {
	case "market":
		return ordrpb.OrderType_MARKET, true
	case "limit":
		return ordrpb.OrderType_LIMIT, true
	case "stop":
		return ordrpb.OrderType_STOP, true
	case "stop_limit":
		return ordrpb.OrderType_STOP_LIMIT, true
	default:
		return 0, false
	}
}

func tifStringToProto(s string) (ordrpb.TimeInForce, bool) {
	switch s {
	case "day":
		return ordrpb.TimeInForce_DAY, true
	case "gtc":
		return ordrpb.TimeInForce_GTC, true
	case "ioc":
		return ordrpb.TimeInForce_IOC, true
	case "fok":
		return ordrpb.TimeInForce_FOK, true
	default:
		return 0, false
	}
}

func statusStringToProto(s string) (ordrpb.OrderStatus, bool) {
	switch s {
	case "pending":
		return ordrpb.OrderStatus_PENDING, true
	case "open":
		return ordrpb.OrderStatus_OPEN, true
	case "partial_fill":
		return ordrpb.OrderStatus_PARTIAL_FILL, true
	case "filled":
		return ordrpb.OrderStatus_FILLED, true
	case "cancelled":
		return ordrpb.OrderStatus_CANCELLED, true
	case "rejected":
		return ordrpb.OrderStatus_REJECTED, true
	default:
		return 0, false
	}
}

var orderTypeProtoToString = map[ordrpb.OrderType]string{
	ordrpb.OrderType_MARKET:     "market",
	ordrpb.OrderType_LIMIT:      "limit",
	ordrpb.OrderType_STOP:       "stop",
	ordrpb.OrderType_STOP_LIMIT: "stop_limit",
}

var tifProtoToString = map[ordrpb.TimeInForce]string{
	ordrpb.TimeInForce_DAY: "day",
	ordrpb.TimeInForce_GTC: "gtc",
	ordrpb.TimeInForce_IOC: "ioc",
	ordrpb.TimeInForce_FOK: "fok",
}

var statusProtoToString = map[ordrpb.OrderStatus]string{
	ordrpb.OrderStatus_PENDING:      "pending",
	ordrpb.OrderStatus_OPEN:         "open",
	ordrpb.OrderStatus_PARTIAL_FILL: "partial_fill",
	ordrpb.OrderStatus_FILLED:       "filled",
	ordrpb.OrderStatus_CANCELLED:    "cancelled",
	ordrpb.OrderStatus_REJECTED:     "rejected",
}

func protoOrderToJSON(o *ordrpb.Order) types.Order {
	order := types.Order{
		ID:          o.Id,
		UserID:      o.UserId,
		Symbol:      o.Symbol,
		Side:        sideProtoToString[o.Side],
		OrderType:   orderTypeProtoToString[o.OrderType],
		TimeInForce: tifProtoToString[o.TimeInForce],
		Qty:         o.Qty,
		FilledQty:   o.FilledQty,
		Price:       o.Price,
		StopPrice:   o.StopPrice,
		Status:      statusProtoToString[o.Status],
	}
	if o.CreatedAt != nil {
		order.CreatedAt = strconv.FormatInt(o.CreatedAt.Seconds, 10)
	}
	if o.UpdatedAt != nil {
		order.UpdatedAt = strconv.FormatInt(o.UpdatedAt.Seconds, 10)
	}
	return order
}
