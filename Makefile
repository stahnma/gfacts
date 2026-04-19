BINARY    := gfacts
MODULE    := github.com/stahnma/gfacts
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
export CGO_ENABLED := 0
GOFLAGS   := -trimpath
LDFLAGS   := -s -w -X main.version=$(VERSION)

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the gfacts binary
	go build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o bin/$(BINARY) ./cmd/gfacts

.PHONY: install
install: ## Install gfacts to GOPATH/bin
	go install $(GOFLAGS) -ldflags '$(LDFLAGS)' ./cmd/gfacts

.PHONY: test
test: ## Run all tests
	go test ./... -v

.PHONY: test-short
test-short: ## Run tests (skip integration)
	go test ./... -short

.PHONY: lint
lint: ## Run go vet
	go vet ./...

.PHONY: fmt
fmt: ## Format code
	gofmt -s -w .

.PHONY: fmt-check
fmt-check: ## Check formatting (CI)
	@test -z "$$(gofmt -l .)" || { gofmt -l .; exit 1; }

.PHONY: tidy
tidy: ## Tidy go.mod
	go mod tidy

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf bin/

.PHONY: run
run: build ## Build and run gfacts
	./bin/$(BINARY)

.PHONY: run-json
run-json: build ## Build and run with JSON output (piped through jq)
	./bin/$(BINARY) | jq .
