# AI Feature (goai SDK)

This module provides a simple client for using [goai](https://github.com/zendev-sh/goai) in your project.

## Installation

```bash
go get github.com/zendev-sh/goai@latest
```

Requires Go 1.25+.

## Usage

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

## Providers

The client supports multiple providers:

| Provider | Model Example | Environment Variable |
|----------|---------------|---------------------|
| OpenAI | `gpt-4o`, `o3` | `OPENAI_API_KEY` |
| Anthropic | `claude-sonnet-4-6` | `ANTHROPIC_API_KEY` |
| Google | `gemini-2.5-flash` | `GOOGLE_GENERATIVE_AI_API_KEY` |
| AWS Bedrock | `anthropic.claude-sonnet-4-6-v1:0` | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` |
| Azure | `gpt-4o` | `AZURE_OPENAI_API_KEY` |
| Ollama | `llama3` | None (localhost) |

## Features

- `Chat(prompt)` — Simple text generation
- `Stream(prompt)` — Streaming text generation
- `Structured[T](prompt)` — Structured output with generics
- `Embed(text)` — Text embeddings
- `Tools()` — Function calling / tool use

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

## Extending

Edit `client.go` to add more providers or customize the client behavior.

## Learn More

- [goai Documentation](https://goai.sh)
- [API Reference](https://goai.sh/api/core-functions.html)
- [Providers](https://goai.sh/providers/)
