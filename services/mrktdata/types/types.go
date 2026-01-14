package types

import "sync"

type StocksSubscriptions struct {
	mut    sync.RWMutex
	stocks map[string]uint64
}

func NewStocksSubscriptions() *StocksSubscriptions {
	return &StocksSubscriptions{
		stocks: make(map[string]uint64),
	}
}

func (s *StocksSubscriptions) AddStocks(stocks []string) {
	s.mut.Lock()
	defer s.mut.Unlock()

	for _, stock := range stocks {
		s.stocks[stock]++
	}
}

func (s *StocksSubscriptions) RemoveStocks(stocks []string) []string {
	s.mut.Lock()
	defer s.mut.Unlock()

	removed := []string{}

	for _, stock := range stocks {
		if count, ok := s.stocks[stock]; ok {
			if count <= 1 {
				removed = append(removed, stock)
				delete(s.stocks, stock)
			} else {
				s.stocks[stock] = count - 1
			}
		}
	}

	return removed
}

func (s *StocksSubscriptions) Empty() bool {
	s.mut.Lock()
	defer s.mut.Unlock()

	return len(s.stocks) == 0
}

func (s *StocksSubscriptions) GetNewSymbols(symbols []string) []string {
	s.mut.Lock()
	defer s.mut.Unlock()

	newSymbols := []string{}
	for _, symbol := range symbols {
		if _, exists := s.stocks[symbol]; !exists {
			newSymbols = append(newSymbols, symbol)
		}
	}
	return newSymbols
}
