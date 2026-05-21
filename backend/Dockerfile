# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /bin/barq-server ./cmd/server

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates chromium font-noto \
    && adduser -D -u 1000 barq \
    && mkdir -p /app/data \
    && chown -R barq:barq /app

COPY --from=builder /bin/barq-server /usr/local/bin/barq-server

# Copy bundled assets (fonts, icons, illustrations).
COPY --chown=barq:barq assets/ /app/assets/

USER barq
WORKDIR /app

EXPOSE 8090

ENV BARQ_SERVER__HOST=0.0.0.0 \
    BARQ_SERVER__PORT=8090 \
    BARQ_LOG__LEVEL=info \
    BARQ_LOG__FORMAT=json

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8090/healthz || exit 1

ENTRYPOINT ["barq-server"]
