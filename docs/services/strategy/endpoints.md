# Strategy endpoints

The strategy service exposes one gRPC service, `StrategyService`, defined in
`proto/strategy/strategy.proto`. The gateway calls it directly for CRUD and deployments;
the advisor service calls `RunBacktest` and `CreateStrategy` when it generates a strategy.

| RPC | Request to response | Description |
|-----|---------------------|-------------|
| `GetStrategies` | `UserSpecifier` to `StrategiesResponse` | a user's saved strategies |
| `CreateStrategy` | `CreateStrategyRequest` to `Strategy` | save a new strategy config |
| `UpdateStrategy` | `UpdateStrategyRequest` to `Strategy` | rename or edit a strategy's config |
| `DeleteStrategy` | `StrategySpecifier` to `Empty` | delete a strategy |
| `DeployStrategy` | `DeployStrategyRequest` to `Deployment` | start running a strategy live on a symbol |
| `StopDeployment` | `DeploymentSpecifier` to `Deployment` | stop the tick loop for a deployment |
| `DeleteDeployment` | `DeploymentSpecifier` to `Empty` | delete a stopped deployment |
| `GetDeployments` | `UserSpecifier` to `DeploymentsResponse` | a user's deployments and their status |
| `RunBacktest` | `BacktestRequest` to `BacktestResponse` | replay a config over historical bars |

`Deployment` carries `in_position`, `entry_price_cents`, and `qty`, which the engine
updates on every tick. `BacktestResponse` carries the full equity curve and per-trade
detail (`BacktestTrade`), not just the summary stats.
