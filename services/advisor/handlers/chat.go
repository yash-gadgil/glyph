package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/services/advisor/cache"
	"github.com/yash-gadgil/glyph/services/advisor/types"
	advisorpb "github.com/yash-gadgil/glyph/services/gen/golang/advisor"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	chatTimeout   = 5 * time.Minute
	maxChatIters  = 4
	maxChatTurns  = 16
	chatPersistMS = 400
)

type chatAction struct {
	Thought string `json:"thought"`
	Tool    string `json:"tool"`
	Input   string `json:"input"`
	Final   bool   `json:"final"`
}

func (h *AdvisorHandler) ChatWithAdvisor(req *advisorpb.ChatRequest, stream advisorpb.AdvisorService_ChatWithAdvisorServer) error {
	ctx := stream.Context()
	log := logger.WithContextFields(ctx, h.log).With(logger.Action("chat_with_advisor"))

	if req.UserId == "" {
		return status.Error(codes.InvalidArgument, "user id is required")
	}
	message := strings.TrimSpace(req.Message)
	if message == "" {
		return status.Error(codes.InvalidArgument, "message is required")
	}
	if h.llm == nil {
		return status.Error(codes.Unavailable, "advisor model unavailable")
	}

	if h.cache != nil && h.cache.Enabled() {
		ok, err := h.cache.AcquireChatLock(ctx, req.UserId, chatTimeout)
		if err != nil {
			log.Warn("chat_lock_error", zap.Error(err))
		}
		if !ok {
			return status.Error(codes.FailedPrecondition, "a previous message is still being answered")
		}
		defer h.cache.ReleaseChatLock(context.Background(), req.UserId)
	}

	session := h.loadSession(ctx, req.UserId)
	hadHistory := len(session.Turns) > 0
	session.Turns = append(session.Turns, cache.ChatTurn{Role: "user", Content: message})
	session.InFlight = true
	session.Partial = ""
	h.saveSession(ctx, req.UserId, session)

	genCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), chatTimeout)
	defer cancel()

	var answer string
	if !isOnTopic(message, hadHistory) {
		log.Info("chat_off_topic_skipped")
		answer = offTopicReply
		_ = stream.Send(&advisorpb.AnalysisChunk{Text: answer})
	} else {
		prov := h.provider(req.Provider)
		observations := h.gatherObservations(genCtx, prov, req.UserId, session, log)
		answer = h.streamAnswer(genCtx, prov, req.UserId, session, observations, stream, log)
	}

	session.Turns = append(session.Turns, cache.ChatTurn{Role: "assistant", Content: answer})
	session.Turns = trimTurns(session.Turns)
	session.InFlight = false
	session.Partial = ""
	h.saveSession(context.Background(), req.UserId, session)

	return stream.Send(&advisorpb.AnalysisChunk{Done: true})
}

func (h *AdvisorHandler) GetChatSession(ctx context.Context, req *advisorpb.GetChatSessionRequest) (*advisorpb.ChatSession, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}
	session := h.loadSession(ctx, req.UserId)
	out := &advisorpb.ChatSession{InFlight: session.InFlight, PartialText: session.Partial}
	for _, t := range session.Turns {
		out.Turns = append(out.Turns, &advisorpb.ChatTurn{Role: t.Role, Content: t.Content})
	}
	return out, nil
}

func (h *AdvisorHandler) ClearChatSession(ctx context.Context, req *advisorpb.GetChatSessionRequest) (*advisorpb.ChatSession, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}
	if h.cache != nil {
		h.cache.DeleteChatSession(ctx, req.UserId)
	}
	return &advisorpb.ChatSession{}, nil
}

func (h *AdvisorHandler) loadSession(ctx context.Context, userID string) *cache.ChatSession {
	if h.cache == nil {
		return &cache.ChatSession{}
	}
	session, err := h.cache.GetChatSession(ctx, userID)
	if err != nil {
		h.log.Warn("chat_session_read_failed", zap.Error(err))
	}
	if session == nil {
		return &cache.ChatSession{}
	}
	return session
}

func (h *AdvisorHandler) saveSession(ctx context.Context, userID string, session *cache.ChatSession) {
	if h.cache == nil {
		return
	}
	if err := h.cache.SetChatSession(ctx, userID, session); err != nil {
		h.log.Warn("chat_session_write_failed", zap.Error(err))
	}
}

