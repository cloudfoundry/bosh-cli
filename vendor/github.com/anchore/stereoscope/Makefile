TEMP_DIR = ./.tmp

# Command templates #################################
LINT_CMD = $(TEMP_DIR)/golangci-lint run --tests=false --config .golangci.yaml

# Tool versions #################################
GOLANGCILINT_VERSION := v1.51.0
GOSIMPORTS_VERSION := v0.3.5
BOUNCER_VERSION := v0.4.0
CHRONICLE_VERSION := v0.5.1

# Formatting variables #################################
BOLD := $(shell tput -T linux bold)
PURPLE := $(shell tput -T linux setaf 5)
GREEN := $(shell tput -T linux setaf 2)
CYAN := $(shell tput -T linux setaf 6)
RED := $(shell tput -T linux setaf 1)
RESET := $(shell tput -T linux sgr0)
TITLE := $(BOLD)$(PURPLE)
SUCCESS := $(BOLD)$(GREEN)

# Test variables #################################
COVERAGE_THRESHOLD := 55  # the quality gate lower threshold for unit test total % coverage (by function statements)

## Build variables #################################
SNAPSHOT_DIR := ./snapshot
VERSION := $(shell git describe --dirty --always --tags)

ifndef VERSION
	$(error VERSION is not set)
endif

ifndef TEMP_DIR
    $(error TEMP_DIR is not set)
endif

define title
    @printf '$(TITLE)$(1)$(RESET)\n'
endef

define safe_rm_rf
	bash -c 'test -z "$(1)" && false || rm -rf $(1)'
endef

define safe_rm_rf_children
	bash -c 'test -z "$(1)" && false || rm -rf $(1)/*'
endef

.PHONY: all
all: static-analysis test ## Run all linux-based checks (linting, license check, unit, integration, and linux compare tests)
	@printf '$(SUCCESS)All checks pass!$(RESET)\n'

.PHONY: static-analysis
static-analysis: check-go-mod-tidy lint check-licenses ## Run all static analysis checks

.PHONY: test
test: unit integration benchmark  ## Run all tests (currently unit and integrations)


## Bootstrapping targets #################################

.PHONY: ci-bootstrap
ci-bootstrap: bootstrap
	curl -sLO https://github.com/sylabs/singularity/releases/download/v3.10.0/singularity-ce_3.10.0-focal_amd64.deb && sudo apt-get install -y -f ./singularity-ce_3.10.0-focal_amd64.deb

.PHONY: bootstrap
bootstrap: $(TEMP_DIR) bootstrap-go bootstrap-tools ## Download and install all tooling dependencies (+ prep tooling in the ./tmp dir)
	$(call title,Bootstrapping dependencies)

.PHONY: bootstrap-tools
bootstrap-tools: $(TEMP_DIR)
	GO111MODULE=off GOBIN=$(realpath $(TEMP_DIR)) go get -u golang.org/x/perf/cmd/benchstat
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TEMP_DIR)/ $(GOLANGCILINT_VERSION)
	curl -sSfL https://raw.githubusercontent.com/wagoodman/go-bouncer/master/bouncer.sh | sh -s -- -b $(TEMP_DIR)/ $(BOUNCER_VERSION)
	curl -sSfL https://raw.githubusercontent.com/anchore/chronicle/main/install.sh | sh -s -- -b $(TEMP_DIR)/ $(CHRONICLE_VERSION)
	# the only difference between goimports and gosimports is that gosimports removes extra whitespace between import blocks (see https://github.com/golang/go/issues/20818)
	GOBIN="$(realpath $(TEMP_DIR))" go install github.com/rinchsan/gosimports/cmd/gosimports@$(GOSIMPORTS_VERSION)

.PHONY: bootstrap-go
bootstrap-go:
	go mod download

$(TEMP_DIR):
	mkdir -p $(TEMP_DIR)

## Static analysis targets #################################

.PHONY: static-analysis
static-analysis: check-licenses lint

