# Makefile for harp-plugins
# Wraps mage targets for each plugin module

.DEFAULT_GOAL := help

.PHONY: help all lint clean
.PHONY: releaser-terraformer
.PHONY: terraformer terraformer-tools terraformer-test terraformer-build terraformer-release terraformer-homebrew

##@ General

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-24s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Aggregate Targets

all: terraformer ## Build all plugins (full pipeline)

lint: ## Lint the Makefile
	checkmake --config=.checkmake Makefile

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf terraformer/bin/

##@ Root-level Docker Releases

releaser-terraformer: ## Build harp-terraformer binaries using docker pipeline
	mage releaser:terraformer

##@ Terraformer Plugin

terraformer: ## Build terraformer plugin (full pipeline)
	cd terraformer && mage build

terraformer-tools: ## Install development tools for terraformer plugin
	cd terraformer && mage tools

terraformer-test: ## Run terraformer plugin tests
	cd terraformer && mage test

terraformer-build: ## Build terraformer plugin artifacts
	cd terraformer && mage compile

terraformer-release: ## Cross-compile terraformer plugin for release
	cd terraformer && mage release

terraformer-homebrew: ## Generate Homebrew formula for terraformer plugin
	cd terraformer && mage homebrew
