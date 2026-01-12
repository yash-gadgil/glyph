package handlers

import (
	"context"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	mks "github.com/alpacahq/alpaca-trade-api-go/v3/marketdata/stream"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
)

const tradeConflation = time.Second

const (
	reconnectBackoffMin = time.Second
	reconnectBackoffMax = 30 * time.Second
)

type subscriber struct {
	id      int64
	symbols map[string]struct{}
	ch      chan *mrktpb.Bar
}

type Hub struct {
	mu          sync.Mutex
	subs        map[int64]*subscriber
	refs        map[string]int
	nextID      int64
	lastTradeAt map[string]time.Time

	client      *mks.StocksClient
	feed        string
	connected   bool
	supervising bool
	ctx         context.Context

	PriceSink func(symbol string, priceCents int64)
}

func NewHub(ctx context.Context) *Hub {
	feed := os.Getenv("ALPACA_DATA_FEED")
	if feed == "" {
		feed = marketdata.IEX
	}
	return &Hub{
		subs:        make(map[int64]*subscriber),
		refs:        make(map[string]int),
		lastTradeAt: make(map[string]time.Time),
		client:      mks.NewStocksClient(feed),
		feed:        feed,
		ctx:         ctx,
	}
}

func (h *Hub) Register() *subscriber {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.nextID++
	sub := &subscriber{
		id:      h.nextID,
		symbols: make(map[string]struct{}),
		ch:      make(chan *mrktpb.Bar, 256),
	}
	h.subs[sub.id] = sub
	return sub
}

func (h *Hub) Unregister(sub *subscriber) {
	h.mu.Lock()
	delete(h.subs, sub.id)
	released := h.releaseLocked(sub, sub.symbols)
	close(sub.ch)
	h.mu.Unlock()

	h.unsubscribeUpstream(released)
}

func (h *Hub) SetSymbols(sub *subscriber, symbols []string) error {
	want := make(map[string]struct{}, len(symbols))
	for _, s := range symbols {
		want[s] = struct{}{}
	}

	h.mu.Lock()
	added := []string{}
	for s := range want {
		if _, has := sub.symbols[s]; !has {
			sub.symbols[s] = struct{}{}
			h.refs[s]++
			if h.refs[s] == 1 {
				added = append(added, s)
			}
		}
	}
	dropped := map[string]struct{}{}
	for s := range sub.symbols {
		if _, keep := want[s]; !keep {
			dropped[s] = struct{}{}
		}
	}
	released := h.releaseLocked(sub, dropped)
	h.mu.Unlock()

	if len(added) > 0 {
		if err := h.ensureConnected(); err != nil {
			log.Printf("hub: upstream connect failed, serving %v degraded: %v", added, err)
			h.rollbackAdded(sub, added)
		} else if err := h.subscribeUpstream(added); err != nil {
			log.Printf("hub: upstream subscribe failed, dropping %v: %v", added, err)
			h.rollbackAdded(sub, added)
		} else {
			log.Printf("hub: subscribed upstream to %v", added)
		}
	}
	h.unsubscribeUpstream(released)
	return nil
}

func (h *Hub) subscribeUpstream(symbols []string) error {
	h.mu.Lock()
	client := h.client
	h.mu.Unlock()

	if err := client.SubscribeToBars(h.barHandler, symbols...); err != nil {
		return err
	}
	if err := client.SubscribeToTrades(h.tradeHandler, symbols...); err != nil {
		_ = client.UnsubscribeFromBars(symbols...)
		return err
	}
	return nil
}

func (h *Hub) rollbackAdded(sub *subscriber, symbols []string) {
	set := make(map[string]struct{}, len(symbols))
	for _, s := range symbols {
		set[s] = struct{}{}
	}
	h.mu.Lock()
	h.releaseLocked(sub, set)
	h.mu.Unlock()
}

func (h *Hub) releaseLocked(sub *subscriber, symbols map[string]struct{}) []string {
	released := []string{}
	for s := range symbols {
		delete(sub.symbols, s)
		if h.refs[s] > 0 {
			h.refs[s]--
			if h.refs[s] == 0 {
				delete(h.refs, s)
				released = append(released, s)
			}
		}
	}
	return released
}

