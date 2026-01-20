-- +goose Up
DROP INDEX IF EXISTS one_running_deployment_per_strategy;

CREATE UNIQUE INDEX one_running_deployment_per_strategy_symbol
    ON strategy_deployments(strategy_id, symbol) WHERE status = 0;

-- +goose Down
DROP INDEX IF EXISTS one_running_deployment_per_strategy_symbol;

CREATE UNIQUE INDEX one_running_deployment_per_strategy
    ON strategy_deployments(strategy_id) WHERE status = 0;
