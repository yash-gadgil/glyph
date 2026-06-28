CREATE TABLE strategies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    name VARCHAR(80) NOT NULL,
    config JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT unique_strategy_name_per_user UNIQUE(user_id, name)
);

CREATE INDEX idx_strategies_user ON strategies(user_id);

CREATE TABLE strategy_deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    strategy_id UUID NOT NULL REFERENCES strategies(id) ON DELETE CASCADE,
    symbol VARCHAR(10) NOT NULL,
    position_size_cents BIGINT NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0,
    in_position BOOLEAN NOT NULL DEFAULT FALSE,
    entry_price_cents BIGINT NOT NULL DEFAULT 0,
    qty BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_deployments_user ON strategy_deployments(user_id);
CREATE INDEX idx_deployments_status ON strategy_deployments(status);

CREATE UNIQUE INDEX one_running_deployment_per_strategy_symbol
    ON strategy_deployments(strategy_id, symbol) WHERE status = 0;

