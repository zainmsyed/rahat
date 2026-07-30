# syntax=docker/dockerfile:1

# --- Stage 1: build the SvelteKit frontend ---
FROM node:22 AS web-builder
WORKDIR /app/web

ARG VITE_API_BASE_URL=""
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}

COPY web/package.json web/package-lock.json ./
RUN npm ci

COPY web/ ./
RUN npm run build

# --- Stage 2: build the Go backend ---
FROM golang:1.25 AS go-builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY db ./db
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/rahat-api ./cmd/server

# --- Stage 3: production runtime image ---
FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN groupadd -r rahat && useradd -r -g rahat rahat

WORKDIR /app

ENV APP_ENV=production \
    RAHAT_HTTP_ADDR=:8080 \
    DATABASE_PATH=/data/rahat.sqlite3 \
    WEB_STATIC_DIR=/app/web/static

COPY --from=go-builder /out/rahat-api /usr/local/bin/rahat-api
COPY --from=go-builder /app/db/migrations /app/db/migrations
COPY --from=web-builder /app/web/build /app/web/static

RUN mkdir -p /data && chown rahat:rahat /data

USER rahat

EXPOSE 8080

CMD ["/usr/local/bin/rahat-api"]
