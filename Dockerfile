# --- Build stage ---
FROM golang:1.23.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main .

# --- Final stage ---
FROM debian:bullseye-slim

WORKDIR /app

# Устанавливаем curl, tar, postgresql-client
RUN apt-get update && apt-get install -y \
  curl tar ca-certificates postgresql-client \
  && rm -rf /var/lib/apt/lists/*

# Скачиваем и устанавливаем golang-migrate из .tar.gz
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz \
  -o migrate.tar.gz && \
  tar -xzf migrate.tar.gz && \
  mv migrate /usr/local/bin/migrate && \
  chmod +x /usr/local/bin/migrate && \
  rm migrate.tar.gz

# Копируем артефакты из builder
COPY --from=builder /app/main /app/main
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/entrypoint.sh ./entrypoint.sh
COPY .env .env

# Устанавливаем права
RUN chmod +x /app/entrypoint.sh && chmod +x /app/main

EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["/app/main"]
