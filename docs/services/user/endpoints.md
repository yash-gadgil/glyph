# User endpoints

The user service runs four gRPC services in one process, all defined in `proto/user.proto`.
They are internal: the gateway calls them, and the auth and order services call a few
directly. Money fields are integer cents.

## AccountService

| RPC | Request to response | Description |
|-----|---------------------|-------------|
| `SignupUser` | `SignupUserInfo` to `UserSpecifier` | create the user row and its account row |
| `SigninUser` | `SigninUserInfo` to `UserSpecifier` | credential lookup, also handles the OAuth path |
| `CheckEmailAvailability` | `CheckEmailRequest` to `CheckEmailResponse` | is an email free |
| `UpdatePasswordByEmail` | `UpdatePasswordRequest` to `Empty` | used by password reset |
| `GetProfile` | `UserSpecifier` to `Profile` | email and user name |
| `ResetAccount` | `UserSpecifier` to `Empty` | clear positions, reservations, settlements, and restore the starting balance |
| `ReserveForOrder` | `ReserveForOrderRequest` to `Empty` | hold buying power for a buy or shares for a sell, called by the order service |
| `ReleaseForOrder` | `ReleaseForOrderRequest` to `Empty` | free whatever is left of an order's hold |
| `AddFunds` | `AddFundsRequest` to `AddFundsResponse` | retired, paper accounts have a fixed balance so it returns an error |

## WatchlistService

| RPC | Request to response | Description |
|-----|---------------------|-------------|
| `GetWatchlists` | `UserSpecifier` to `WatchlistsResponse` | all watchlists plus the first one's symbols |
| `GetWatchlist` | `WatchlistSpecifier` to `Watchlist` | one watchlist |
| `CreateWatchlist` | `CreateWatchlistRequest` to `WatchlistSpecifier` | create one |
| `ModifyWatchlist` | `ModifyWatchlistRequest` to `Empty` | subscribe, unsubscribe, or rename |
| `DeleteWatchlist` | `WatchlistSpecifier` to `Empty` | delete one |
| `DeleteSymbolFromWatchlist` | `DeleteSymbolRequest` to `Empty` | remove a single symbol |

## PortfolioService

| RPC | Request to response | Description |
|-----|---------------------|-------------|
| `GetPortfolio` | `UserSpecifier` to `PortfolioResponse` | cash, reserved cash, currency |
| `GetHoldings` | `UserSpecifier` to `HoldingsResponse` | positions valued at the latest price with pnl |
| `GetPositions` | `UserSpecifier` to `PositionsResponse` | raw positions |
| `GetPortfolioHistory` | `PortfolioHistoryRequest` to `PortfolioHistoryResponse` | account value points for the chart |

## StrategyService

| RPC | Request to response | Description |
|-----|---------------------|-------------|
| `GetStrategies` | `UserSpecifier` to `StrategiesResponse` | your strategies |
| `CreateStrategy` | `CreateStrategyRequest` to `Strategy` | create one |
| `UpdateStrategy` | `UpdateStrategyRequest` to `Strategy` | update one |
| `DeleteStrategy` | `StrategySpecifier` to `Empty` | delete one |
| `DeployStrategy` | `DeployStrategyRequest` to `Deployment` | run a strategy live on a symbol |
| `StopDeployment` | `DeploymentSpecifier` to `Deployment` | stop a running deployment |
| `DeleteDeployment` | `DeploymentSpecifier` to `Empty` | delete a stopped deployment |
| `GetDeployments` | `UserSpecifier` to `DeploymentsResponse` | your deployments |
| `RunBacktest` | `BacktestRequest` to `BacktestResponse` | replay a config over historical bars |
