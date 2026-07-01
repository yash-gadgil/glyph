# Advisor

The advisor service is Kenaz, the trading copilot. It runs a tool-calling chat agent over
your portfolio and market data, and generates deployable strategies on request. It is a Go
gRPC service backed by Redis for session state, with no database of its own.

Path: `services/advisor/` · Port: 50058 · Depends on: Redis, the inference service (or the
Gemini API), the user service (portfolio), the strategy service (validate and persist
strategies), the market data service (prices, movers, news).

## Chat

`ChatWithAdvisor` is a streaming RPC, not a single completion. Each incoming message goes
through:

1. **A topic gate.** Keyword matching, ticker patterns, and the recent turns decide
   whether the message is in scope. Anything else gets a fixed refusal instead of reaching
   the model.
2. **A tool loop.** The model can call `get_portfolio`, `get_price`, `get_market`,
   `get_news`, or `generate_strategy` before it answers. Each tool call resolves against a
   downstream service (user, mrktdata, or this service's own strategy-generation path) and
   the result is fed back to the model as context.
3. **Streaming.** Tokens stream back as `AnalysisChunk` messages as the model generates
   them, ending with a `done` chunk.

Sessions are per-user, held in Redis with turn history, an `in_flight` flag, and any
partial text from a response that's still streaming, so a page refresh mid-answer doesn't
lose anything. Sessions are trimmed to the most recent 16 turns. A Redis lock
(`AcquireChatLock`) makes sure only one message processes at a time per user; the lock
times out after 5 minutes.

## Providers

Every chat and strategy-generation request names a provider:

- `inference` - the local model, hosted by the inference service.
- `gemini` - Google's Gemini API, called directly with `GEMINI_API_KEY` and
  `GEMINI_MODEL`.

Both implement the same internal `Provider` interface (`Stream` and `CompleteShort`).
Local inference gets a longer timeout (6 minutes) than Gemini (90 seconds) to allow for
CPU-bound generation. If a requested provider fails during strategy generation, the
service retries once against its own default provider before giving up.

## Strategy generation

`StartStrategyGeneration` returns a job immediately and does the work in a background
goroutine:

1. Prompt the model with the symbol, recent market conditions (trend, RSI, volatility),
   and the caller's portfolio snapshot.
2. Parse the model's response as a strategy config. On a parse or validation failure,
   re-prompt with the error, up to two attempts total.
3. Backtest the candidate through the strategy service over the last year of daily bars.
   A backtest with zero trades is rejected and the model is asked to loosen the rules.
4. Persist the surviving strategy through the strategy service's `CreateStrategy`, with a
   generated name and the model's rationale as the description.

Job state (`running`, `succeeded`, `failed`), the backtest summary, and any error are
cached in Redis and polled through `GetStrategyJob`.

## Endpoints

See [endpoints.md](endpoints.md) for the full `AdvisorService` RPC list.
