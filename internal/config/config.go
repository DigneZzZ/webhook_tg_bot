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

	// Debug settings
	Debug bool

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

	// Discourse API settings for creating topics
	DiscourseAPIKey      string
	DiscourseAPIUsername string
	DiscourseBaseURL     string

	// Announcement settings
	AnnouncementCategoryID int  // Category ID where to post announcements
	EnableAnnouncements    bool // Whether to enable automatic announcements

	// Subscription info mapping: category_id -> subscription info
	CategorySubscriptions map[int]*SubscriptionInfo
}

// SubscriptionInfo содержит информацию о подписке для группы категорий
type SubscriptionInfo struct {
	Name    string // Название подписки (например "VIP подписка", "Доступ к SHM")
	BotText string // Текст для бота (например "Оформить VIP можно в тг-боте: https://t.me/gig_combot")
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
		cfg.OpenAIModel = "gpt-5-mini"
	}

	// Debug settings
	debugStr := os.Getenv("DEBUG")
	cfg.Debug = debugStr == "true" || debugStr == "1"

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

	// Discourse API settings
	cfg.DiscourseAPIKey = os.Getenv("DISCOURSE_API_KEY")
	cfg.DiscourseAPIUsername = os.Getenv("DISCOURSE_API_USERNAME")
	cfg.DiscourseBaseURL = os.Getenv("DISCOURSE_BASE_URL")
	if cfg.DiscourseBaseURL == "" {
		cfg.DiscourseBaseURL = cfg.BaseURL // fallback to BASE_URL
	}

	// Announcement settings
	announcementCategoryStr := os.Getenv("ANNOUNCEMENT_CATEGORY_ID")
	if announcementCategoryStr != "" {
		if id, err := strconv.Atoi(announcementCategoryStr); err == nil {
			cfg.AnnouncementCategoryID = id
		}
	}

	enableAnnouncementsStr := os.Getenv("ENABLE_ANNOUNCEMENTS")
	cfg.EnableAnnouncements = enableAnnouncementsStr == "true" || enableAnnouncementsStr == "1"

	// Category subscriptions mapping
	cfg.CategorySubscriptions = make(map[int]*SubscriptionInfo)
	// Формат:
	// SUBSCRIPTION_NAME_1=VIP подписка
	// SUBSCRIPTION_CATEGORIES_1=10,11,12,13,22,23,24
	// SUBSCRIPTION_BOT_TEXT_1=Оформить VIP можно в тг-боте: https://t.me/gig_combot
	//
	// SUBSCRIPTION_NAME_2=Доступ к материалам SHM
	// SUBSCRIPTION_CATEGORIES_2=5
	// SUBSCRIPTION_BOT_TEXT_2=Приобрести доступ к SHM можно в боте: https://t.me/gig_combot
	for i := 1; i <= 10; i++ { // поддерживаем до 10 разных подписок
		nameKey := fmt.Sprintf("SUBSCRIPTION_NAME_%d", i)
		categoriesKey := fmt.Sprintf("SUBSCRIPTION_CATEGORIES_%d", i)
		botTextKey := fmt.Sprintf("SUBSCRIPTION_BOT_TEXT_%d", i)

		subscriptionName := os.Getenv(nameKey)
		categoriesStr := os.Getenv(categoriesKey)
		botText := os.Getenv(botTextKey)

		if subscriptionName != "" && categoriesStr != "" && botText != "" {
			subscriptionInfo := &SubscriptionInfo{
				Name:    subscriptionName,
				BotText: botText,
			}

			categories := strings.Split(categoriesStr, ",")
			for _, catStr := range categories {
				catStr = strings.TrimSpace(catStr)
				if catStr != "" {
					categoryID, err := strconv.Atoi(catStr)
					if err != nil {
						return nil, fmt.Errorf("invalid category ID '%s' in %s: %v", catStr, categoriesKey, err)
					}
					cfg.CategorySubscriptions[categoryID] = subscriptionInfo
				}
			}
		}
	}

	return cfg, nil
}

// IsPremiumCategory проверяет, является ли категория платной
func (cfg *Config) IsPremiumCategory(categoryID int) bool {
	// Проверяем новый маппинг подписок
	if _, exists := cfg.CategorySubscriptions[categoryID]; exists {
		return true
	}

	// Fallback на старый способ для обратной совместимости
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

// GetSubscriptionInfo возвращает информацию о подписке для категории
func (cfg *Config) GetSubscriptionInfo(categoryID int) *SubscriptionInfo {
	return cfg.CategorySubscriptions[categoryID]
}
