VERSION --arg-scope-and-set 0.7

LET go_version = 1.21
LET distro = alpine3.19

FROM golang:${go_version}-${distro}
ARG --global ALPINE=3.19
ARG --global ALPINE_DIND=earthly/dind:alpine-3.19
ARG --global K3S_VERSION=v1.29.3+k3s1
ARG --global OLLAMA_VERSION=0.1.32
ARG --global ELEMENTAL_TOOLKIT=ghcr.io/rancher/elemental-toolkit/elemental-cli:v2.1.0
ARG --global REGISTRY=ghcr.io/llmos-ai
ARG --global VERSION=main

WORKDIR /llmos

build-airgap:
    ARG TARGETARCH # system arg
    FROM $ALPINE_DIND
    RUN apk add --no-cache curl zstd bash envsubst yq
    RUN echo "Downloading k3s version: ${K3S_VERSION}"
    COPY scripts  ./scripts
    WITH DOCKER
        RUN bash ./scripts/package-airgap
    END
    SAVE ARTIFACT dist/artifacts AS LOCAL dist/artifacts
    SAVE IMAGE --cache-from ${REGISTRY}/llmos-airgap:${VERSION} --push ${REGISTRY}/llmos-airgap:${VERSION}

build-models:
    ARG TARGETARCH # system arg
    FROM alpine:$ALPINE
    ARG OLLAMA_MODELS=dist/models
    ENV OLLAMA_MODELS=${OLLAMA_MODELS}
    RUN apk add --no-cache curl bash gcompat build-base tar zstd
    RUN echo "Downloading ollama version: ${OLLAMA_VERSION}-${TARGETARCH} "
    RUN curl -sfL https://ollama.com/download/ollama-linux-${TARGETARCH}?version=${OLLAMA_VERSION} -o /usr/bin/ollama
    RUN chmod +x /usr/bin/ollama
    COPY scripts  ./scripts
    RUN ./scripts/pull-models
    #SAVE ARTIFACT dist/models AS LOCAL dist/models
    SAVE IMAGE --cache-from ${REGISTRY}/llmos-models:${VERSION} --push ${REGISTRY}/llmos-models:${VERSION}

build-repo:
    FROM nginx:alpine$ALPINE
    ARG GIT_REPO
    ENV GIT_REPO=${GIT_REPO}
    RUN apk update && apk add git helm yq jq bash
    COPY scripts ./scripts
    RUN ./scripts/package-charts
    RUN mv dist/charts /usr/share/nginx/html/charts
    RUN ls -l /usr/share/nginx/charts || true
    RUN rm -rf /dist
    EXPOSE 80
    CMD ["nginx", "-g", "daemon off;"]
    SAVE ARTIFACT /usr/share/nginx/html/charts AS LOCAL dist/charts
    SAVE IMAGE --cache-from ${REGISTRY}/llmos-repo:${VERSION} --push ${REGISTRY}/llmos-repo:${VERSION}
