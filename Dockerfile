FROM golang:1.21-alpine AS builder

WORKDIR /app

# Копируем go mod и sum файлы
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o webhook_tg_bot .

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Копируем исполняемый файл из builder
COPY --from=builder /app/webhook_tg_bot .

# Открываем порт
EXPOSE 8080

# Команда запуска
CMD ["./webhook_tg_bot"]
