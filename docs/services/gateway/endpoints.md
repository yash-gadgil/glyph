# Gateway endpoints

The gateway is the public HTTP, WebSocket, and SSE API. Everything here is what the web
app calls. Money fields are integer cents.

Every group except `/auth` is behind the auth middleware, which checks the JWT cookie and
injects your user id. The WebSocket routes authenticate from the cookie inside the
handler instead, and `/advisor/chat` authenticates normally but responds as an SSE stream
instead of a single JSON body.

## Auth `/auth`

Public, no cookie required.

| Method | Path | Description |
|--------|------|-------------|
| POST | `/auth/signup` | start a signup, sends a verification email |
| POST | `/auth/signin` | sign in, sets the auth cookies |
| POST | `/auth/signout` | clear the auth cookies |
| POST | `/auth/forgot-password` | send a reset email |
| POST | `/auth/reset-password` | set a new password from a reset token |
| GET | `/auth/oauth/{provider}` | start an OAuth login, redirects to the provider |
| GET | `/auth/oauth/{provider}/callback` | OAuth return, sets the cookies |
| GET | `/auth/verify` | verify an email from the link token |
| GET | `/auth/refresh` | rotate the access and refresh cookies |

## Orders `/orders`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/orders` | list your orders, optional status filter |
| POST | `/orders` | place an order |
| DELETE | `/orders/{id}` | cancel an order |

## Portfolio `/portfolio`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/portfolio` | cash balance, reserved cash, currency |
| GET | `/portfolio/holdings` | positions valued at the latest price, with unrealized and realized pnl |
| GET | `/portfolio/positions` | raw positions |
| GET | `/portfolio/history` | account value points over a lookback window for the chart |

## Watchlists `/watchlists`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/watchlists` | your watchlists and the first one's symbols |
| GET | `/watchlists/symbols` | the catalog of tradable symbols |
| POST | `/watchlists/history` | historical bars for a symbol and timeframe |
| GET | `/watchlists/stream` | WebSocket, live prices for an explicit `?symbols=` set |
| GET | `/watchlists/{id}/info` | one watchlist with its symbols |
| GET | `/watchlists/{id}` | WebSocket, live prices for a saved watchlist |
| POST | `/watchlists` | create a watchlist |
| PATCH | `/watchlists/{id}` | subscribe, unsubscribe, or rename |
| DELETE | `/watchlists/{id}` | delete a watchlist |

## Account `/account`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/account` | the account summary |
| DELETE | `/account` | delete the account |
| GET | `/account/me` | the signed in user |
| GET | `/account/profile` | email and user name |
| GET | `/account/trades` | your fills |
| POST | `/account/reset` | reset the account back to the starting balance |

## Explore `/explore`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/explore/news` | market news |
| GET | `/explore/movers` | top movers |

## Strategies `/strategies`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/strategies` | your strategies |
| POST | `/strategies` | create a strategy |
| PATCH | `/strategies/{id}` | update a strategy |
| DELETE | `/strategies/{id}` | delete a strategy |
| POST | `/strategies/backtest` | run a config over historical bars and return the results |
| POST | `/strategies/{id}/deploy` | deploy a strategy live on a symbol |
| GET | `/strategies/deployments` | your deployments |
| POST | `/strategies/deployments/{id}/stop` | stop a running deployment |
| DELETE | `/strategies/deployments/{id}` | delete a stopped deployment |
| GET | `/strategies/{id}/trades` | the fills a strategy has made |

## Advisor `/advisor`

Kenaz, the chat agent, and strategy generation.

| Method | Path | Description |
|--------|------|-------------|
| POST | `/advisor/chat` | send a message, SSE stream of response tokens back |
| GET | `/advisor/chat/session` | current session turns and in-flight status |
| DELETE | `/advisor/chat/session` | clear the session |
| POST | `/advisor/strategy` | start strategy generation for a symbol, returns a job |
| GET | `/advisor/strategy/status` | poll the strategy generation job |
