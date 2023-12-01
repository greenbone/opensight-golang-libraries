PROJECT := opensight-golang-libraries
REGISTRY := docker-gps.greenbone.net
# Define submodules
SUBMODULES := configReader jobQueue openSearch/escomposite postgres/dbcrypt query/basicResponse
PKG_DIR := pkg

all: build test

.PHONY: help
help: ## show this help.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2}'

.EXPORT_ALL_VARIABLES:
GOPRIVATE=github.com/greenbone

GOIMPORTS       = go run golang.org/x/tools/cmd/goimports@latest
GOFUMPT			= go run mvdan.cc/gofumpt@latest
GOLANGCI-LINT   = go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest
GO-MOD-OUTDATED = go run github.com/psampaz/go-mod-outdated@latest
GO-MOD-UPGRADE  = go run github.com/oligot/go-mod-upgrade@latest
SWAG            = github.com/swaggo/swag/cmd/swag@v1.8.12

INSTALL_MOCKERY	= go install github.com/vektra/mockery/v2@v2.28.2

OS="$(shell go env var GOOS | xargs)"

ALL_GO_DIRS := $(shell find $(PKG_DIR) -name '*.go' -exec dirname {} \; | sort -u)

# Build each submodule
build: $(SUBMODULES)

$(SUBMODULES):
	@echo "Building $(PKG_DIR)/$@"
	@cd $(PKG_DIR)/$@ && go build

# Test each submodule
test:
	@for module in $(SUBMODULES); do \
		echo "Testing $(PKG_DIR)/$$module"; \
		cd $(PKG_DIR)/$$module && go test && cd -; \
	done

# Clean up
clean:
	@for module in $(SUBMODULES); do \
		cd $(PKG_DIR)/$$module && go clean && cd -; \
	done

.PHONY: go-version
go-version: ## prints the golang version used
	@ go version

.PHONY: go-mod-cleanup
go-mod-cleanup: ## cleans up the Go modules
	go mod tidy && go mod download
	go mod verify

.PHONY: clean
clean: ## removes object files from package source directories
	go clean
	@for dir in $(ALL_GO_DIRS); do \
		cd $$dir && go clean && cd -; \
	done

.PHONY: format
format: ## format and tidy
	@echo "\033[36m  Format code  \033[0m"
	$(GOIMPORTS) -l -w .
	GOFUMPT_SPLIT_LONG_LINES=on $(GOFUMPT) -l -w ./pkg
	go fmt ./...

.PHONY: lint
lint: format ## lint go code
	@echo "\033[36m  Lint code  \033[0m"
	$(GOLANGCI-LINT) run

.PHONY: build-common
build-common: go-version clean go-mod-cleanup lint ## execute common build tasks

.PHONY: build
build: build-common ## build go library packages
	@for dir in $(ALL_GO_DIRS); do \
		echo "Building $$dir"; \
		cd $$dir && go build && cd -; \
		echo "Built $$dir"; \
	done

.PHONY: test
test: ## run all tests
	@echo "\033[36m  Run tests  \033[0m"
	go test -test.short ./...

.PHONY: all build test clean $(SUBMODULES)