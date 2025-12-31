-- +goose Up
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    side SMALLINT NOT NULL,
    order_type SMALLINT NOT NULL,
    time_in_force SMALLINT NOT NULL DEFAULT 0,
    qty BIGINT NOT NULL,
    filled_qty BIGINT NOT NULL DEFAULT 0,
    price BIGINT,
    stop_price BIGINT,
    status SMALLINT NOT NULL DEFAULT 0,
    strategy_id UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_strategy_id ON orders(strategy_id) WHERE strategy_id IS NOT NULL;
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_user_status ON orders(user_id, status);

-- +goose Down
DROP TABLE orders;
