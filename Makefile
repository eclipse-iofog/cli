SHELL = /bin/bash
OS = $(shell uname -s | tr '[:upper:]' '[:lower:]')

# Build variables
BINARY_NAME = iofogctl
BUILD_DIR ?= bin
PACKAGE_DIR = cmd/iofogctl
MAJOR ?= $(shell cat version | grep MAJOR | sed 's/MAJOR=//g')
MINOR ?= $(shell cat version | grep MINOR | sed 's/MINOR=//g')
PATCH ?= $(shell cat version | grep PATCH | sed 's/PATCH=//g')
SUFFIX ?= $(shell cat version | grep SUFFIX | sed 's/SUFFIX=//g')
VERSION = $(MAJOR).$(MINOR).$(PATCH)$(SUFFIX)
COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null)
BUILD_DATE ?= $(shell date +%FT%T%z)
PREFIX = github.com/eclipse-iofog/iofogctl/v3/pkg/util
LDFLAGS += -X $(PREFIX).versionNumber=$(VERSION) -X $(PREFIX).commit=$(COMMIT) -X $(PREFIX).date=$(BUILD_DATE) -X $(PREFIX).platform=$(GOOS)/$(GOARCH)
LDFLAGS += -X $(PREFIX).portManagerTag=develop
LDFLAGS += -X $(PREFIX).kubeletTag=develop
LDFLAGS += -X $(PREFIX).operatorTag=develop
LDFLAGS += -X $(PREFIX).proxyTag=develop
LDFLAGS += -X $(PREFIX).routerTag=develop
LDFLAGS += -X $(PREFIX).controllerTag=develop
LDFLAGS += -X $(PREFIX).agentTag=3.0.0-dev-b14126 # CA certs fix
LDFLAGS += -X $(PREFIX).controllerVersion=0.0.0-dev
LDFLAGS += -X $(PREFIX).agentVersion=3.0.0-dev-b14126 # CA certs fix
LDFLAGS += -X $(PREFIX).repo=gcr.io/focal-freedom-236620
GO_SDK_MODULE = iofog-go-sdk/v3@7192f877049e0ceb93c41f2a9a268342e904b47a
OPERATOR_MODULE = iofog-operator/v3@c39e28894350a0ad83ca6e841fa479aa0d6619e4
REPORTS_DIR ?= reports
TEST_RESULTS ?= TEST-iofogctl.txt
TEST_REPORT ?= TEST-iofogctl.xml

# Go variables
export CGO_ENABLED ?= 0
export GOOS ?= $(OS)
export GOARCH ?= amd64
GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./client/*")

.PHONY: all
all: bootstrap build install ## Bootstrap env, build and install binary

.PHONY: bootstrap
bootstrap: ## Bootstrap environment
	@cp gitHooks/* .git/hooks/
	@script/bootstrap.sh

.PHONY: build
build: GOARGS += -mod=vendor -tags "$(GOTAGS)" -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)
build: fmt ## Build the binary
	@cd pkg/util && rice embed-go
	@go build -v $(GOARGS) $(PACKAGE_DIR)/main.go

.PHONY: install
install: ## Install the iofogctl binary to /usr/local/bin
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin

.PHONY: modules
modules: ## Get modules
	@for module in $(OPERATOR_MODULE) $(GO_SDK_MODULE); do \
		go get github.com/eclipse-iofog/$$module; \
	done
	@go get github.com/eclipse-iofog/iofogctl@v1.3

.PHONY: get
get: export GOFLAGS=-mod=vendor
get: ## Pull modules

.PHONY: vendor
vendor: modules ## Vendor all modules
	@go mod vendor
	@for module in GeertJohan akavel jessevdk jstemmer nkovacs valyala; do \
		git checkout -- vendor/github.com/$$module; \
	done

.PHONY: lint
lint: fmt ## Lint the source
	@golangci-lint run --timeout 5m0s

.PHONY: fmt
fmt: ## Format the source
	@gofmt -s -w $(GOFILES_NOVENDOR)

.PHONY: test
test: ## Run unit tests
	mkdir -p $(REPORTS_DIR)
	rm -f $(REPORTS_DIR)/*
	set -o pipefail; find ./internal ./pkg -name '*_test.go' -not -path vendor | sed -E "s|(/.*/).*_test.go|\1|g" | xargs -n1 go test -mod=vendor -ldflags "$(LDFLAGS)" -coverprofile=$(REPORTS_DIR)/coverage.txt -v -parallel 1 2>&1 | tee $(REPORTS_DIR)/$(TEST_RESULTS)
	cat $(REPORTS_DIR)/$(TEST_RESULTS) | go-junit-report -set-exit-code > $(REPORTS_DIR)/$(TEST_REPORT)

.PHONY: list
list: ## List all make targets
	@$(MAKE) -pRrn : -f $(MAKEFILE_LIST) 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | sort

.PHONY: clean
clean: ## Clean working env
	rm -rf $(BUILD_DIR)/
	rm -rf $(REPORTS_DIR)

.PHONY: help
.DEFAULT_GOAL := help
help: ## Get help output
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Variable outputting/exporting rules
var-%: ; @echo $($*)
varexport-%: ; @echo $*=$($*)
