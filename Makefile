BINARY := mcp42
PORT   ?= 8080

.DEFAULT_GOAL := help

.PHONY: help build run run-http inspect clean

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*##"}; {printf "  %-12s %s\n", $$1, $$2}'

build: ## Build the binary
	go build -o $(BINARY) .

run: build ## Run with stdio transport (for Claude Code)
	./$(BINARY)

run-http: build ## Run HTTP server on PORT (default 8080)
	./$(BINARY) --transport http --port $(PORT)

inspect: ## Open MCP Inspector against local HTTP server
	npx @modelcontextprotocol/inspector http://localhost:$(PORT)/mcp

clean: ## Remove the built binary
	rm -f $(BINARY)
