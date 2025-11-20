-- +goose Up
CREATE TABLE positions (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    symbol VARCHAR(10) NOT NULL,
    PRIMARY KEY (user_id, symbol),
    qty BIGINT NOT NULL,
    realized_pnl BIGINT NOT NULL,
    cost_basis BIGINT NOT NULL DEFAULT 0,
    reserved_qty BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- +goose Down
DROP TABLE positions;
