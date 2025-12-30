CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT,
    user_name TEXT NOT NULL
);

CREATE TABLE watchlists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    w_name VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT unique_watchlistname_per_user UNIQUE(user_id, w_name)
);

CREATE TABLE watchlist_symbols (
    watchlist_id UUID NOT NULL REFERENCES watchlists(id) ON DELETE CASCADE,
    symbol VARCHAR(10) NOT NULL,
    PRIMARY KEY (watchlist_id, symbol)
);

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    cash_balance BIGINT NOT NULL DEFAULT 10000000,
    reserved_cash BIGINT NOT NULL DEFAULT 0,
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    multiplier INTEGER NOT NULL DEFAULT 1,
    margin_used BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

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
