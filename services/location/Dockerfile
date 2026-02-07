# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY proto/go.mod proto/go.sum ./proto/
COPY services/location/go.mod services/location/go.sum ./services/location/
WORKDIR /src/services/location
RUN go mod edit -replace github.com/daffahilmyf/ride-hailing/proto=../../proto
RUN go mod download

WORKDIR /src
COPY proto ./proto
COPY services/location ./services/location

WORKDIR /src/services/location
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/location ./cmd/location

# Runtime stage
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /out/location /location
COPY --from=builder /src/services/location/config /config
USER nonroot:nonroot
ENTRYPOINT ["/location", "serve", "--config", "/config/config.yaml"]
