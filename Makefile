PROJECT := opensight-golang-libraries
REGISTRY := docker-gps.greenbone.net
# Define submodules
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

INSTALL_GOMARKDOC = go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest
INSTALL_MOCKERY	= go install github.com/vektra/mockery/v2@v2.28.2

OS="$(shell go env var GOOS | xargs)"

ALL_GO_DIRS := $(shell find $(PKG_DIR) -name '*.go' -exec dirname {} \; | sort -u)

# Clean up
clean:
	go clean -i ./...

.PHONY: go-version
go-version: ## prints the golang version used
	@ go version

.PHONY: go-mod-cleanup
go-mod-cleanup: ## cleans up the Go modules
	go mod tidy && go mod download
	go mod verify

.PHONY: format
format: ## format and tidy
	@echo "\033[36m  Format code  \033[0m"
	$(GOIMPORTS) -l -w .
	GOFUMPT_SPLIT_LONG_LINES=on $(GOFUMPT) -l -w ./pkg
	go fmt ./...

.PHONY: lint
lint: format ## lint go code
	$(GOLANGCI-LINT) run

.PHONY: build-common
build-common: go-version clean go-mod-cleanup lint ## execute common build tasks

.PHONY: build
build: build-common ## build go library packages
	go build -trimpath ./...

.PHONY: test
test: ## run all tests
	go test -test.short ./...

.PHONY: all build test clean

SUBDIRS = pkg/configReader	pkg/dbcrypt	pkg/errorResponses pkg/jobQueue	pkg/openSearch/esextension\
	pkg/openSearch/openSearchClient	pkg/openSearch/openSearchQuery	pkg/query/	pkg/query/filter	pkg/query/paging\
	pkg/query/sorting	pkg/slices	pkg/testFolder pkg/postgres/query
.PHONY: generate_docs $(SUBDIRS)
generate_docs: check_tools $(SUBDIRS)
$(SUBDIRS):
	@cd $@; gomarkdoc -e --output README.md .
	@echo "Generated documentation in $@"

check_tools:
	@command -v gomarkdoc >/dev/null || $(INSTALL_GOMARKDOC)