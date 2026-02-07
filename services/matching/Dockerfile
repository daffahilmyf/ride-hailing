# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY proto/go.mod proto/go.sum ./proto/
COPY services/matching/go.mod services/matching/go.sum ./services/matching/
WORKDIR /src/services/matching
RUN go mod edit -replace github.com/daffahilmyf/ride-hailing/proto=../../proto
RUN go mod download

WORKDIR /src
COPY proto ./proto
COPY services/matching ./services/matching

WORKDIR /src/services/matching
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/matching ./cmd/matching

# Runtime stage
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /out/matching /matching
COPY --from=builder /src/services/matching/config /config
USER nonroot:nonroot
ENTRYPOINT ["/matching", "serve", "--config", "/config/config.yaml"]
