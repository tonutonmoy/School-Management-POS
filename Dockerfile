FROM golang:1.24-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go run github.com/a-h/templ/cmd/templ@latest generate
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.20
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata postgresql16-client wget

COPY --from=builder /server /app/server
COPY migrations /app/migrations
COPY web/static /app/web/static

RUN mkdir -p /app/backups && chown -R nobody:nobody /app/backups

ENV APP_PORT=8085
ENV BACKUP_DIR=/app/backups
EXPOSE 8085

HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
  CMD wget -qO- http://127.0.0.1:${APP_PORT}/health || exit 1

USER nobody
CMD ["/app/server"]
