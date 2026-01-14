package handlers

import (
	"context"
	"testing"
	"time"

	mks "github.com/alpacahq/alpaca-trade-api-go/v3/marketdata/stream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func registerWatching(h *Hub, symbols ...string) *subscriber {
	sub := h.Register()
	h.mu.Lock()
	for _, s := range symbols {
		sub.symbols[s] = struct{}{}
		h.refs[s]++
	}
	h.mu.Unlock()
	return sub
}

func TestHubBroadcastsToEverySubscriberOfASymbol(t *testing.T) {
	h := NewHub(context.Background())
	a := registerWatching(h, "AAPL")
	b := registerWatching(h, "AAPL", "TSLA")
	c := registerWatching(h, "TSLA")

	h.barHandler(mks.Bar{Symbol: "AAPL", Close: 101.5, Timestamp: time.Now()})

	for _, sub := range []*subscriber{a, b} {
		select {
		case bar := <-sub.ch:
			assert.Equal(t, "AAPL", bar.Symbol)
			assert.InDelta(t, 101.5, bar.Close, 0.001)
		default:
			t.Fatal("expected the AAPL bar to reach every AAPL subscriber")
		}
	}

	select {
	case bar := <-c.ch:
		t.Fatalf("TSLA-only subscriber received %s", bar.Symbol)
	default:
	}
}

func TestHubDropsBarsForSlowSubscribers(t *testing.T) {
	h := NewHub(context.Background())
	sub := registerWatching(h, "AAPL")

	done := make(chan struct{})
	go func() {
		for i := 0; i < cap(sub.ch)+10; i++ {
			h.barHandler(mks.Bar{Symbol: "AAPL", Timestamp: time.Now()})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("broadcast blocked on a full subscriber channel")
	}
	assert.Len(t, sub.ch, cap(sub.ch))
}

func TestHubConflatesTradeTicks(t *testing.T) {
	h := NewHub(context.Background())
	sub := registerWatching(h, "AAPL")

	now := time.Now()
	h.tradeHandler(mks.Trade{Symbol: "AAPL", Price: 100, Timestamp: now})
	h.tradeHandler(mks.Trade{Symbol: "AAPL", Price: 101, Timestamp: now})
	h.tradeHandler(mks.Trade{Symbol: "AAPL", Price: 102, Timestamp: now})

	require.Len(t, sub.ch, 1, "trades within the conflation window collapse to one update")
	bar := <-sub.ch
	assert.InDelta(t, 100.0, bar.Close, 0.001)
}

func TestHubRollbackAddedReleasesRefcounts(t *testing.T) {
	h := NewHub(context.Background())
	sub := h.Register()

	h.mu.Lock()
	for _, s := range []string{"AAPL", "TSLA"} {
		sub.symbols[s] = struct{}{}
		h.refs[s]++
	}
	h.mu.Unlock()

	h.rollbackAdded(sub, []string{"AAPL", "TSLA"})

	h.mu.Lock()
	defer h.mu.Unlock()
	assert.Empty(t, h.refs, "rolled-back symbols leave no dangling refcounts")
	assert.NotContains(t, sub.symbols, "AAPL")
	assert.NotContains(t, sub.symbols, "TSLA")
}

func TestHubUnregisterReleasesSymbols(t *testing.T) {
	h := NewHub(context.Background())
	a := registerWatching(h, "AAPL")
	b := registerWatching(h, "AAPL")

	h.Unregister(a)
	h.mu.Lock()
	assert.Equal(t, 1, h.refs["AAPL"], "second subscriber keeps the symbol alive")
	h.mu.Unlock()

	h.Unregister(b)
	h.mu.Lock()
	_, present := h.refs["AAPL"]
	h.mu.Unlock()
	assert.False(t, present, "last unregister releases the symbol")

	h.barHandler(mks.Bar{Symbol: "AAPL", Timestamp: time.Now()})
}
