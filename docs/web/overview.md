# Web

The web app is the frontend, a Next.js app using the App Router. It is the only client,
and it talks to the gateway over HTTP, WebSocket, and SSE. Nothing in the browser talks to
a backend service directly.

Path: `web/` · Port: 3000.

## What is in it

- a dashboard with your portfolio, account value chart, and holdings
- watchlists with live streaming prices
- charts built on lightweight-charts
- placing and tracking orders
- a strategy builder that produces the JSON config the strategy service runs, backtests,
  and deploys
- an explore page with news and movers
- a global chat widget for Kenaz, the advisor's chat agent, with a picker for the
  inference or Gemini provider

## How it talks to the backend

All backend calls go through a small `api()` helper that talks to the gateway with cookie
based auth. Server state is managed with TanStack Query. Live data, like watchlist prices
and chart updates, comes over WebSocket connections to the gateway, and the chat widget
reads its response as an SSE stream.

Setting `NEXT_PUBLIC_MOCK_API=true` swaps the real client for a mock one, which is how the
app runs in tests and demos without a live backend.

## Routing and auth

Middleware lets the API routes through and bounces the root path to the dashboard while the
auth cookies are present, so a signed in user lands on the dashboard and a signed out one
is sent to login.

## Testing

The web app has unit tests and a Playwright end to end suite. The end to end suite boots
the app in mock mode so it does not need the backend running. `pnpm build` is the full
TypeScript type check.
