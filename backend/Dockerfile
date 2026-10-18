# Build stage
# Must be >= the `go` directive in go.mod (1.25.0).
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8000

# /health is a liveness probe only: a missing database must not restart the
# container. Readiness (/readyz) is what orchestrators should route on.
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
    CMD wget -qO- "http://127.0.0.1:${PORT:-8000}/health" >/dev/null || exit 1

CMD ["./server"]
