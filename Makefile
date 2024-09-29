# Directory of Makefile
export ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

## Common configs
ARCH?=$(shell uname -m)
TARGETARCH=$(shell go env GOARCH)
PLATFORM?=linux/$(ARCH)
CONTAINER_TOOL?=docker
DOCKER_BUILDER?=default
DOCKER_SOCK?=/var/run/docker.sock
BUILDKIT_PROGRESS?=plain
REGISTRY?=ghcr.io/llmos-ai

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)
GLOBALBIN ?= /usr/local/bin

## Tool Binaries
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)

## Tool Versions
GOLANGCI_LINT_VERSION ?= v1.54.2

## k3s configs
K3S_VERSION?=v1.29.3+k3s1

## ISO Configs
FLAVOR?=leap
REPO?=$(REGISTRY)/llmos-$(FLAVOR)

## CLI configs
CLI_REPO?=$(REGISTRY)/llmos-cli
MODELS_REPO=$(REGISTRY)/llmos-models

## Elemental configs
ELEMENTAL_TOOLKIT?=ghcr.io/rancher/elemental-toolkit/elemental-cli:v2.1.0

## ollama config
OLLAMA_VERSION?=0.1.32

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

# default target
.PHONY: all
all: build

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: build
build: test lint build-cli ## build LLMOS components

##@ Build
.PHONY: build-cli
build-cli: ## build LLMOS CLI
	EXPORT_ENV=true source ./scripts/version && \
	REGISTRY=$(REGISTRY) \
	BUILDER=$(DOCKER_BUILDER) \
	goreleaser release --snapshot --clean


##@ Release
.PHONY: release-cli
release-cli: ## Release LLMOS CLI
	EXPORT_ENV=true source ./scripts/version && \
	REGISTRY=$(REGISTRY) \
	BUILDER=$(DOCKER_BUILDER) \
	goreleaser release --clean

##@ Development
.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: fmt vet ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test $$(go list ./... | grep -v /e2e) -coverprofile cover.out

# Utilize Kind or modify the e2e tests to load the image locally, enabling compatibility with other vendors.
.PHONY: test-e2e  # Run the e2e tests against a Kind k8s instance that is spun up.
test-e2e:
	go test ./test/e2e/ -v -ginkgo.v

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter & yamllint
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

##@ Dependencies
## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,${GOLANGCI_LINT_VERSION})

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary (ideally with version)
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f $(1) ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv "$$(echo "$(1)" | sed "s/-$(3)$$//")" $(1) ;\
}
endef
