package main

import (
	"log"

	"webhook_tg_bot/internal/ai"
	"webhook_tg_bot/internal/config"

	"github.com/joho/godotenv"
)

// Тестовый харнесс: воспроизводит путь bot.SendCompleteNotification -> ai.GenerateSummary
// для одного поста, чтобы увидеть реальную ошибку OpenAI в логах.
func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load .env: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	provider, err := ai.NewProvider(cfg)
	if err != nil {
		log.Fatalf("failed to create AI provider: %v", err)
	}

	content := "![screenshot](upload://abc123.png)\n\nВсем привет! Написал скрипт для автоматического бэкапа конфигов Docker-контейнеров. Делюсь кодом и инструкцией по установке.\n\n```bash\n./backup.sh\n```"

	summary, err := provider.GenerateSummary(
		content,
		"Автобэкап Docker-контейнеров",
		"user",
		"Docker",
	)
	if err != nil {
		log.Printf("Failed to generate AI summary, falling back to neutral phrase: %v", err)
		summary = "Подробности — по ссылке ниже"
	}

	log.Printf("Итог: бот отправит '📋 %s'", summary)
}
