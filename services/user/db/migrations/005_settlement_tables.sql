-- +goose Up
CREATE TABLE order_reservations (
    order_id        UUID PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    symbol          VARCHAR(10) NOT NULL,
    side            SMALLINT NOT NULL,
    qty             BIGINT NOT NULL,
    remaining_qty   BIGINT NOT NULL,
    cents_per_share BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_order_reservations_user ON order_reservations(user_id);

CREATE TABLE settlements (
    trade_id           UUID NOT NULL,
    order_id           UUID NOT NULL,
    user_id            UUID NOT NULL,
    symbol             VARCHAR(10) NOT NULL,
    side               SMALLINT NOT NULL,
    qty                BIGINT NOT NULL,
    price_cents        BIGINT NOT NULL,
    cash_delta_cents   BIGINT NOT NULL,
    realized_pnl_cents BIGINT NOT NULL DEFAULT 0,
    applied_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (trade_id, order_id)
);

CREATE INDEX idx_settlements_user ON settlements(user_id, applied_at DESC);

-- +goose Down
DROP TABLE settlements;
DROP TABLE order_reservations;
