# Makefile for harp-plugins
# Wraps mage targets for each plugin module

.DEFAULT_GOAL := help

.PHONY: help all list-all test test-all build-all release-all homebrew-all lint clean
.PHONY: releaser-server releaser-terraformer
.PHONY: assertion assertion-generate assertion-test assertion-build assertion-release assertion-homebrew
.PHONY: aws aws-generate aws-test aws-build aws-release aws-homebrew
.PHONY: server server-generate server-test server-build server-release server-homebrew
.PHONY: yubikey yubikey-generate yubikey-test yubikey-build yubikey-release yubikey-homebrew
.PHONY: terraformer terraformer-tools terraformer-test terraformer-build terraformer-release terraformer-homebrew

##@ General

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-24s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Aggregate Targets

all: assertion aws server yubikey terraformer ## Build all plugins (full pipeline)

list-all: ## List all buildable artifacts
	@echo "Buildable artifacts:"
	@echo "  harp-assertion   - Harp assertion manager"
	@echo "  harp-aws         - Harp AWS Operations"
	@echo "  harp-server      - Harp Crate Server"
	@echo "  harp-yubikey     - Harp Yubikey identity manager"
	@echo "  harp-terraformer - Harp CSO Vault Policy generator"

test: test-all ## Alias for test-all

test-all: assertion-test aws-test server-test yubikey-test terraformer-test ## Run tests for all plugins

build-all: assertion-build aws-build server-build yubikey-build terraformer-build ## Build all plugin artifacts

release-all: assertion-release aws-release server-release yubikey-release terraformer-release ## Release all plugins

homebrew-all: assertion-homebrew aws-homebrew server-homebrew yubikey-homebrew terraformer-homebrew ## Generate Homebrew formulas for all plugins

lint: ## Lint the Makefile
	checkmake --config=.checkmake Makefile

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf assertion/bin/
	rm -rf aws/bin/
	rm -rf server/bin/
	rm -rf yubikey/bin/
	rm -rf terraformer/bin/

##@ Root-level Docker Releases

releaser-server: ## Build harp-server binaries using docker pipeline
	mage releaser:server

releaser-terraformer: ## Build harp-terraformer binaries using docker pipeline
	mage releaser:terraformer

##@ Assertion Plugin

assertion: ## Build assertion plugin (full pipeline)
	cd assertion && mage build

assertion-generate: ## Generate code for assertion plugin
	cd assertion && mage generate

assertion-test: ## Run assertion plugin tests
	cd assertion && mage test

assertion-build: ## Build assertion plugin artifacts
	cd assertion && mage compile

assertion-release: ## Cross-compile assertion plugin for release
	cd assertion && mage release

assertion-homebrew: ## Generate Homebrew formula for assertion plugin
	cd assertion && mage homebrew

##@ AWS Plugin

aws: ## Build AWS plugin (full pipeline)
	cd aws && mage build

aws-generate: ## Generate code for AWS plugin
	cd aws && mage generate

aws-test: ## Run AWS plugin tests
	cd aws && mage test

aws-build: ## Build AWS plugin artifacts
	cd aws && mage compile

aws-release: ## Cross-compile AWS plugin for release
	cd aws && mage release

aws-homebrew: ## Generate Homebrew formula for AWS plugin
	cd aws && mage homebrew

##@ Server Plugin

server: ## Build server plugin (full pipeline)
	cd server && mage build

server-generate: ## Generate code for server plugin
	cd server && mage generate

server-test: ## Run server plugin tests
	cd server && mage test

server-build: ## Build server plugin artifacts
	cd server && mage compile

server-release: ## Cross-compile server plugin for release
	cd server && mage release

server-homebrew: ## Generate Homebrew formula for server plugin
	cd server && mage homebrew

##@ Yubikey Plugin

yubikey: ## Build yubikey plugin (full pipeline)
	cd yubikey && mage build

yubikey-generate: ## Generate code for yubikey plugin
	cd yubikey && mage generate

yubikey-test: ## Run yubikey plugin tests
	cd yubikey && mage test

yubikey-build: ## Build yubikey plugin artifacts
	cd yubikey && mage compile

yubikey-release: ## Cross-compile yubikey plugin for release
	cd yubikey && mage release

yubikey-homebrew: ## Generate Homebrew formula for yubikey plugin
	cd yubikey && mage homebrew

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
