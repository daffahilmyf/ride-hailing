# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY services/notify/go.mod services/notify/go.sum ./services/notify/
WORKDIR /src/services/notify
RUN go mod download

WORKDIR /src
COPY services/notify ./services/notify

WORKDIR /src/services/notify
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/notify ./cmd/notify

# Runtime stage
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /out/notify /notify
COPY --from=builder /src/services/notify/config /config
USER nonroot:nonroot
ENTRYPOINT ["/notify", "serve", "--config", "/config/config.yaml"]
