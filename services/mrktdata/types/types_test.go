package types

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddStocksCountsReferences(t *testing.T) {
	s := NewStocksSubscriptions()
	assert.True(t, s.Empty())

	s.AddStocks([]string{"AAPL", "TSLA"})
	assert.False(t, s.Empty())

	s.AddStocks([]string{"AAPL"})
	removed := s.RemoveStocks([]string{"AAPL"})
	assert.Empty(t, removed)
	assert.False(t, s.Empty())

	removed = s.RemoveStocks([]string{"AAPL"})
	assert.Equal(t, []string{"AAPL"}, removed)

	removed = s.RemoveStocks([]string{"TSLA"})
	assert.Equal(t, []string{"TSLA"}, removed)
	assert.True(t, s.Empty())
}

func TestRemoveUnknownStockIsNoop(t *testing.T) {
	s := NewStocksSubscriptions()
	removed := s.RemoveStocks([]string{"GHOST"})
	assert.Empty(t, removed)
	assert.True(t, s.Empty())
}

func TestGetNewSymbolsReturnsOnlyUnsubscribed(t *testing.T) {
	s := NewStocksSubscriptions()
	s.AddStocks([]string{"AAPL", "TSLA"})

	news := s.GetNewSymbols([]string{"AAPL", "NVDA", "TSLA", "MSFT"})
	assert.Equal(t, []string{"NVDA", "MSFT"}, news)

	news = s.GetNewSymbols([]string{"AAPL"})
	assert.Empty(t, news)
}

func TestStocksSubscriptionsConcurrentAccess(t *testing.T) {
	s := NewStocksSubscriptions()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			s.AddStocks([]string{"AAPL", "TSLA"})
		}()
		go func() {
			defer wg.Done()
			s.RemoveStocks([]string{"AAPL"})
			s.GetNewSymbols([]string{"NVDA"})
			s.Empty()
		}()
	}
	wg.Wait()
}
