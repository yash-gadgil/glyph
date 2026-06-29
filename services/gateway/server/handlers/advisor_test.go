package handlers_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yash-gadgil/glyph/services/gateway/server/handlers"
	"github.com/yash-gadgil/glyph/services/gateway/tests/mocks"
	advisorpb "github.com/yash-gadgil/glyph/services/gen/golang/advisor"
	"google.golang.org/grpc"
)

type fakeChatStream struct {
	grpc.ClientStream
	chunks []*advisorpb.AnalysisChunk
	i      int
}

func (f *fakeChatStream) Recv() (*advisorpb.AnalysisChunk, error) {
	if f.i >= len(f.chunks) {
		return nil, io.EOF
	}
	c := f.chunks[f.i]
	f.i++
	return c, nil
}

type fakeAdvisorClient struct {
	chunks []*advisorpb.AnalysisChunk
}

func (f *fakeAdvisorClient) ChatWithAdvisor(ctx context.Context, in *advisorpb.ChatRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[advisorpb.AnalysisChunk], error) {
	return &fakeChatStream{chunks: f.chunks}, nil
}

func (f *fakeAdvisorClient) GetChatSession(ctx context.Context, in *advisorpb.GetChatSessionRequest, opts ...grpc.CallOption) (*advisorpb.ChatSession, error) {
	return &advisorpb.ChatSession{}, nil
}

func (f *fakeAdvisorClient) StartStrategyGeneration(ctx context.Context, in *advisorpb.StartStrategyGenerationRequest, opts ...grpc.CallOption) (*advisorpb.StrategyJob, error) {
	return &advisorpb.StrategyJob{State: "running"}, nil
}

func (f *fakeAdvisorClient) GetStrategyJob(ctx context.Context, in *advisorpb.GetStrategyJobRequest, opts ...grpc.CallOption) (*advisorpb.StrategyJob, error) {
	return &advisorpb.StrategyJob{}, nil
}

func TestChatWithAdvisorStreamsSSE(t *testing.T) {
	client := &fakeAdvisorClient{chunks: []*advisorpb.AnalysisChunk{
		{Text: "Your "},
		{Text: "book "},
		{Text: "is concentrated."},
		{Done: true},
	}}

	cfg := handlers.NewTestConfig(new(mocks.MockAuthClient)).WithAdvisorClient(client)
	req := handlers.WithUserID(
		httptest.NewRequest(http.MethodPost, "/advisor/chat", strings.NewReader(`{"message":"how is my book?"}`)),
		"user-1",
	)
	w := httptest.NewRecorder()

	cfg.ChatWithAdvisor(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))

	body := w.Body.String()
	require.Contains(t, body, "data: Your")
	require.Contains(t, body, "data: book")
	require.Contains(t, body, "data: is concentrated.")
	require.Contains(t, body, "event: done")

	combined := strings.Builder{}
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "data: ") && line != "data: end" {
			combined.WriteString(strings.TrimPrefix(line, "data: "))
		}
	}
	assert.Equal(t, "Your book is concentrated.", combined.String())
}