func (h *AdvisorHandler) gatherObservations(ctx context.Context, prov types.Provider, userID string, session *cache.ChatSession, log *zap.Logger) []string {
	var observations []string
	seen := make(map[string]bool)

	for i := 0; i < maxChatIters; i++ {
		out, err := prov.CompleteShort(ctx, chatPlannerSystem, buildAgentPrompt(session.Turns, observations), 256)
		if err != nil {
			log.Warn("chat_planner_error", zap.Error(err))
			break
		}
		act, ok := parseAction(out)
		if !ok || act.Final || act.Tool == "" {
			break
		}
		key := act.Tool + "|" + strings.ToUpper(strings.TrimSpace(act.Input))
		if seen[key] {
			break
		}
		seen[key] = true
		obs := h.runTool(ctx, userID, act.Tool, act.Input, log)
		log.Info("chat_tool_used", logger.KV("tool", act.Tool), logger.KV("input", act.Input))
		observations = append(observations, fmt.Sprintf("%s(%s) -> %s", act.Tool, act.Input, obs))
	}
	return observations
}

func (h *AdvisorHandler) runTool(ctx context.Context, userID, tool, input string, log *zap.Logger) string {
	switch tool {
	case "get_portfolio":
		snapshot, err := buildSnapshot(ctx, h.portfolio, userID)
		if err != nil {
			return "portfolio is currently unavailable"
		}
		return snapshot
	case "get_price":
		symbol := strings.ToUpper(strings.TrimSpace(input))
		if symbol == "" || h.mrkt == nil {
			return "price is currently unavailable"
		}
		resp, err := h.mrkt.GetLatestPrices(ctx, &mrktpb.LatestPricesRequest{Symbols: []string{symbol}})
		if err != nil || len(resp.Prices) == 0 {
			return fmt.Sprintf("no price available for %s", symbol)
		}
		return fmt.Sprintf("%s last price %s", symbol, dollars(resp.Prices[0].PriceCents))
	case "generate_strategy":
		symbol := strings.ToUpper(strings.TrimSpace(input))
		if symbol == "" {
			symbol = defaultSymbol
		}
		_, err := h.StartStrategyGeneration(ctx, &advisorpb.StartStrategyGenerationRequest{UserId: userID, Symbol: symbol})
		if err != nil {
			return fmt.Sprintf("could not start strategy generation for %s", symbol)
		}
		return fmt.Sprintf("started a strategy generation job for %s; it will appear on the Strategies page when ready", symbol)
	case "get_market":
		return h.marketTool(ctx)
	case "get_news":
		return h.newsTool(ctx, input)
	default:
		return "unknown tool"
	}
}

func (h *AdvisorHandler) marketTool(ctx context.Context) string {
	if h.mrkt == nil {
		return "market data unavailable"
	}
	movers, err := h.mrkt.GetMovers(ctx, &mrktpb.MoversRequest{Limit: 5})
	if err != nil {
		return "market data unavailable"
	}
	var gainers, losers []string
	for _, m := range movers.Gainers {
		gainers = append(gainers, fmt.Sprintf("%s %+.1f%%", m.Symbol, m.ChangePercent))
	}
	for _, m := range movers.Losers {
		losers = append(losers, fmt.Sprintf("%s %+.1f%%", m.Symbol, m.ChangePercent))
	}
	var parts []string
	if len(gainers) > 0 {
		parts = append(parts, "top gainers: "+strings.Join(gainers, ", "))
	}
	if len(losers) > 0 {
		parts = append(parts, "top losers: "+strings.Join(losers, ", "))
	}
	if len(parts) == 0 {
		return "no notable movers right now"
	}
	return strings.Join(parts, " | ")
}

func (h *AdvisorHandler) newsTool(ctx context.Context, input string) string {
	if h.mrkt == nil {
		return "news unavailable"
	}
	symbols := parseSymbols(input)
	req := &mrktpb.NewsRequest{Limit: 6}
	if len(symbols) > 0 {
		req.Symbols = symbols
	}
	resp, err := h.mrkt.GetNews(ctx, req)
	if err != nil || len(resp.Articles) == 0 {
		if len(symbols) > 0 {
			return fmt.Sprintf("no recent news for %s", strings.Join(symbols, ", "))
		}
		return "no recent news"
	}
	var heads []string
	for _, a := range resp.Articles {
		if a.Headline == "" {
			continue
		}
		if len(a.Symbols) > 0 {
			heads = append(heads, fmt.Sprintf("[%s] %s", strings.Join(a.Symbols, ","), a.Headline))
		} else {
			heads = append(heads, a.Headline)
		}
	}
	return strings.Join(heads, "; ")
}

