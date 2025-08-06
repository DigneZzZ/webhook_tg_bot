package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Telegram settings
	TelegramBotToken string
	TelegramChatID   int64
	TelegramThreadID int

	// Additional threads mapping
	CategoryThreads map[int]int // category_id -> thread_id

	// Webhook settings
	WebhookSecret string
	WebhookPort   string
	WebhookPath   string

	// AI settings
	OpenAIAPIKey string
	OpenAIModel  string

	// Premium categories (paid sections)
	PremiumCategories []int

	// Categories to monitor (if empty, monitor all except ignored)
	MonitoredCategories []int

	// Categories to ignore (priority over monitored)
	IgnoredCategories []int

	// User IDs to ignore (including negative IDs for bots)
	IgnoredUsers []int

	// Base URL for topics
	BaseURL string
}

func Load() (*Config, error) {
	cfg := &Config{}

	// Telegram settings
	cfg.TelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	if cfg.TelegramBotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	chatIDStr := os.Getenv("TELEGRAM_CHAT_ID")
	if chatIDStr == "" {
		return nil, fmt.Errorf("TELEGRAM_CHAT_ID is required")
	}
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid TELEGRAM_CHAT_ID: %v", err)
	}
	cfg.TelegramChatID = chatID

	threadIDStr := os.Getenv("TELEGRAM_THREAD_ID")
	if threadIDStr != "" {
		threadID, err := strconv.Atoi(threadIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid TELEGRAM_THREAD_ID: %v", err)
		}
		cfg.TelegramThreadID = threadID
	}

	// Загружаем дополнительные thread'ы
	cfg.CategoryThreads = make(map[int]int)
	for i := 1; i <= 5; i++ {
		threadIDKey := fmt.Sprintf("TELEGRAM_THREAD_ID_%d", i)
		categoriesKey := fmt.Sprintf("THREAD_CATEGORIES_%d", i)

		threadIDStr := os.Getenv(threadIDKey)
		categoriesStr := os.Getenv(categoriesKey)

		if threadIDStr != "" && categoriesStr != "" {
			threadID, err := strconv.Atoi(threadIDStr)
			if err != nil {
				return nil, fmt.Errorf("invalid %s: %v", threadIDKey, err)
			}

			categories := strings.Split(categoriesStr, ",")
			for _, catStr := range categories {
				catStr = strings.TrimSpace(catStr)
				if catStr != "" {
					categoryID, err := strconv.Atoi(catStr)
					if err != nil {
						return nil, fmt.Errorf("invalid category ID '%s' in %s: %v", catStr, categoriesKey, err)
					}
					cfg.CategoryThreads[categoryID] = threadID
				}
			}
		}
	}

	// Webhook settings
	cfg.WebhookSecret = os.Getenv("WEBHOOK_SECRET")
	if cfg.WebhookSecret == "" {
		return nil, fmt.Errorf("WEBHOOK_SECRET is required")
	}

	cfg.WebhookPort = os.Getenv("WEBHOOK_PORT")
	if cfg.WebhookPort == "" {
		cfg.WebhookPort = "8080"
	}

	cfg.WebhookPath = os.Getenv("WEBHOOK_PATH")
	if cfg.WebhookPath == "" {
		cfg.WebhookPath = "/webhook"
	}

	// AI settings
	cfg.OpenAIAPIKey = os.Getenv("OPENAI_API_KEY")
	cfg.OpenAIModel = os.Getenv("OPENAI_MODEL")
	if cfg.OpenAIModel == "" {
		cfg.OpenAIModel = "gpt-4.1-nano"
	}

	// Premium categories
	premiumCategoriesStr := os.Getenv("PREMIUM_CATEGORIES")
	if premiumCategoriesStr != "" {
		categoryIDs := strings.Split(premiumCategoriesStr, ",")
		for _, categoryID := range categoryIDs {
			if id, err := strconv.Atoi(strings.TrimSpace(categoryID)); err == nil {
				cfg.PremiumCategories = append(cfg.PremiumCategories, id)
			}
		}
	}

	// Monitored categories
	monitoredCategoriesStr := os.Getenv("MONITORED_CATEGORIES")
	if monitoredCategoriesStr != "" {
		categoryIDs := strings.Split(monitoredCategoriesStr, ",")
		for _, categoryID := range categoryIDs {
			if id, err := strconv.Atoi(strings.TrimSpace(categoryID)); err == nil {
				cfg.MonitoredCategories = append(cfg.MonitoredCategories, id)
			}
		}
	}

	// Ignored categories
	ignoredCategoriesStr := os.Getenv("IGNORED_CATEGORIES")
	if ignoredCategoriesStr != "" {
		categoryIDs := strings.Split(ignoredCategoriesStr, ",")
		for _, categoryID := range categoryIDs {
			if id, err := strconv.Atoi(strings.TrimSpace(categoryID)); err == nil {
				cfg.IgnoredCategories = append(cfg.IgnoredCategories, id)
			}
		}
	}

	// Ignored users
	ignoredUsersStr := os.Getenv("IGNORED_USERS")
	if ignoredUsersStr != "" {
		userIDs := strings.Split(ignoredUsersStr, ",")
		for _, userID := range userIDs {
			if id, err := strconv.Atoi(strings.TrimSpace(userID)); err == nil {
				cfg.IgnoredUsers = append(cfg.IgnoredUsers, id)
			}
		}
	}

	// Base URL
	cfg.BaseURL = os.Getenv("BASE_URL")
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://your-forum.com"
	}

	return cfg, nil
}

// IsPremiumCategory проверяет, является ли категория платной
func (cfg *Config) IsPremiumCategory(categoryID int) bool {
	for _, premiumID := range cfg.PremiumCategories {
		if premiumID == categoryID {
			return true
		}
	}
	return false
}

// ShouldMonitorCategory проверяет, нужно ли отслеживать категорию
func (cfg *Config) ShouldMonitorCategory(categoryID int) bool {
	// Сначала проверяем игнорируемые категории (приоритет)
	for _, ignoredID := range cfg.IgnoredCategories {
		if ignoredID == categoryID {
			return false
		}
	}

	// Если список отслеживаемых категорий пуст, отслеживаем все (кроме игнорируемых)
	if len(cfg.MonitoredCategories) == 0 {
		return true
	}

	// Проверяем, есть ли категория в списке отслеживаемых
	for _, monitoredID := range cfg.MonitoredCategories {
		if monitoredID == categoryID {
			return true
		}
	}

	return false
}

// GetThreadIDForCategory возвращает thread ID для указанной категории
func (cfg *Config) GetThreadIDForCategory(categoryID int) int {
	if threadID, exists := cfg.CategoryThreads[categoryID]; exists {
		return threadID
	}
	return cfg.TelegramThreadID // возвращаем дефолтный thread ID
}

// ShouldIgnoreUser проверяет, нужно ли игнорировать пользователя
func (cfg *Config) ShouldIgnoreUser(userID int) bool {
	for _, ignoredUserID := range cfg.IgnoredUsers {
		if ignoredUserID == userID {
			return true
		}
	}
	return false
}
