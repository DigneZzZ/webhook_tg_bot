package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"webhook_tg_bot/internal/bot"
	"webhook_tg_bot/internal/config"
	"webhook_tg_bot/internal/server"

	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные окружения из .env файла
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using system environment variables")
	}

	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализируем Telegram бота
	telegramBot, err := bot.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create telegram bot: %v", err)
	}

	// Инициализируем веб-сервер для вебхуков
	webhookServer := server.New(cfg, telegramBot)

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("Starting webhook server on port %s", cfg.WebhookPort)
		if err := webhookServer.Start(); err != nil {
			log.Fatalf("Failed to start webhook server: %v", err)
		}
	}()

	// Ожидаем сигнал завершения
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// Graceful shutdown: даём в полёте находящимся вебхукам дообработаться
	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := webhookServer.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
	log.Println("Server stopped")
}
