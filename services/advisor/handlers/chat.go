package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/services/advisor/cache"
	advisorpb "github.com/yash-gadgil/glyph/services/gen/golang/advisor"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	chatTimeout   = 75 * time.Second
	maxChatTools  = 3
	maxChatTurns  = 16
	chatPersistMS = 400
)

type chatPlan struct {
	Tools []struct {
		Tool  string `json:"tool"`
		Input string `json:"input"`
	} `json:"tools"`
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
	session.Turns = append(session.Turns, cache.ChatTurn{Role: "user", Content: message})
	session.InFlight = true
	session.Partial = ""
	h.saveSession(ctx, req.UserId, session)

	genCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), chatTimeout)
	defer cancel()

	observations := h.gatherObservations(genCtx, req.UserId, session, log)
	answer := h.streamAnswer(genCtx, req.UserId, session, observations, stream, log)

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

func (h *AdvisorHandler) gatherObservations(ctx context.Context, userID string, session *cache.ChatSession, log *zap.Logger) []string {
	out, err := h.llm.CompleteShort(ctx, chatPlannerSystem, buildPlanPrompt(session.Turns), 256)
	if err != nil {
		log.Warn("chat_planner_error", zap.Error(err))
		return nil
	}
	plan, ok := parsePlan(out)
	if !ok {
		return nil
	}

	var observations []string
	seen := make(map[string]bool)
	for _, call := range plan.Tools {
		if call.Tool == "" {
			continue
		}
		key := call.Tool + "|" + strings.ToUpper(strings.TrimSpace(call.Input))
		if seen[key] {
			continue
		}
		seen[key] = true
		obs := h.runTool(ctx, userID, call.Tool, call.Input, log)
		log.Info("chat_tool_used", logger.KV("tool", call.Tool), logger.KV("input", call.Input))
		observations = append(observations, fmt.Sprintf("%s(%s) -> %s", call.Tool, call.Input, obs))
		if len(observations) >= maxChatTools {
			break
		}
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
	default:
		return "unknown tool"
	}
}

func (h *AdvisorHandler) streamAnswer(ctx context.Context, userID string, session *cache.ChatSession, observations []string, stream advisorpb.AdvisorService_ChatWithAdvisorServer, log *zap.Logger) string {
	var b strings.Builder
	clientAlive := true
	lastPersist := time.Now()

	err := h.llm.Stream(ctx, chatAnswerSystem, buildAnswerPrompt(session.Turns, observations), func(tok string) error {
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

func parsePlan(out string) (chatPlan, bool) {
	start := strings.Index(out, "{")
	end := strings.LastIndex(out, "}")
	if start < 0 || end <= start {
		return chatPlan{}, false
	}
	var plan chatPlan
	if err := json.Unmarshal([]byte(out[start:end+1]), &plan); err != nil {
		return chatPlan{}, false
	}
	return plan, true
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

func buildPlanPrompt(turns []cache.ChatTurn) string {
	var b strings.Builder
	renderConversation(&b, turns)
	b.WriteString("\nList the tools needed to answer the latest message as JSON.")
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

const chatPlannerSystem = `You are the planning step of a trading assistant inside Glyph, a simulated paper-trading app. Decide which tools, if any, are needed to answer the user's latest message.

Respond with ONLY a JSON object, no markdown: {"tools":[{"tool":"<name>","input":"<arg>"}]}. Use an empty list when no tool is needed.

Tools:
- get_portfolio: input "". Returns the user's cash, equity and current holdings. Use it when the user asks about their account, holdings, allocation, risk or P&L.
- get_price: input is a ticker symbol. Returns that stock's latest price. Use it when the user asks about a price or quote.
- generate_strategy: input is a ticker symbol. Starts a strategy generation job for that symbol. Use it ONLY when the user explicitly asks you to create, generate or build a strategy.

Request only the tools needed for the latest message. If the conversation already contains the answer, return an empty list. Never request the same tool with the same input twice.`

const chatAnswerSystem = `You are Glyph's trading assistant: a sharp, friendly analyst the user is chatting with inside a simulated paper-trading app. Hold a natural conversation and be genuinely helpful. Respond to the latest user message.

You help with two kinds of things:
1. The user's own account, live prices, and kicking off strategy generation, using the data gathered from tools. Use only the figures provided; never invent prices, holdings or P&L.
2. General trading questions from your own knowledge: explain strategies, indicators, risk concepts, and how trading works. When asked to explain or teach something, actually explain it clearly with a brief example. Never deflect a knowledge question to another page.

How to respond:
- Answer the specific question asked. For account questions give your honest read; for "what should I do" give concrete ideas; for "explain X" teach it plainly.
- Talk like a person, not a report. Do not echo the question back or repeat an assessment you already gave earlier.
- Only call a position "concentrated" if it is a large share of equity, roughly a quarter or more; never call a small position concentrated. Mention idle cash only when it is large, and at most once.
- Plain text, no markdown or bullet symbols. Keep it tight: a few sentences, or a short rundown when teaching a concept.
- Only mention the Strategies page if you actually started a strategy generation job.
- This is a simulated account for educational use, not financial advice.`
