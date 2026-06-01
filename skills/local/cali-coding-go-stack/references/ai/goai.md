# goai — LLM/AI Client

Simple client for integrating LLMs into your project.

**Repo:** [github.com/zendev-sh/goai](https://github.com/zendev-sh/goai)
**Docs:** [goai.sh](https://goai.sh) | [API Reference](https://goai.sh/api/core-functions.html) | [Providers](https://goai.sh/providers/)

---

## Installation

```bash
go get github.com/zendev-sh/goai@latest
```

Requires Go 1.25+.

---

## Quick Start

```go
import (
    "context"
    "your-project/features/ai"
)

client := ai.NewClient("openai", "gpt-4o")

result, err := client.Chat(ctx, "Hello, world!")
if err != nil {
    // handle error
}
fmt.Println(result)
```

---

## Supported Providers

| Provider | Model Example | Environment Variable |
|----------|---------------|---------------------|
| OpenAI | `gpt-4o`, `o3` | `OPENAI_API_KEY` |
| Anthropic | `claude-sonnet-4-6` | `ANTHROPIC_API_KEY` |
| Google | `gemini-2.5-flash` | `GOOGLE_GENERATIVE_AI_API_KEY` |
| AWS Bedrock | `anthropic.claude-sonnet-4-6-v1:0` | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` |
| Azure | `gpt-4o` | `AZURE_OPENAI_API_KEY` |
| Ollama | `llama3` | None (localhost) |

---

## Available Methods

| Method | Description |
|--------|-------------|
| `Chat(prompt)` | Simple text generation |
| `ChatWithSystem(system, prompt)` | Text generation with system prompt |
| `Stream(prompt)` | Streaming text generation |
| `Structured[T](prompt)` | Structured output with generics |
| `Embed(text)` | Text embeddings |
| `Tools()` | Function calling / tool use |

---

## Environment Variables

Most providers auto-detect credentials from environment variables:

```bash
# OpenAI
export OPENAI_API_KEY=sk-...

# Anthropic
export ANTHROPIC_API_KEY=sk-ant-...

# Google
export GOOGLE_GENERATIVE_AI_API_KEY=...  # or GEMINI_API_KEY

# AWS Bedrock
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
export AWS_REGION=us-east-1
```

---

## When to Use vs Voice AI

| Need | Solution |
|------|----------|
| Text/chat generation | goai |
| Embeddings | goai |
| Structured output | goai |
| Voice interaction | LiveKit + Gemini |

---

## Client Implementation

The scaffold includes a wrapper at `features/ai/client.go`:

```go
type Client struct {
    model goai.LanguageModel
}

func NewClient(provider, model string) *Client
func (c *Client) Chat(ctx context.Context, prompt string) (string, error)
func (c *Client) ChatWithSystem(ctx context.Context, system, prompt string) (string, error)
func (c *Client) Stream(ctx context.Context, prompt string) (*goai.StreamTextResult, error)
func (c *Client) Structured[T any](ctx context.Context, prompt string) (*StructuredResult[T], error)
```

Extend `client.go` to add more providers or customize behavior.