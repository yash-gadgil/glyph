package llm

import (
	"context"
	"errors"
	"io"
	"strings"

	inferpb "github.com/yash-gadgil/glyph/services/gen/golang/inference"
)

type Client struct {
	infer       inferpb.InferenceServiceClient
	maxTokens   int32
	temperature float32
}

func New(infer inferpb.InferenceServiceClient) *Client {
	return &Client{infer: infer, maxTokens: 512, temperature: 0.4}
}

func (c *Client) generate(ctx context.Context, system, prompt string, maxTokens int32, temperature float32, emit func(string) error) error {
	stream, err := c.infer.Generate(ctx, &inferpb.GenerateRequest{
		System:      system,
		Prompt:      prompt,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	})
	if err != nil {
		return err
	}

	for {
		tok, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if tok.Text != "" {
			if err := emit(tok.Text); err != nil {
				return err
			}
		}
		if tok.Done {
			return nil
		}
	}
}

func (c *Client) Stream(ctx context.Context, system, prompt string, emit func(string) error) error {
	return c.generate(ctx, system, prompt, c.maxTokens, c.temperature, emit)
}

func (c *Client) CompleteShort(ctx context.Context, system, prompt string, maxTokens int32) (string, error) {
	var b strings.Builder
	err := c.generate(ctx, system, prompt, maxTokens, 0.2, func(text string) error {
		b.WriteString(text)
		return nil
	})
	return b.String(), err
}
