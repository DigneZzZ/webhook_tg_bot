FROM golang:1.23-alpine AS builder

WORKDIR /app

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git ca-certificates tzdata

# Копируем go mod и sum файлы
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение с оптимизацией
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o webhook_tg_bot .

# Финальный образ
FROM alpine:latest

# Устанавливаем необходимые пакеты
RUN apk --no-cache add ca-certificates tzdata curl

WORKDIR /root/

# Копируем исполняемый файл из builder
COPY --from=builder /app/webhook_tg_bot .

# Копируем timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Создаем пользователя без root прав
RUN adduser -D -s /bin/sh webhook
USER webhook
WORKDIR /home/webhook

COPY --from=builder --chown=webhook:webhook /app/webhook_tg_bot .

# Используем переменную окружения для порта
ENV WEBHOOK_PORT=8080
EXPOSE $WEBHOOK_PORT

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:${WEBHOOK_PORT}/health || exit 1

# Команда запуска
CMD ["./webhook_tg_bot"]
