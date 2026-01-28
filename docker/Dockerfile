# Multi-stage build for envtest container
# Supports multiple Kubernetes versions via build arg

ARG GO_VERSION=1.22
ARG ALPINE_VERSION=3.19
ARG KUBERNETES_VERSION=1.31.0

# Stage 1: Download envtest binaries using setup-envtest
FROM golang:${GO_VERSION}-alpine AS builder

ARG KUBERNETES_VERSION
ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache bash curl

# Install setup-envtest
RUN go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

# Download envtest binaries for the specified Kubernetes version
RUN setup-envtest use ${KUBERNETES_VERSION} --bin-dir /envtest-bins -p path

# Stage 2: Runtime image
FROM alpine:${ALPINE_VERSION}

ARG KUBERNETES_VERSION

LABEL org.opencontainers.image.title="testcontainers-envtest"
LABEL org.opencontainers.image.description="Envtest container for Kubernetes controller testing"
LABEL org.opencontainers.image.source="https://github.com/roma-glushko/testcontainers-envtest"
LABEL org.opencontainers.image.version="${KUBERNETES_VERSION}"

# Install runtime dependencies
RUN apk add --no-cache bash curl jq openssl

# Copy envtest binaries from builder
COPY --from=builder /envtest-bins /usr/local/bin/envtest

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Environment variables
ENV KUBERNETES_VERSION=${KUBERNETES_VERSION}
ENV API_SERVER_PORT=6443
ENV ETCD_PORT=2379
ENV KUBECONFIG_PATH=/tmp/kubeconfig

# Expose API server port
EXPOSE 6443

# Health check - verify API server is responding
HEALTHCHECK --interval=5s --timeout=3s --start-period=10s --retries=3 \
    CMD curl -sk https://localhost:${API_SERVER_PORT}/healthz || exit 1

ENTRYPOINT ["/entrypoint.sh"]
