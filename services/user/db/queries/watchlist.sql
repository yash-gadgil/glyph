-- name: CreateWatchlist :one
INSERT INTO watchlists (w_name, user_id)
VALUES ($1, $2)
RETURNING id;

-- name: GetWatchlistsMetadata :many
SELECT id, w_name
FROM watchlists
WHERE user_id = $1;

-- name: GetWatchlist :one
SELECT w.user_id, w.w_name,
        COALESCE(ARRAY_AGG(ws.symbol)
        FILTER (WHERE ws.symbol IS NOT NULL),
         ARRAY[]::text[]) AS symbols
FROM watchlists w
LEFT JOIN watchlist_symbols ws ON w.id = ws.watchlist_id
WHERE w.id = $1 AND w.user_id = $2
GROUP BY w.user_id, w.w_name;

-- name: AddSymbols :many
INSERT INTO watchlist_symbols (watchlist_id, symbol)
SELECT w.id, s.symbol
FROM watchlists w
CROSS JOIN unnest($3::text[]) AS s(symbol)
WHERE w.id = $1 AND w.user_id = $2
ON CONFLICT (watchlist_id, symbol) DO NOTHING
RETURNING watchlist_id, symbol;

-- name: RemoveSymbols :many
DELETE FROM watchlist_symbols ws
USING watchlists w
WHERE ws.watchlist_id = w.id
    AND w.id = $1 AND w.user_id = $2
    AND ws.symbol = ANY($3::text[])
RETURNING ws.watchlist_id, ws.symbol;

-- name: DeleteWatchlist :exec
DELETE FROM watchlists
WHERE id = $1 AND user_id = $2;
