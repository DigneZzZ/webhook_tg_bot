package announcements

import (
	"fmt"
	"log"
	"strings"
	"webhook_tg_bot/internal/ai"
	"webhook_tg_bot/internal/config"
	"webhook_tg_bot/internal/discourse"
	"webhook_tg_bot/internal/models"
)

// AnnouncementService сервис для создания анонсов
type AnnouncementService struct {
	config          *config.Config
	discourseClient *discourse.Client
	ai              ai.AIProvider
}

// NewAnnouncementService создает новый сервис анонсов
func NewAnnouncementService(cfg *config.Config) (*AnnouncementService, error) {
	// Проверяем, включены ли анонсы
	if !cfg.EnableAnnouncements {
		return nil, fmt.Errorf("announcements are disabled")
	}

	// Инициализируем AI провайдер
	aiProvider, err := ai.NewProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI provider: %v", err)
	}

	// Создаем клиент Discourse только если все настройки заданы
	var discourseClient *discourse.Client
	if cfg.DiscourseAPIKey != "" && cfg.DiscourseAPIUsername != "" {
		client, err := discourse.NewClient(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create Discourse client: %v", err)
		}

		// Тестируем подключение
		if err := client.TestConnection(); err != nil {
			log.Printf("Warning: Discourse API connection test failed: %v", err)
			return nil, fmt.Errorf("failed to connect to Discourse API: %v", err)
		}

		discourseClient = client
	}

	return &AnnouncementService{
		config:          cfg,
		discourseClient: discourseClient,
		ai:              aiProvider,
	}, nil
}

// ShouldCreateAnnouncement проверяет, нужно ли создавать анонс для этой темы
func (s *AnnouncementService) ShouldCreateAnnouncement(processed *models.ProcessedWebhook) bool {
	log.Printf("[Announcement Check] Topic ID: %d, Title: %s, Category: %d",
		processed.TopicID, processed.TopicTitle, processed.CategoryID)

	// Проверяем, что анонсы включены
	if !s.config.EnableAnnouncements {
		log.Printf("[Announcement Check] ❌ Announcements are disabled")
		return false
	}
	log.Printf("[Announcement Check] ✅ Announcements enabled")

	// Проверяем, что настроена категория для анонсов
	if s.config.AnnouncementCategoryID == 0 {
		log.Printf("[Announcement Check] ❌ Announcement category ID is not set")
		return false
	}
	log.Printf("[Announcement Check] ✅ Announcement category: %d", s.config.AnnouncementCategoryID)

	// Проверяем, что есть клиент Discourse
	if s.discourseClient == nil {
		log.Printf("[Announcement Check] ❌ Discourse client is nil")
		return false
	}
	log.Printf("[Announcement Check] ✅ Discourse client is available")

	// Проверяем, что тема из платной категории
	if !s.config.IsPremiumCategory(processed.CategoryID) {
		log.Printf("[Announcement Check] ❌ Category %d is not premium", processed.CategoryID)
		return false
	}
	log.Printf("[Announcement Check] ✅ Category %d is premium", processed.CategoryID)

	// Не создаем анонс для той же категории анонсов (предотвращаем рекурсию)
	if processed.CategoryID == s.config.AnnouncementCategoryID {
		log.Printf("[Announcement Check] ❌ Skipping - topic is already in announcement category %d", processed.CategoryID)
		return false
	}
	log.Printf("[Announcement Check] ✅ Not in announcement category")

	// Не создаем анонс, если заголовок уже содержит "📢 Анонс:" (дополнительная защита)
	if strings.HasPrefix(processed.TopicTitle, "📢 Анонс:") {
		log.Printf("[Announcement Check] ❌ Skipping - title already has announcement prefix")
		return false
	}
	log.Printf("[Announcement Check] ✅ Title doesn't have announcement prefix")

	log.Printf("[Announcement Check] ✅✅✅ ALL CHECKS PASSED - will create announcement!")
	return true
}

