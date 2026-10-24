# Build stage
# Must be >= the `go` directive in go.mod (1.25.0).
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
FROM alpine:3.20

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
