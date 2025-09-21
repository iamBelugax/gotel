MODULE_PATH := github.com/iamBelugax/gotel

# ANSI Color Codes
CYAN := \033[36m
RESET := \033[0m
GREEN := \033[32m
YELLOW := \033[33m

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	| awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-20s$(RESET) %s\n", $$1, $$2}'

tidy: ## Tidy go modules
	@echo "$(CYAN)Tidying Go modules...$(RESET)"
	@go mod tidy
	@echo "$(GREEN)Go modules tidied.$(RESET)"

deps: ## Download and verify Go modules
	@echo "$(CYAN)Downloading Go modules...$(RESET)"
	@go mod download
	@go mod verify
	@echo "$(GREEN)Go modules downloaded.$(RESET)"

fmt: ## Format Go code
	@echo "$(CYAN)Formatting Go code...$(RESET)"
	@go fmt ./...
	@echo "$(GREEN)Formatting complete.$(RESET)"

test: ## Run all unit tests with Ginkgo
	@echo "$(CYAN)Running unit tests...$(RESET)"
	@ginkgo -r -v --randomize-all --fail-on-pending --race --trace ./...

coverage: ## Generate and display test coverage
	@echo "$(CYAN)Generating test coverage...$(RESET)"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out
	@rm coverage.out
	@echo "$(GREEN)Coverage report complete.$(RESET)"

lint: ## Lint the code with golangci-lint
	@echo "$(CYAN)Linting code...$(RESET)"
	@command -v golangci-lint >/dev/null 2>&1 || (echo "$(YELLOW)Installing golangci-lint...$(RESET)"; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@golangci-lint run ./...
	@echo "$(GREEN)Linting complete.$(RESET)"

install-ginkgo: ## Install Ginkgo CLI
	@echo "$(CYAN)Installing Ginkgo CLI...$(RESET)"
	@go get github.com/onsi/ginkgo/v2
	@go get github.com/onsi/gomega/...
	@echo "$(GREEN)Ginkgo CLI installed.$(RESET)"

bootstrap: tidy install-ginkgo ## Bootstrap the development environment
	@echo "$(GREEN)Bootstrap complete.$(RESET)"
