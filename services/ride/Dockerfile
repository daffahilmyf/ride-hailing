# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY proto/go.mod proto/go.sum ./proto/
COPY services/ride/go.mod services/ride/go.sum ./services/ride/
WORKDIR /src/services/ride
RUN go mod edit -replace github.com/daffahilmyf/ride-hailing/proto=../../proto
RUN go mod download

WORKDIR /src
COPY proto ./proto
COPY services/ride ./services/ride

WORKDIR /src/services/ride
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/ride ./cmd/ride

# Runtime stage
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /out/ride /ride
COPY --from=builder /src/services/ride/config /config
USER nonroot:nonroot
ENTRYPOINT ["/ride", "serve", "--config", "/config/config.yaml"]
