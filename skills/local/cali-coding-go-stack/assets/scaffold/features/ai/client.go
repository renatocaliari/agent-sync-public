package ai

import (
	"context"

	"github.com/zendev-sh/goai"
	"github.com/zendev-sh/goai/provider/anthropic"
	"github.com/zendev-sh/goai/provider/openai"
)

type Client struct {
	model goai.LanguageModel
}

func NewClient(provider, model string) *Client {
	var m goai.LanguageModel
	switch provider {
	case "openai":
		m = openai.Chat(model)
	case "anthropic":
		m = anthropic.Chat(model)
	default:
		m = openai.Chat("gpt-4o")
	}
	return &Client{model: m}
}

func (c *Client) Chat(ctx context.Context, prompt string) (string, error) {
	result, err := goai.GenerateText(ctx, c.model, goai.WithPrompt(prompt))
	return result.Text, err
}

func (c *Client) ChatWithSystem(ctx context.Context, system, prompt string) (string, error) {
	result, err := goai.GenerateText(ctx, c.model,
		goai.WithSystem(system),
		goai.WithPrompt(prompt),
	)
	return result.Text, err
}

func (c *Client) Stream(ctx context.Context, prompt string) (*goai.StreamTextResult, error) {
	stream, err := goai.StreamText(ctx, c.model, goai.WithPrompt(prompt))
	if err != nil {
		return nil, err
	}
	for text := range stream.TextStream() {
		print(text)
	}
	return stream.Result(), stream.Err()
}

type StructuredResult[T any] struct {
	Object T
	Usage  goai.Usage
}

func (c *Client) Structured[T any](ctx context.Context, prompt string) (*StructuredResult[T], error) {
	result, err := goai.GenerateObject[T](ctx, c.model, goai.WithPrompt(prompt))
	if err != nil {
		return nil, err
	}
	return &StructuredResult[T]{
		Object: result.Object,
		Usage:  result.Usage,
	}, nil
}
