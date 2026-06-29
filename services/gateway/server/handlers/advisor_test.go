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

type fakeAnalyzeStream struct {
	grpc.ClientStream
	chunks []*advisorpb.AnalysisChunk
	i      int
}

func (f *fakeAnalyzeStream) Recv() (*advisorpb.AnalysisChunk, error) {
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

func (f *fakeAdvisorClient) AnalyzePortfolio(ctx context.Context, in *advisorpb.AnalyzeRequest, opts ...grpc.CallOption) (advisorpb.AdvisorService_AnalyzePortfolioClient, error) {
	return &fakeAnalyzeStream{chunks: f.chunks}, nil
}

func (f *fakeAdvisorClient) StartStrategyGeneration(ctx context.Context, in *advisorpb.StartStrategyGenerationRequest, opts ...grpc.CallOption) (*advisorpb.StrategyJob, error) {
	return &advisorpb.StrategyJob{State: "running"}, nil
}

func (f *fakeAdvisorClient) GetStrategyJob(ctx context.Context, in *advisorpb.GetStrategyJobRequest, opts ...grpc.CallOption) (*advisorpb.StrategyJob, error) {
	return &advisorpb.StrategyJob{}, nil
}

func TestAnalyzePortfolioStreamsSSE(t *testing.T) {
	client := &fakeAdvisorClient{chunks: []*advisorpb.AnalysisChunk{
		{Text: "Your "},
		{Text: "book "},
		{Text: "is concentrated."},
		{Done: true},
	}}

	cfg := handlers.NewTestConfig(new(mocks.MockAuthClient)).WithAdvisorClient(client)
	req := handlers.WithUserID(httptest.NewRequest(http.MethodGet, "/advisor/analyze", nil), "user-1")
	w := httptest.NewRecorder()

	cfg.AnalyzePortfolio(w, req)

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