.PHONY: lint
lint: ## Run gofmt + golangci lint checks
	$(call title,Running linters)
	@printf "files with gofmt issues: [$(shell gofmt -l -s .)]\n"
	@test -z "$(shell gofmt -l -s .)"
	$(LINT_CMD)

.PHONY: lint-fix
lint-fix: ## Auto-format all source code + run golangci lint fixers
	$(call title,Running lint fixers)
	gofmt -w -s .
	$(LINT_CMD) --fix
	go mod tidy

.PHONY: check-licenses
check-licenses:
	$(call title,Validating licenses for go dependencies)
	$(TEMP_DIR)/bouncer check

check-go-mod-tidy:
	@ .github/scripts/go-mod-tidy-check.sh && echo "go.mod and go.sum are tidy!"

## Testing targets #################################

.PHONY: unit
unit: $(TEMP_DIR) ## Run unit tests (with coverage)
	$(call title,Running unit tests)
	go test -coverprofile $(TEMP_DIR)/unit-coverage-details.txt $(shell go list ./... | grep -v anchore/stereoscope/test)
	@.github/scripts/coverage.py $(COVERAGE_THRESHOLD) $(TEMP_DIR)/unit-coverage-details.txt


.PHONY: integration
integration: integration-tools ## Run integration tests
	$(call title,Running integration tests)
	go test -v ./test/integration

## Benchmark test targets #################################


.PHONY: benchmark
benchmark: $(TEMP_DIR) ## Run benchmark tests and compare against the baseline (if available)
	$(call title,Running benchmark tests)
	go test -cpu 2 -p 1 -run=^Benchmark -bench=. -count=5 -benchmem ./... | tee $(TEMP_DIR)/benchmark-$(VERSION).txt
	(test -s $(TEMP_DIR)/benchmark-main.txt && \
		$(TEMP_DIR)/benchstat $(TEMP_DIR)/benchmark-main.txt $(TEMP_DIR)/benchmark-$(VERSION).txt || \
		$(TEMP_DIR)/benchstat $(TEMP_DIR)/benchmark-$(VERSION).txt) \
			| tee $(TEMP_DIR)/benchstat.txt


.PHONY: show-benchstat
show-benchstat:
	@cat $(TEMP_DIR)/benchstat.txt

## Test-fixture-related targets #################################

# note: this is used by CI to determine if the integration test fixture cache (docker image tars) should be busted
.PHONY: integration-fingerprint
integration-fingerprint:
	find test/integration/test-fixtures/image-* -type f -exec md5sum {} + | awk '{print $1}' | sort | md5sum | tee test/integration/test-fixtures/cache.fingerprint

.PHONY: integration-tools-fingerprint
integration-tools-fingerprint:
	@cd test/integration/tools && make fingerprint

.PHONY: integration-tools
integration-tools:
	@cd test/integration/tools && make

.PHONY: integration-tools
integration-tools-load:
	@cd test/integration/tools && make load-cache

.PHONY: integration-tools
integration-tools-save:
	@cd test/integration/tools && make save-cache

## Build-related targets #################################

.PHONY: snapshot
snapshot: clean-snapshot ## Build the binary
	$(call title,Build compatability test)
	@.github/scripts/build.sh $(SNAPSHOT_DIR)

## Cleanup targets #################################

.PHONY: clean
clean: clear-test-cache clean-snapshot ## Delete all generated artifacts
	$(call safe_rm_rf_children,$(TEMP_DIR))

.PHONY: clean-snapshot
clean-snapshot: ## Delete all snapshot builds
	$(call safe_rm_rf,$(SNAPSHOT_DIR))

.PHONY: clear-test-cache
clear-test-cache: ## Delete all test cache (built docker image tars)
	find . -type f -wholename "**/test-fixtures/cache/*.tar" -delete


## Halp! #################################

.PHONY: help
help:  ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(BOLD)$(CYAN)%-25s$(RESET)%s\n", $$1, $$2}'
