# Directory of Makefile
export ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

## Common configs
ARCH?=$(shell uname -m)
GOLANG_ARCH=$(shell echo $(ARCH) | sed -e 's/aarch64/arm64/g' -e 's/x86_64/amd64/g' -e 's/riscv64/riscv64/g')
PLATFORM?=linux/$(ARCH)
CONTAINER_TOOL?=docker
DOCKER_BUILDX?=desktop-linux
DOCKER_SOCK?=/var/run/docker.sock
BUILDKIT_PROGRESS=plain

GIT_COMMIT?=$(shell git rev-parse HEAD)
GIT_COMMIT_SHORT?=$(shell git rev-parse --short HEAD)
GIT_TAG?=$(shell git describe --candidates=50 --abbrev=0 --tags 2>/dev/null || echo "v0.0.0-dev" )
VERSION?=$(GIT_TAG)-g$(GIT_COMMIT_SHORT)

## ISO Configs
FLAVOR?=opensuse
REPO?=docker.io/guangbo/llmos-$(FLAVOR)

## CLI configs
LLMOS_CLI_REPO?=docker.io/guangbo/llmos-cli

## Elemental configs
ELEMENTAL_TOOLKIT?=ghcr.io/rancher/elemental-toolkit/elemental-cli:v1.1.2

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
build: build-cli build-os build-iso-local ## build all components(cli, LLMOS image, iso)

.PHONY: build-cli
build-cli: ## build LLMOS CLI
	LLMOS_CLI_REPO=$(LLMOS_CLI_REPO) BUILDER=$(DOCKER_BUILDX) \
	VERSION=$(VERSION) \
	goreleaser release --snapshot --clean

.PHONY: build-local-cli
build-local-cli: ## build LLMOS CLI with local single-target
	INSTALLER_REPO=$(INSTALLER_REPO) BUILDER=$(DOCKER_BUILDX) \
	VERSION=$(VERSION) \
	goreleaser build --single-target --snapshot --clean

.PHONY: build-os
build-os: ## build LLMOS image
	$(CONTAINER_TOOL) buildx build --progress=$(BUILDKIT_PROGRESS) --platform $(PLATFORM) ${DOCKER_ARGS} \
			--build-arg ELEMENTAL_TOOLKIT=$(ELEMENTAL_TOOLKIT) \
			--build-arg LLMOS_CLI_REPO=$(LLMOS_CLI_REPO) \
			--build-arg VERSION=$(VERSION) \
			--build-arg GOLANG_ARCH=$(GOLANG_ARCH) \
			--build-arg REPO=$(REPO) -t $(REPO):$(VERSION) \
			$(BUILD_OPTS) -f iso/images/$(FLAVOR)/Dockerfile .

.PHONY: push-os
push-os: ## push LLMOS image
	$(CONTAINER_TOOL) push $(REPO):$(VERSION)


.PHONY: build-iso
build-iso: ## build LLMOS ISO
	$(CONTAINER_TOOL) buildx build --progress=$(BUILDKIT_PROGRESS) \
			--build-arg REPO=$(REPO) \
			--build-arg VERSION=$(VERSION) \
			--build-arg FLAVOR=$(FLAVOR) \
			--platform $(PLATFORM) \
			-t $(REPO)-iso:$(VERSION) \
			-f Dockerfile.iso .

.PHONY: build-iso-local
build-iso-local: ## build LLMOS ISO locally
	@echo Building $(ARCH) ISO
	rm -rf $(ROOT_DIR)/build
	mkdir -p $(ROOT_DIR)/build
	$(CONTAINER_TOOL) run --rm -v $(DOCKER_SOCK):$(DOCKER_SOCK) -v $(ROOT_DIR)/build:/build -v $(ROOT_DIR)/manifest.yaml:/manifest.yaml \
		--entrypoint /usr/bin/elemental $(REPO):$(VERSION) --debug build-iso --bootloader-in-rootfs -n "LLMOS-$(FLAVOR).$(ARCH)" \
		--local --platform $(PLATFORM) --squash-no-compression --config-dir . \
		-o /build dir:/
