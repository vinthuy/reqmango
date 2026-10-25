# Build stage
# Must be >= the `go` directive in go.mod (1.25.0) and the same minor as
# GO_VERSION in .github/workflows/ci.yml, so the compiler that builds the image
# is the one CI tested with. The tag floats on the patch release on purpose:
# `golang:1.25-alpine` currently resolves to 1.25.14-alpine, exactly like the
# `1.25` setup-go pin, so a patch CVE is picked up without a commit.
#
# Cross-compilation: buildx runs this stage on the build platform and Go
# produces the target binary, so a multi-arch release does not pay for QEMU.
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
# Injected into internal/version (see `GET /version`).
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -trimpath \
      -ldflags="-s -w \
        -X github.com/reqmango/backend/internal/version.Version=${VERSION} \
        -X github.com/reqmango/backend/internal/version.Commit=${COMMIT} \
        -X github.com/reqmango/backend/internal/version.BuildDate=${BUILD_DATE}" \
      -o /app/server ./cmd/server

# Runtime stage
# The binary is CGO_ENABLED=0, so the only things this layer adds are
# ca-certificates and tzdata. Kept on the same Alpine release as the builder
# (3.24): 3.20 stopped receiving packages in 2026 and the image was six months
# behind on base-layer CVE fixes.
FROM alpine:3.24

ARG VERSION=dev
ARG COMMIT=unknown

LABEL org.opencontainers.image.title="reqmango-backend" \
      org.opencontainers.image.description="Reqmango API server" \
      org.opencontainers.image.source="https://github.com/vinthuy/reqmango" \
      org.opencontainers.image.url="https://github.com/vinthuy/reqmango" \
      org.opencontainers.image.documentation="https://github.com/vinthuy/reqmango/blob/master/README.md" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}"

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8000

# /health is a liveness probe only: a missing database must not restart the
# container. Readiness (/readyz) is what orchestrators should route on.
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
    CMD wget -qO- "http://127.0.0.1:${PORT:-8000}/health" >/dev/null || exit 1

CMD ["./server"]
