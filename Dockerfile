# ── Stage 1: build ────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder
ARG VERSION=dev
WORKDIR /build
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags="-s -w -X github.com/truedem0n/playbridge-stream-resolver/version.Version=${VERSION}" \
    -o stream-addons-resolver .

# ── Stage 2: runtime ──────────────────────────────────────────────────────────
FROM alpine:3.21
# ffmpeg ships ffprobe; ca-certificates needed for HTTPS addon requests
RUN apk add --no-cache ffmpeg ca-certificates tzdata

WORKDIR /app
COPY --from=builder /build/stream-addons-resolver .
COPY config.example.json .

# /app/config.json  — mount your config here (required)
# /app/cache        — mount a volume here for persistent cache
EXPOSE 7001

ENTRYPOINT ["./stream-addons-resolver", "-config", "/app/config.json"]
