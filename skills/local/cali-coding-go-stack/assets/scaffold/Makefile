APP_NAME ?= app
PORT ?= 8080

.PHONY: dev run build test lint templ clean

## dev — live reload with Air
dev:
	@echo "→ Starting Air live reload..."
	@air

## run — run once without watcher
run: templ
	@go run -tags=dev ./cmd/web

## build — production-ish dev build with dev build tag
build: templ
	@go build -tags=dev -o ./tmp/$(APP_NAME) ./cmd/web

## test — run Go tests
test:
	@go test ./...

## lint — run golangci-lint if installed
lint:
	@golangci-lint run ./...

## templ — generate Templ files
templ:
	@go tool templ generate

## clean — remove generated dev artifacts
clean:
	@rm -rf tmp
