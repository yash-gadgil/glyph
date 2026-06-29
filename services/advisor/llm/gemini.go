package llm

import (
	"context"

	"github.com/yash-gadgil/glyph/services/advisor/types"
	"google.golang.org/genai"
)

type geminiProvider struct {
	client *genai.Client
	model  string
}

func NewGemini(ctx context.Context, apiKey, model string) (types.Provider, error) {
	c, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}
	return &geminiProvider{client: c, model: model}, nil
}

func systemConfig(system string, temperature float32, maxTokens int32) *genai.GenerateContentConfig {
	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(temperature),
	}
	if system != "" {
		config.SystemInstruction = &genai.Content{Parts: []*genai.Part{{Text: system}}}
	}
	if maxTokens > 0 {
		config.MaxOutputTokens = maxTokens
	}
	return config
}

func (p *geminiProvider) Stream(ctx context.Context, system, prompt string, emit func(string) error) error {
	config := systemConfig(system, 0.4, 0)
	for result, err := range p.client.Models.GenerateContentStream(ctx, p.model, genai.Text(prompt), config) {
		if err != nil {
			return err
		}
		text := result.Text()
		if text == "" {
			continue
		}
		if err := emit(text); err != nil {
			return err
		}
	}
	return nil
}

func (p *geminiProvider) CompleteShort(ctx context.Context, system, prompt string, maxTokens int32) (string, error) {
	config := systemConfig(system, 0.2, maxTokens)
	result, err := p.client.Models.GenerateContent(ctx, p.model, genai.Text(prompt), config)
	if err != nil {
		return "", err
	}
	return result.Text(), nil
}
