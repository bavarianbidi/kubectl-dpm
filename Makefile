# SPDX-License-Identifier: MIT

# Image URL to use all building/pushing image targets
IMG ?= controller:latest

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: fmt vet ## Run tests.
	go test ./... -coverprofile cover.out

##@ Build

.PHONY: build
build: fmt vet ## Build manager binary.
	go build -o bin/kubectl-dpm cmd/kubectl-dpm.go

.PHONY: run
run: fmt vet ## Run a controller from your host.
	go run ./cmd/kubectl-dpm.go

##@ SBOM

.PHONY: sbom
sbom: kbom sbom-generate

.PHONY: sbom-generate
sbom-generate: kbom ## Generate SBOM
	mkdir -p tmp
	$(KBOM) generate --output tmp/kubectl-dpm.bom.spdx --format json .

##@ Release

.PHONY: release
release: sbom-generate goreleaser ## Create a new release
	$(GORELEASER) release --clean

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
GORELEASER ?= $(LOCALBIN)/goreleaser
NANCY ?= $(LOCALBIN)/nancy
GOVULNCHECK ?= $(LOCALBIN)/govulncheck
KBOM ?= $(LOCALBIN)/bom

## Tool Versions
GOLANGCI_LINT_VERSION ?= v1.60.3
GORELEASER_VERSION ?= v2.0.1
NANCY_VERSION ?= v1.0.46
GOVULNCHECK_VERSION ?= v1.1.3
KBOM_VERSION ?= v0.6.0

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary. If wrong version is installed, it will be overwritten.
$(GOLANGCI_LINT): $(LOCALBIN)
	test -s $(LOCALBIN)/golangci-lint && $(LOCALBIN)/golangci-lint --version | grep -q $(GOLANGCI_LINT_VERSION) || \
	GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: goreleaser
goreleaser: $(GORELEASER) ## Download goreleaser locally if necessary. If wrong version is installed, it will be overwritten.
$(GORELEASER): $(LOCALBIN)
	test -s $(LOCALBIN)/goreleaser && $(LOCALBIN)/goreleaser --version | grep -q $(GORELEASER_VERSION) || \
	GOBIN=$(LOCALBIN) go install github.com/goreleaser/goreleaser/v2@$(GORELEASER_VERSION)

.PHONY: nancy
nancy: $(NANCY) ## Download nancy locally if necessary. If wrong version is installed, it will be overwritten.
$(NANCY): $(LOCALBIN)
	test -s $(LOCALBIN)/nancy && $(LOCALBIN)/nancy --version | grep -q $(NANCY_VERSION) || \
	GOBIN=$(LOCALBIN) go install github.com/sonatype-nexus-community/nancy@$(NANCY_VERSION)

.PHONY: govulncheck
govulncheck: $(GOVULNCHECK) ## Download govulncheck locally if necessary. If wrong version is installed, it will be overwritten.
$(GOVULNCHECK): $(LOCALBIN)
	test -s $(LOCALBIN)/govulncheck && $(LOCALBIN)/govulncheck -version | grep -q $(GOVULNCHECK_VERSION) || \
	GOBIN=$(LOCALBIN) go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

.PHONY: kbom
kbom: $(KBOM) ## Download kbom locally if necessary. If wrong version is installed, it will be overwritten.
$(KBOM): $(LOCALBIN)
	test -s $(LOCALBIN)/bom && $(LOCALBIN)/bom version | grep -q $(KBOM_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/bom/cmd/bom@$(KBOM_VERSION)

##@ Lint / Verify
.PHONY: lint
lint: $(GOLANGCI_LINT) ## Run linting.
	$(GOLANGCI_LINT) run -v $(GOLANGCI_LINT_EXTRA_ARGS)

.PHONY: lint-fix
lint-fix: $(GOLANGCI_LINT) ## Lint the codebase and run auto-fixers if supported by the linte
	GOLANGCI_LINT_EXTRA_ARGS=--fix $(MAKE) lint

ALL_VERIFY_CHECKS = security license

.PHONY: verify
verify: $(addprefix verify-,$(ALL_VERIFY_CHECKS)) ## Run all verify-* targets

.PHONY: verify-license
verify-license: ## Verify license headers
	./hack/verify-license.sh

.PHONY: verify-security
verify-security: govulncheck-scan nancy-scan ## Verify security by running govulncheck and nancy
	@echo "Security checks passed"

.PHONY: govulncheck-scan
govulncheck-scan: govulncheck ## Perform govulncheck scan
	$(GOVULNCHECK) ./...

.PHONY: nancy-scan
nancy-scan: nancy ## Perform nancy scan
	go list -json -deps ./... | $(NANCY) sleuth
