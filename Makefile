# Makefile for trongo - Go TRON format library
.PHONY: help build test test/coverage test/run clean deps go/fmt go/vet go/lint go/staticcheck check docs/serve

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_/-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

# Build targets
build: go/build ## Build the project

go/build: ## Build Go binaries
	go build ./...

# Test targets
test: test/run ## Run all tests

test/run: ## Run tests
	go test -v ./...

test/coverage: ## Run tests with coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Code quality targets
go/fmt: ## Format Go code
	go fmt ./...

go/vet: ## Run go vet
	go vet ./...

go/lint: ## Run golint (if available)
	@if command -v golint >/dev/null 2>&1; then \
		golint ./...; \
	else \
		echo "golint not installed, skipping"; \
	fi

go/staticcheck: ## Run staticcheck (if available)
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo "staticcheck not installed, skipping"; \
	fi

# Pre-commit check
check: go/fmt go/vet test/run ## Run pre-commit checks (format, vet, test)

# Dependencies
deps: ## Install/update dependencies
	go mod tidy
	go mod download

# Documentation
docs/serve: ## Start documentation server
	@echo "Starting documentation server at http://localhost:6060"
	@echo "Visit http://localhost:6060/pkg/github.com/tron-format/trongo/ for package docs"
	godoc -http=:6060

docs/pkg: ## View package documentation (requires PKG variable)
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make docs/pkg PKG=./pkg/tron"; \
		exit 1; \
	fi
	go doc $(PKG)

# Cleanup
clean: ## Clean build artifacts
	go clean ./...
	rm -f coverage.out coverage.html

clean/all: clean ## Clean all generated files
	go clean -cache -testcache -modcache