func (h *Hub) unsubscribeUpstream(symbols []string) {
	if len(symbols) == 0 {
		return
	}
	h.mu.Lock()
	connected := h.connected
	client := h.client
	h.mu.Unlock()
	if !connected {
		return
	}
	if err := client.UnsubscribeFromBars(symbols...); err != nil {
		log.Printf("hub: bar unsubscribe failed: %v", err)
	}
	if err := client.UnsubscribeFromTrades(symbols...); err != nil {
		log.Printf("hub: trade unsubscribe failed: %v", err)
	}
}

func (h *Hub) ensureConnected() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.supervising {
		return nil
	}
	if err := h.client.Connect(h.ctx); err != nil {
		return err
	}
	h.connected = true
	h.supervising = true
	log.Println("hub: connected to Alpaca market data stream")
	go h.superviseConnection()
	return nil
}

func (h *Hub) superviseConnection() {
	for {
		h.mu.Lock()
		client := h.client
		h.mu.Unlock()

		err := <-client.Terminated()
		if h.ctx.Err() != nil {
			return
		}
		log.Printf("hub: upstream stream terminated (%v); reconnecting", err)

		h.mu.Lock()
		h.connected = false
		h.mu.Unlock()

		h.reconnect()
	}
}

func (h *Hub) reconnect() {
	backoff := reconnectBackoffMin
	for {
		if h.ctx.Err() != nil {
			return
		}

		client := mks.NewStocksClient(h.feed)
		if err := client.Connect(h.ctx); err != nil {
			log.Printf("hub: reconnect failed (%v); retrying in %s", err, backoff)
			select {
			case <-time.After(backoff):
			case <-h.ctx.Done():
				return
			}
			backoff = min(backoff*2, reconnectBackoffMax)
			continue
		}

		h.mu.Lock()
		h.client = client
		h.connected = true
		symbols := make([]string, 0, len(h.refs))
		for s := range h.refs {
			symbols = append(symbols, s)
		}
		h.mu.Unlock()

		if len(symbols) > 0 {
			if err := client.SubscribeToBars(h.barHandler, symbols...); err != nil {
				log.Printf("hub: resubscribe bars after reconnect failed: %v", err)
			}
			if err := client.SubscribeToTrades(h.tradeHandler, symbols...); err != nil {
				log.Printf("hub: resubscribe trades after reconnect failed: %v", err)
			}
		}
		log.Printf("hub: reconnected to Alpaca; resubscribed %d symbols", len(symbols))
		return
	}
}

func (h *Hub) barHandler(b mks.Bar) {
	bar := &mrktpb.Bar{
		Symbol: b.Symbol,
		Open:   float32(b.Open),
		High:   float32(b.High),
		Low:    float32(b.Low),
		Close:  float32(b.Close),
		Volume: int64(b.Volume),
		Vwap:   float32(b.VWAP),
		Time:   b.Timestamp.UTC().Format(time.RFC3339),
	}
	h.broadcast(bar)

	if h.PriceSink != nil {
		h.PriceSink(b.Symbol, int64(math.Round(b.Close*100)))
	}
}

func (h *Hub) tradeHandler(t mks.Trade) {
	h.mu.Lock()
	last := h.lastTradeAt[t.Symbol]
	now := time.Now()
	if now.Sub(last) < tradeConflation {
		h.mu.Unlock()
		return
	}
	h.lastTradeAt[t.Symbol] = now
	h.mu.Unlock()

	price := float32(t.Price)
	h.broadcast(&mrktpb.Bar{
		Symbol: t.Symbol,
		Open:   price,
		High:   price,
		Low:    price,
		Close:  price,
		Volume: int64(t.Size),
		Time:   t.Timestamp.UTC().Format(time.RFC3339),
	})

	if h.PriceSink != nil {
		h.PriceSink(t.Symbol, int64(math.Round(t.Price*100)))
	}
}

func (h *Hub) broadcast(bar *mrktpb.Bar) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, sub := range h.subs {
		if _, watching := sub.symbols[bar.Symbol]; !watching {
			continue
		}
		select {
		case sub.ch <- bar:
		default:
		}
	}
}
