package ai

import (
	"context"

	"github.com/zendev-sh/goai"
	"github.com/zendev-sh/goai/provider"
	"github.com/zendev-sh/goai/provider/anthropic"
	"github.com/zendev-sh/goai/provider/openai"
)

type Client struct {
	model provider.LanguageModel
}

func NewClient(providerName, model string) *Client {
	var m provider.LanguageModel
	switch providerName {
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

func (c *Client) Stream(ctx context.Context, prompt string) (*goai.TextStream, error) {
	stream, err := goai.StreamText(ctx, c.model, goai.WithPrompt(prompt))
	if err != nil {
		return nil, err
	}
	for text := range stream.TextStream() {
		print(text)
	}
	return stream, stream.Err()
}

type StructuredResult struct {
	Object any
	Usage  provider.Usage
}

func (c *Client) StructuredObject(ctx context.Context, prompt string) (*StructuredResult, error) {
	result, err := goai.GenerateObject[any](ctx, c.model, goai.WithPrompt(prompt))
	if err != nil {
		return nil, err
	}
	return &StructuredResult{
		Object: result.Object,
		Usage:  result.Usage,
	}, nil
}
