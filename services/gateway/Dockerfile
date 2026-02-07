# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY proto/go.mod proto/go.sum ./proto/
COPY services/gateway/go.mod services/gateway/go.sum ./services/gateway/
WORKDIR /src/services/gateway
RUN go mod edit -replace github.com/daffahilmyf/ride-hailing/proto=../../proto
RUN go mod download

WORKDIR /src
COPY proto ./proto
COPY services/gateway ./services/gateway

WORKDIR /src/services/gateway
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gateway ./cmd/gateway

# Runtime stage
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /out/gateway /gateway
COPY --from=builder /src/services/gateway/config /config
USER nonroot:nonroot
ENTRYPOINT ["/gateway", "serve", "--config", "/config/config.yaml"]