// CreateAnnouncement создает анонс темы в категории анонсов и возвращает URL
func (s *AnnouncementService) CreateAnnouncement(processed *models.ProcessedWebhook, subscriptionInfo *config.SubscriptionInfo) (string, error) {
	log.Printf("[Create Announcement] Starting for topic %d", processed.TopicID)

	if !s.ShouldCreateAnnouncement(processed) {
		log.Printf("[Create Announcement] ❌ ShouldCreateAnnouncement returned false - skipping")
		return "", nil
	}

	log.Printf("[Create Announcement] ✅ Passed all checks, proceeding to create")

	// Валидация данных перед созданием
	if processed.TopicTitle == "" {
		log.Printf("[Create Announcement] ❌ Error: topic title is empty")
		return "", fmt.Errorf("topic title is empty")
	}
	if processed.Author == "" {
		log.Printf("[Create Announcement] ❌ Error: author is empty")
		return "", fmt.Errorf("author is empty")
	}
	if processed.Content == "" {
		log.Printf("[Create Announcement] ⚠️  Warning: Content is empty for topic %d, using title as content", processed.TopicID)
	}

	// Генерируем заголовок анонса
	title := s.generateAnnouncementTitle(processed)
	log.Printf("[Create Announcement] Generated title: %s", title)

	// Генерируем содержание анонса с AI резюме
	log.Printf("[Create Announcement] Generating content with AI...")
	content := s.generateAnnouncementContent(processed, subscriptionInfo)
	log.Printf("[Create Announcement] Content generated (length: %d chars)", len(content))

	log.Printf("[Create Announcement] 🚀 Creating announcement for topic %d (%s) in category %d",
		processed.TopicID, processed.TopicTitle, s.config.AnnouncementCategoryID)

	// Создаем тему в Discourse
	response, err := s.discourseClient.CreateTopic(
		s.config.AnnouncementCategoryID,
		title,
		content,
		[]string{"анонс", "премиум"},
	)
	if err != nil {
		log.Printf("[Create Announcement] ❌ FAILED to create topic: %v", err)
		return "", fmt.Errorf("failed to create announcement topic: %v", err)
	}

	// Формируем URL анонса
	announcementURL := fmt.Sprintf("%s/t/%s/%d", s.config.DiscourseBaseURL, response.TopicSlug, response.TopicID)
	log.Printf("[Create Announcement] ✅ SUCCESS! Created announcement topic: %s (ID: %d, URL: %s)", title, response.TopicID, announcementURL)
	return announcementURL, nil
}

// generateAnnouncementTitle генерирует заголовок для анонса
func (s *AnnouncementService) generateAnnouncementTitle(processed *models.ProcessedWebhook) string {
	prefix := "📢 Анонс: "
	// Discourse обычно ограничивает заголовки до 255 символов, но лучше использовать 200 для безопасности
	maxTotalLength := 200
	maxTitleLength := maxTotalLength - len(prefix)

	title := processed.TopicTitle

	// Обрезаем заголовок если он слишком длинный, учитывая место для "..."
	if len(title) > maxTitleLength {
		title = title[:maxTitleLength-3] + "..."
	}

	return prefix + title
}

// generateAnnouncementContent генерирует содержание анонса с безопасной логикой
func (s *AnnouncementService) generateAnnouncementContent(processed *models.ProcessedWebhook, subscriptionInfo *config.SubscriptionInfo) string {
	// Определяем префикс роли автора
	var roleEmoji string
	switch processed.AuthorRole {
	case "admin":
		roleEmoji = "👑"
	case "moderator":
		roleEmoji = "🛡️"
	case "staff":
		roleEmoji = "⭐"
	case "leader":
		roleEmoji = "🔥"
	default:
		roleEmoji = "👤"
	}

	// Формируем теги
	tagsStr := "нет"
	if len(processed.Tags) > 0 {
		tags := make([]string, len(processed.Tags))
		for i, tag := range processed.Tags {
			tags[i] = "#" + tag
		}
		tagsStr = strings.Join(tags, ", ")
	}

	// Генерируем только AI резюме (безопасно, так как ИИ делает краткую выжимку)
	aiSummary, err := s.ai.GenerateSummary(processed.Content, processed.TopicTitle, processed.AuthorRole, processed.Category)
	if err != nil {
		log.Printf("Failed to generate AI summary for announcement: %v", err)
		// Fallback: безопасная заглушка без раскрытия контента
		aiSummary = "Новая тема в премиум разделе. Подробности доступны только подписчикам."
	}

	// Используем переданную информацию о подписке или fallback
	var subscriptionName, subscriptionText string
	if subscriptionInfo != nil {
		subscriptionName = subscriptionInfo.Name
		subscriptionText = subscriptionInfo.BotText
	} else {
		// Fallback на дефолтные значения
		subscriptionName = "VIP подписку"
		subscriptionText = "Оформить VIP можно в тг-боте: https://t.me/gig_combot"
	}

	content := fmt.Sprintf(`## %s %s создал новую тему в премиум разделе

**Название темы:** %s

**📋 Краткое описание:**
%s

**🏷 Теги:** %s  
**📂 Категория:** %s

---

## 💎 О премиум контенте

Это анонс платного контента. Полная тема со всеми материалами, комментариями и обсуждением доступна только подписчикам с доступом **"%s"**.

### 🚀 Что даёт эта подписка:
- Доступ к премиум разделам по данной теме
- Эксклюзивные материалы от экспертов
- Участие в закрытых обсуждениях
- Первоочередная поддержка

### 🤖 Как получить доступ:
%s

🔗 **[Перейти к оригинальной теме](%s)** (требуется подписка)

---

💬 **Есть вопросы о теме?** Задавайте их в комментариях к этому анонсу! Мы ответим на общие вопросы, а для получения полной информации рекомендуем оформить подписку.`,
		roleEmoji,
		processed.Author,
		processed.TopicTitle,
		aiSummary,
		tagsStr,
		processed.Category,
		subscriptionName,
		subscriptionText,
		processed.URL,
	)

	return content
}

// IsAnnouncementCategory проверяет, является ли категория категорией анонсов
func (s *AnnouncementService) IsAnnouncementCategory(categoryID int) bool {
	return s.config.AnnouncementCategoryID == categoryID
}
