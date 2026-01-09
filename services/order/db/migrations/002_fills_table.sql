-- +goose Up
CREATE TABLE fills (
    trade_id    UUID NOT NULL,
    order_id    UUID NOT NULL,
    symbol      VARCHAR(10) NOT NULL,
    side        SMALLINT NOT NULL,
    qty         BIGINT NOT NULL,
    price_cents BIGINT NOT NULL,
    liquidity   VARCHAR(5) NOT NULL,
    executed_at TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (trade_id, order_id)
);

CREATE INDEX idx_fills_order_id ON fills(order_id);
CREATE INDEX idx_fills_executed_at ON fills(executed_at DESC);

-- +goose Down
DROP TABLE fills;