func parseSymbols(input string) []string {
	fields := strings.FieldsFunc(input, func(r rune) bool {
		return r == ',' || r == ' ' || r == ';'
	})
	var out []string
	seen := make(map[string]bool)
	for _, f := range fields {
		s := strings.ToUpper(strings.TrimSpace(f))
		if len(s) < 1 || len(s) > 5 || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

func (h *AdvisorHandler) streamAnswer(ctx context.Context, prov types.Provider, userID string, session *cache.ChatSession, observations []string, stream advisorpb.AdvisorService_ChatWithAdvisorServer, log *zap.Logger) string {
	var b strings.Builder
	clientAlive := true
	lastPersist := time.Now()

	err := prov.Stream(ctx, chatAnswerSystem, buildAnswerPrompt(session.Turns, observations), func(tok string) error {
		b.WriteString(tok)
		if clientAlive {
			if sendErr := stream.Send(&advisorpb.AnalysisChunk{Text: tok}); sendErr != nil {
				clientAlive = false
			}
		}
		if time.Since(lastPersist) > chatPersistMS*time.Millisecond {
			session.Partial = b.String()
			h.saveSession(ctx, userID, session)
			lastPersist = time.Now()
		}
		return nil
	})

	if err != nil {
		log.Warn("chat_answer_error", zap.Error(err))
		if b.Len() == 0 {
			fallback := "Sorry, I ran into a problem answering that. Please try again."
			if isRateLimited(err) {
				fallback = "The assistant is rate limited right now (the free tier quota was reached). Please wait a minute and try again."
			}
			if clientAlive {
				_ = stream.Send(&advisorpb.AnalysisChunk{Text: fallback})
			}
			return fallback
		}
	}

	return b.String()
}

func isRateLimited(err error) bool {
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "resource_exhausted") || strings.Contains(s, "429") || strings.Contains(s, "quota")
}

func parseAction(out string) (chatAction, bool) {
	start := strings.Index(out, "{")
	end := strings.LastIndex(out, "}")
	if start < 0 || end <= start {
		return chatAction{}, false
	}
	var act chatAction
	if err := json.Unmarshal([]byte(out[start:end+1]), &act); err != nil {
		return chatAction{}, false
	}
	return act, true
}

func trimTurns(turns []cache.ChatTurn) []cache.ChatTurn {
	if len(turns) <= maxChatTurns {
		return turns
	}
	return turns[len(turns)-maxChatTurns:]
}

func renderConversation(b *strings.Builder, turns []cache.ChatTurn) {
	b.WriteString("Conversation so far:\n")
	for _, t := range turns {
		role := "User"
		if t.Role == "assistant" {
			role = "Assistant"
		}
		fmt.Fprintf(b, "%s: %s\n", role, t.Content)
	}
}

func buildAgentPrompt(turns []cache.ChatTurn, observations []string) string {
	var b strings.Builder
	renderConversation(&b, turns)
	if len(observations) > 0 {
		b.WriteString("\nTool results so far:\n")
		for _, o := range observations {
			fmt.Fprintf(&b, "- %s\n", o)
		}
	}
	b.WriteString("\nDecide the next single step as JSON.")
	return b.String()
}

func buildAnswerPrompt(turns []cache.ChatTurn, observations []string) string {
	var b strings.Builder
	renderConversation(&b, turns)
	if len(observations) > 0 {
		b.WriteString("\nData you gathered (use it, do not invent figures):\n")
		for _, o := range observations {
			fmt.Fprintf(&b, "- %s\n", o)
		}
	}
	b.WriteString("\nWrite the assistant's reply to the latest user message.")
	return b.String()
}

const chatPlannerSystem = `You are the reasoning loop of Kenaz, a trading assistant inside Glyph, a simulated paper-trading app. You take one step at a time: either call a tool to gather data, or finish once you have enough to answer the user's latest message.

Respond with ONLY a JSON object, no markdown.
To call a tool: {"thought":"why","tool":"<name>","input":"<arg>"}.
When the gathered tool results are enough to answer: {"final":true}.

Tools:
- get_portfolio: input "". Returns the user's cash, equity and the exact holdings they own. Call this for any question about their account, holdings, allocation, risk or P&L, and as the FIRST step whenever the question refers to "my stocks", "what I own", "my positions" or similar.
- get_price: input is a ticker symbol. Returns that stock's latest price.
- get_market: input "". Returns today's biggest movers (gainers and losers).
- get_news: input is an optional comma-separated list of ticker symbols. With symbols it returns recent news for those tickers; with empty input it returns general market news.
- generate_strategy: input is a ticker symbol. Starts a strategy generation job. Use ONLY when the user explicitly asks to create, generate or build a strategy.

Chain tools when needed. For "what's in the news about the stocks I own": first call get_portfolio, then read the tickers it returns and call get_news with those tickers. Never invent tickers, prices or holdings; if you need a fact, call the tool that provides it. Do not repeat a tool call with the same input. When you already have what you need, return {"final":true}.`

const chatAnswerSystem = `You are Kenaz, the trading assistant inside Glyph, a simulated paper-trading app. You are a sharp, friendly analyst the user is chatting with. Hold a natural conversation and be genuinely helpful. Respond to the latest user message.

You help with three kinds of things:
1. The user's own account, live prices, today's market movers and headlines, and kicking off strategy generation, using the data gathered from tools. Use only the figures provided; never invent prices, holdings, movers or P&L.
2. General trading questions from your own knowledge: explain strategies, indicators, risk concepts, and how trading works. When asked to explain or teach something, actually explain it clearly with a brief example. Never deflect a knowledge question to another page.
3. Reasoning that combines the two, for example reading the user's holdings against what the market is doing today.

How to respond:
- Answer the specific question asked, and address every part of a multi-part question. For account questions give your honest read; for "what should I do" give concrete ideas; for "explain X" teach it plainly.
- Talk like a person, not a report. Do not echo the question back or repeat an assessment you already gave earlier; build on the conversation.
- Never state any holding, share count, price, P&L or news headline that is not present in the gathered data above. If the data needed to answer was not gathered, say you could not pull it up rather than guessing. Do not invent tickers the user owns.
- Only call a position "concentrated" if it is a large share of equity, roughly a quarter or more; never call a small position concentrated. Mention idle cash only when it is large, and at most once.
- Plain text, no markdown or bullet symbols. Keep it tight: a few sentences, or a short rundown when teaching a concept.
- Only mention the Strategies page if you actually started a strategy generation job.
- This is a simulated account for educational use, not financial advice.`

var topicKeywords = []string{
	"portfolio", "holding", "position", "stock", "share", "ticker", "equity", "cash", "invest", "trade", "trading", "trader",
	"market", "price", "quote", "buy", "sell", "short", "long", "hedge", "risk", "diversif", "concentrat", "allocation",
	"rebalance", "p&l", "pnl", "profit", "loss", "gain", "return", "dividend", "etf", "option", "future", "volatility",
	"strategy", "strateg", "backtest", "indicator", "rsi", "macd", "sma", "ema", "bollinger", "vwap", "atr", "stochastic",
	"momentum", "breakout", "reversion", "trend", "earnings", "mover", "gainer", "loser", "sector", "index", "nasdaq",
	"s&p", "dow", "bull", "bear", "account", "watchlist", "order", "deploy", "glyph", "kenaz", "news",
}

var tickerLike = regexp.MustCompile(`\b[A-Z]{2,5}\b`)

const offTopicReply = "I'm Kenaz, your trading copilot. I stick to the markets, trading strategies, and your Glyph account, so I can't help with that. Ask me about a stock, your portfolio, or a strategy."

func isOnTopic(message string, hadHistory bool) bool {
	lower := strings.ToLower(message)
	for _, k := range topicKeywords {
		if strings.Contains(lower, k) {
			return true
		}
	}
	if tickerLike.MatchString(message) {
		return true
	}
	if hadHistory && len(strings.Fields(message)) <= 4 {
		return true
	}
	return false
}
