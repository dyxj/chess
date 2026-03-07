# Build binary
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o game-server ./cmd/game-server

# Build image
FROM alpine:3.23

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /app/game-server ./game-server

USER app

EXPOSE 8080

# Requires environment variables for configuration.
#  -e HOST=0.0.0.0 \
#  -e PORT=8080 \
#  -e READ_HEADER_TIMEOUT=5s \
#  -e READ_TIMEOUT=10s \
#  -e IDLE_TIMEOUT=60s \
#  -e HANDLER_TIMEOUT=30s \
#  -e SHUT_DOWN_TIMEOUT=10s \
#  -e SHUT_DOWN_HARD_TIMEOUT=15s \
#  -e SHUT_DOWN_READY_DELAY=5s \
ENTRYPOINT ["./game-server"]
