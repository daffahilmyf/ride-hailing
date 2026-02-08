# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY proto/go.mod proto/go.sum ./proto/
COPY services/user/go.mod services/user/go.sum ./services/user/
WORKDIR /src/services/user
RUN go mod edit -replace github.com/daffahilmyf/ride-hailing/proto=../../proto
RUN go mod download

WORKDIR /src
COPY proto ./proto
COPY services/user ./services/user

WORKDIR /src/services/user
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/user ./cmd/user

# Runtime stage
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /out/user /user
COPY --from=builder /src/services/user/config /config
USER nonroot:nonroot
ENTRYPOINT ["/user", "serve", "--config", "/config/config.yaml"]
