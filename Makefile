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

GIT_COMMIT?=$(shell git rev-parse HEAD)
GIT_COMMIT_SHORT?=$(shell git rev-parse --short HEAD)
GIT_TAG?=$(shell git describe --candidates=50 --abbrev=0 --tags 2>/dev/null || echo "v0.0.0-dev" )
VERSION?=$(GIT_TAG)

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
build: build-cli build-airgap build-models build-os build-iso ## build all components(cli, LLMOS image, iso)

##@ Build
.PHONY: build-cli
build-cli: lint test ## build LLMOS CLI
	REGISTRY=$(REGISTRY) \
	BUILDER=$(DOCKER_BUILDER) \
	VERSION=$(VERSION) \
	goreleaser release --snapshot --clean

.PHONY: build-os
build-os: ## build LLMOS image
	$(CONTAINER_TOOL) buildx build --load --progress=$(BUILDKIT_PROGRESS) --platform $(PLATFORM) ${DOCKER_ARGS} \
			--build-arg REPO=$(REPO) \
			--build-arg ELEMENTAL_TOOLKIT=$(ELEMENTAL_TOOLKIT) \
			--build-arg CLI_REPO=$(CLI_REPO) \
			--build-arg MODELS_REPO=$(MODELS_REPO) \
			--build-arg VERSION=$(VERSION) \
			--build-arg ARCH=$(ARCH) \
			--build-arg FLAVOR=$(FLAVOR) \
			--build-arg TARGETARCH=$(TARGETARCH) \
			--build-arg K3S_VERSION=$(K3S_VERSION) \
			-t $(REPO):$(VERSION)-$(TARGETARCH) \
			$(BUILD_OPTS) -f iso/images/$(FLAVOR)/Dockerfile .

.PHONY: build-iso
build-iso: ## build LLMOS ISO
	$(CONTAINER_TOOL) buildx build --progress=$(BUILDKIT_PROGRESS) --platform $(PLATFORM) ${DOCKER_ARGS} \
			--build-arg OS_IMAGE=$(REPO):$(VERSION)-$(TARGETARCH) \
			--build-arg VERSION=$(VERSION) \
			--build-arg FLAVOR=$(FLAVOR) \
			--build-arg ARCH=$(ARCH) \
			-t $(REPO)-iso:$(VERSION)-$(TARGETARCH) \
			$(BUILD_OPTS) --output type=local,dest=${ROOT_DIR}/dist/iso/$(VERSION) \
			-f package/Dockerfile-iso .

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


.PHONY: build-cli-local
build-cli-local: ## build LLMOS CLI with single target
	REGISTRY=$(REGISTRY) \
	BUILDER=$(DOCKER_BUILDER) \
	VERSION=$(VERSION) \
	goreleaser build --snapshot --clean

.PHONY: package-airgap
export K3S_VERSION TARGETARCH OLLAMA_VERSION
package-airgap: ## packaging air-gap artifacts on local
	@echo "packaging air-gap artifacts locally"
	bash $(ROOT_DIR)/scripts/package-airgap

.PHONY: build-airgap ## dind is required for building air-gap image in CI
build-airgap: ## building air-gap image using earthly
	@echo "Building airgap artifacts"
	earthly -P +build-airgap --REGISTRY=$(REGISTRY) --VERSION=$(VERSION)

.PHONY: build-models
build-models: ## build the ollama models
	@echo Building ollama models
	earthly -P +build-models --REGISTRY=$(REGISTRY) --VERSION=$(VERSION)

.PHONY: build-iso-local
build-iso-local: ## build LLMOS ISO locally
	@echo Building $(ARCH) ISO
	$(CONTAINER_TOOL) run --rm -v $(DOCKER_SOCK):$(DOCKER_SOCK) -v $(ROOT_DIR)/dist/iso/$(VERSION):/build \
		-v $(ROOT_DIR)/iso/manifest.yaml:/manifest.yaml \
		--entrypoint /usr/bin/elemental $(REPO):$(VERSION)-$(TARGETARCH) --debug build-iso \
		--local --platform $(PLATFORM) --config-dir . \
		-n "LLMOS-$(FLAVOR)-$(ARCH)" \
		-o /build dir:/

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
