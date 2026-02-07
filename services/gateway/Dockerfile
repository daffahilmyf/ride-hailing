# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gateway ./cmd/gateway

# Runtime stage
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /out/gateway /gateway
COPY --from=builder /src/config /config
USER nonroot:nonroot
ENTRYPOINT ["/gateway", "serve"]
