# Makefile for harp-terraformer plugin

.DEFAULT_GOAL := help

.PHONY: help all tools test build release homebrew clean

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

all: ## Run full build pipeline (default mage target)
	mage build

tools: ## Install development tools
	mage tools
	go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest

test: ## Run tests
	mage test

build: ## Build artifacts
	mage compile

release: ## Cross-compile for release
	mage release

homebrew: ## Generate Homebrew formula
	mage homebrew

clean: ## Clean build artifacts
	rm -rf bin/
