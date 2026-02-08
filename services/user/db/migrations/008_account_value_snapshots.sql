-- +goose Up
CREATE TABLE account_value_snapshots (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    equity_cents BIGINT NOT NULL,
    cash_cents BIGINT NOT NULL,
    market_value_cents BIGINT NOT NULL,
    captured_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_snapshots_user_time ON account_value_snapshots(user_id, captured_at);

-- +goose Down
DROP TABLE account_value_snapshots;
