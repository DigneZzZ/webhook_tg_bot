package main

import (
	"log"

	"webhook_tg_bot/internal/config"

	"github.com/joho/godotenv"
)

// Проверка маршрутизации категорий: мониторинг + thread ID
func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load .env: %v", err)
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	names := map[int]string{
		30: "Docker", 31: "ОС", 32: "Разработка и ИИ",
		18: "Умный дом", 33: "Информационная безопасность", 34: "Сети и оборудование",
	}
	for _, id := range []int{30, 31, 32, 18, 33, 34} {
		log.Printf("category %d (%s): monitored=%v threadID=%d",
			id, names[id], cfg.ShouldMonitorCategory(id), cfg.GetThreadIDForCategory(id))
	}
}
