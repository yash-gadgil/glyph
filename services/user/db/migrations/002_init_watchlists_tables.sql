-- +goose Up
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

-- +goose Down
DROP TABLE watchlists;
DROP TABLE watchlist_symbols;
