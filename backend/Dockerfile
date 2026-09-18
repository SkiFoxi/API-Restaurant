# Первый этап: сборка бекенда
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
RUN go build -o /avito-backend ./cmd/app

# Второй этап: финальный образ
FROM alpine:latest

# Устанавливаем ca-certificates (для HTTPS) и bash (опционально)
RUN apk add --no-cache ca-certificates bash

# Копируем бинарник бекенда
COPY --from=builder /avito-backend /avito-backend

#Место под линтер в контейнере
COPY --from=builder /go/bin/golangci-lint /usr/local/bin/golangci-lint

# Копируем миграции в образ (папка /migrations)
COPY migrations /migrations

# Копируем утилиту migrate из официального образа
COPY --from=migrate/migrate:v4.17.0 /usr/bin/migrate /usr/bin/migrate

# Делаем бинарник бекенда исполняемым
RUN chmod +x /avito-backend

EXPOSE 8080
CMD ["/avito-backend"]