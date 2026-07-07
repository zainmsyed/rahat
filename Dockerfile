FROM golang:1.23 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/rahat-api ./cmd/server

FROM debian:bookworm-slim
WORKDIR /app

ENV APP_ENV=production \
    RAHAT_HTTP_ADDR=:8080 \
    DATABASE_PATH=/data/rahat.sqlite3

COPY --from=builder /out/rahat-api /usr/local/bin/rahat-api

EXPOSE 8080
CMD ["/usr/local/bin/rahat-api"]
