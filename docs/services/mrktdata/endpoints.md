# Market data endpoints

The market data service exposes one gRPC service, `MrktdataService`, defined in
`proto/mrktdata.proto`. The gateway calls it for reads and the live stream, and the user
service calls it for latest prices. Money fields are integer cents.

| RPC | Request to response | Description |
|-----|---------------------|-------------|
| `GetHistoricalStockData` | `HistoricalStockDataRequest` to `HistoricalStockDataResponse` | bars for a symbol in day, hour, or minute timeframes |
| `GetAvailableSymbols` | `Empty` to `AvailableSymbolsResponse` | the catalog of tradable symbols, cached in memory |
| `GetLatestPrices` | `LatestPricesRequest` to `LatestPricesResponse` | latest price per symbol, used to value positions |
| `WatchlistStream` | stream `WatchlistStreamRequest` to stream `MarketUpdate` | bidirectional live price feed, with reference counted symbol subscriptions |
| `GetNews` | `NewsRequest` to `NewsResponse` | market news for the explore page |
| `GetMovers` | `MoversRequest` to `MoversResponse` | top movers for the explore page |
