VERSION := $(shell git describe --tags --always)
COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell date +%Y-%m-%dT%H:%M:%S%z)
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

GO_BASE := $(shell pwd)
GO_BIN := $(GO_BASE)/dist
GO_ENV_VARS := GO_BIN=$(GO_BIN)
GO_BINARY := dex-router
GO_CMD := $(GO_BASE)/cmd

LINT := $$(go env GOPATH)/bin/golangci-lint run --timeout=5m -E whitespace -E gosec -E gci -E gomnd -E gofmt -E goimports -E golint --exclude-use-default=false --max-same-issues 0
BUILD := $(GO_ENV_VARS) go build $(LDFLAGS) -o $(GO_BIN)/$(GO_BINARY) $(GO_CMD)

.PHONY: build
build: ## Build the binary locally into ./dist
	$(BUILD)

.PHONY: lint
lint: ## runs linter
	$(LINT)

.PHONY: test
test: ## Runs only short tests without checking race conditions
	go test --cover -short -p 1 ./...

.PHONY: install-linter
install-linter: ## Installs the linter
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.46.2

.PHONY: gen-script
gen-script: ## generate go files
	cd ./etherman/uniswap ./script.sh