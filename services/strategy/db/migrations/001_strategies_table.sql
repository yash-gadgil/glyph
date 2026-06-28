-- +goose Up
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

-- +goose Down
DROP TABLE strategies;

