---
name: cali-coding-standards
description: >
  [Cali] Universal coding standards and principles for all languages and frameworks.
  Use when writing, reviewing, or refactoring code. Covers KISS, DRY, LoB, SoC,
  Fail Fast, Convention over Configuration, and more. Automatically activates on
  Go projects alongside cali-coding-go-standards. Provides file/function size limits,
  CI enforcement patterns, and tie-breaker rules for conflicting principles.

metadata:
  frequency: daily
  category: code
  context-cost: low
---

# Coding Standards

> **Tools:** See `references/cli-tools/` for tool-specific patterns.

## Overview

This skill defines universal coding principles that apply to ALL projects regardless of language or framework. For Go-specific rules, see `cali-coding-go-standards`. For Go + Datastar stack patterns, see `cali-coding-go-stack`.

## When to Use

- Writing new code (any language)
- Reviewing pull requests
- Refactoring existing code
- Setting up CI linting rules
- Teaching coding best practices

## Core Principles

### 1. KISS — Keep It Simple, Stupid
Prefer the boring solution. Clever code is a liability for both humans and LLMs.

- No function >50 lines (Go: 100 lines — see below)
- No file >400 lines (Go: 500 lines — see below)
- Cyclomatic complexity per function <10
- Max 3 indentation levels (use early returns / guard clauses)

### 2. DRY — Don't Repeat Yourself
Wait for the third repetition before abstracting. Premature DRY creates the wrong abstraction.

- Logic duplication → extract shared function
- Configuration duplication → centralize in constant/config
- Template duplication → create partial/component
- Duplication >5% of file is a warning sign

### 3. Convention over Configuration
Follow established conventions of the language/framework before introducing custom config.

- Sensible defaults, standardized names, predictable directory structure
- Explicit configuration only when deviating from convention
- Predictable patterns are a force multiplier for LLMs and developers

### 4. Progressive Disclosure
Simple by default, complexity behind toggles.

- Essentials first, advanced later
- Skills should structure content in 3 tiers: metadata → instructions → references
- Avoid overwhelming with options upfront

### 5. Polymorphism When Useful
Interfaces for extensibility only when it adds real value.

- Prefer concrete types over premature abstractions
- Don't create interfaces for a single implementation
- Add interfaces when you have 2+ concrete implementations

### 6. Locality of Behavior (LoB)
For **Datastar, HTMX, or Alpine.js frontend**: behavior lives in the HTML that uses it.

- Zero custom JavaScript: use native attributes (`data-*`, `@get`, `@post`, `data-on`, `data-bind`, `data-signal`)
- Inline JS only when framework doesn't offer native behavior
- Frontend is a dumb reactive terminal — minimum possible logic on the client

### 7. Separation of Concerns (SoC)
For **backend** (Go handlers, services, repos) and **non-Datastar frameworks** (React, Vue, Svelte).

- Separate template, logic, data, and style in distinct layers
- Component/layer does one thing: handler shouldn't call database directly
- Business logic outside template: extract to services, repositories, helpers

### 8. Fail Fast
Validate at the boundary. Return errors immediately.

- Never silently swallow errors
- Never defer validation to a later layer
- Guard clauses at function entry
- Return meaningful error messages

### 9. SSE-First
Prefer Server-Sent Events over WebSockets for one-directional updates.

- SSE is simpler, has automatic browser reconnection, more efficient than polling
- WebSockets only for real bidirectional communication (chat, collaboration, gaming)
- Server drives state; client renders it

### 10. HATEOAS
Backend determines which actions the user can take.

- Links and forms are discovered via hypertext
- Actions trigger requests, backend responds with HTML/Datastar morphs into DOM
- Frontend doesn't hardcode URLs or action availability

### 11. YAGNI — You Aren't Gonna Need It
Don't build for future needs. Implement only what's needed now.

- No speculative features
- No "we might need this later" abstractions
- Refactor when the need actually arrives

## Tie-Breaker Rule

When LoB and SoC conflict:

| Context | Principle |
|---|---|
| Datastar frontend (`data-*` attributes) | ✅ **LoB** — behavior in the HTML that uses it |
| Datastar project backend (Go handlers) | ✅ **SoC** — separation into layers |
| Non-Datastar project (React, Vue, etc.) | ✅ **SoC** — everything in separate layers |
| Mix Datastar + other framework | ⚠️ LoB on Datastar frontend, SoC on rest |
| Unsure | **SoC** is the safe default |

## Go-Specific Overrides

When working on Go + Datastar projects, `cali-coding-go-stack` extends these principles with stack-specific rules:

| Rule | Universal | Go Override |
|---|---|---|
| Lines per function | 50 | 100 |
| Lines per file | 400 | 500 |
| HTML in code | Allowed | `fmt.Sprintf` with HTML tags: **BLOCKED by CI** |

**Why Go relaxes limits:** Typed language, explicit error handling adds lines, Go convention favors longer but linear functions.

**Why HTML in Go is blocked:** Go's `html/template` handles XSS escaping automatically. `fmt.Sprintf` bypasses this safety. CI enforces: `grep -r 'fmt\.Sprintf.*<' .` must return empty.

## File and Function Size Limits

See `references/file-function-sizes.md` for detailed limits by language and enforcement patterns.

| Metric | Universal | Go Override |
|---|---|---|
| Lines per function | 50 | 100 |
| Lines per file | 400 | 500 |
| Cyclomatic complexity | 10 | 10 |
| Indentation depth | 3 levels | 3 levels |

## CI Enforcement

See `references/ci-enforcement.md` for CI patterns to enforce these standards.

## Output Structure

When applying these principles, produce code that:

1. **Follows the principle hierarchy** — KISS > DRY > Convention > Progressive Disclosure
2. **Uses the tie-breaker rule** — determines LoB vs SoC by context
3. **Respects size limits** — functions <50 lines, files <400 lines (Go: 100/500)
4. **Has clear boundaries** — each function does one thing, each file has one responsibility
5. **Fails fast** — validates at boundaries, returns errors immediately

## Expected Behavior

### Strong Output
- Code that follows all 11 principles naturally
- Functions under the size limit
- Clear separation between frontend (LoB) and backend (SoC)
- Error handling at boundaries
- No premature abstractions

### Weak Output
- Functions >50 lines (Go: >100 lines)
- Files >400 lines (Go: >500 lines)
- Business logic in templates
- `fmt.Sprintf` with HTML tags in Go
- Error swallowing (`if err != nil { return nil }`)
- Premature interfaces for single implementations

## Relationship with Other Skills

| Skill | Relationship |
|---|---|
| `cali-coding-go-standards` | Go-specific engineering rules (concurrency, linting, security) |
| `cali-coding-go-stack` | Go + Datastar stack patterns (extends these principles) |
| `cali-product-tech-planning` | Uses these principles for tech plan generation |

## Installation

This skill is part of the global skills collection at `~/.agents/skills/`. No installation needed — it's automatically discovered by any CLI that supports Agent Skills.
