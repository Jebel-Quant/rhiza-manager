# Makefile for rhiza-manager VS Code Extension

.PHONY: help install build test lint check-types watch clean update compile all

# Default target - show help
.DEFAULT_GOAL := help

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

# Package manager detection
PACKAGE_MANAGER := $(shell command -v pnpm 2> /dev/null && echo pnpm || echo npm)

help: ## Display this help message
	@echo "$(BLUE)rhiza-manager - Available Make Commands$(NC)"
	@echo ""
	@echo "$(GREEN)Development Commands:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BLUE)Package Manager:$(NC) $(PACKAGE_MANAGER)"

install: ## Install dependencies
	@echo "$(GREEN)Installing dependencies with $(PACKAGE_MANAGER)...$(NC)"
	$(PACKAGE_MANAGER) install

build: ## Build the extension
	@echo "$(GREEN)Building extension...$(NC)"
	$(PACKAGE_MANAGER) run build

test: ## Run tests
	@echo "$(GREEN)Running tests...$(NC)"
	$(PACKAGE_MANAGER) run test

lint: ## Run linter
	@echo "$(GREEN)Running linter...$(NC)"
	$(PACKAGE_MANAGER) run lint --fix

check-types: ## Run TypeScript type checking
	@echo "$(GREEN)Running type checking...$(NC)"
	$(PACKAGE_MANAGER) run check-types

watch: ## Run in watch mode for development
	@echo "$(GREEN)Starting watch mode...$(NC)"
	$(PACKAGE_MANAGER) run watch

clean: ## Clean build artifacts and temporary files
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	rm -rf dist/
	rm -rf out/
	rm -rf .vscode-test/
	rm -rf .vscode-test-web/
	rm -f *.vsix
	rm -f *.tsbuildinfo
	@echo "$(GREEN)Clean complete!$(NC)"

update: ## Update dependencies
	@echo "$(GREEN)Updating dependencies with $(PACKAGE_MANAGER)...$(NC)"
	$(PACKAGE_MANAGER) update

compile: check-types lint build ## Full compilation (type check + lint + build)
	@echo "$(GREEN)Full compilation complete!$(NC)"

all: install compile test ## Install, compile, and test
	@echo "$(GREEN)All tasks complete!$(NC)"
