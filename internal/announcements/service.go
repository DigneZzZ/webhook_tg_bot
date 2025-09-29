package announcements

import (
	"fmt"
	"log"
	"strings"
	"webhook_tg_bot/internal/config"
	"webhook_tg_bot/internal/discourse"
	"webhook_tg_bot/internal/models"
)

// AnnouncementService сервис для создания анонсов
type AnnouncementService struct {
	config          *config.Config
	discourseClient *discourse.Client
}

// NewAnnouncementService создает новый сервис анонсов
func NewAnnouncementService(cfg *config.Config) (*AnnouncementService, error) {
	// Проверяем, включены ли анонсы
	if !cfg.EnableAnnouncements {
		return nil, fmt.Errorf("announcements are disabled")
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
	}, nil
}

// ShouldCreateAnnouncement проверяет, нужно ли создавать анонс для этой темы
func (s *AnnouncementService) ShouldCreateAnnouncement(processed *models.ProcessedWebhook) bool {
	// Проверяем, что анонсы включены
	if !s.config.EnableAnnouncements {
		return false
	}

	// Проверяем, что настроена категория для анонсов
	if s.config.AnnouncementCategoryID == 0 {
		return false
	}

	// Проверяем, что есть клиент Discourse
	if s.discourseClient == nil {
		return false
	}

	// Проверяем, что тема из платной категории
	if !s.config.IsPremiumCategory(processed.CategoryID) {
		return false
	}

	// Не создаем анонс для той же категории анонсов (предотвращаем рекурсию)
	if processed.CategoryID == s.config.AnnouncementCategoryID {
		return false
	}

	// Не создаем анонс, если заголовок уже содержит "📢 Анонс:" (дополнительная защита)
	if strings.HasPrefix(processed.TopicTitle, "📢 Анонс:") {
		return false
	}

	return true
}

// CreateAnnouncement создает анонс темы в категории анонсов и возвращает URL
func (s *AnnouncementService) CreateAnnouncement(processed *models.ProcessedWebhook) (string, error) {
	if !s.ShouldCreateAnnouncement(processed) {
		return "", nil
	}

	// Генерируем заголовок анонса
	title := s.generateAnnouncementTitle(processed)

	// Генерируем содержание анонса
	content := s.generateAnnouncementContent(processed)

	// Создаем тему в Discourse
	response, err := s.discourseClient.CreateTopic(
		s.config.AnnouncementCategoryID,
		title,
		content,
		[]string{"анонс", "премиум"},
	)
	if err != nil {
		return "", fmt.Errorf("failed to create announcement topic: %v", err)
	}

	// Формируем URL анонса
	announcementURL := fmt.Sprintf("%s/t/%s/%d", s.config.DiscourseBaseURL, response.TopicSlug, response.TopicID)
	log.Printf("Created announcement topic: %s (ID: %d, URL: %s)", title, response.TopicID, announcementURL)
	return announcementURL, nil
}

// generateAnnouncementTitle генерирует заголовок для анонса
func (s *AnnouncementService) generateAnnouncementTitle(processed *models.ProcessedWebhook) string {
	// Ограничиваем длину заголовка
	maxLength := 80
	title := processed.TopicTitle
	
	if len(title) > maxLength {
		title = title[:maxLength-3] + "..."
	}

	return fmt.Sprintf("📢 Анонс: %s", title)
}

// generateAnnouncementContent генерирует содержание анонса
func (s *AnnouncementService) generateAnnouncementContent(processed *models.ProcessedWebhook) string {
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

	// Генерируем краткое описание (первые 200 символов контента)
	description := processed.Content
	if len(description) > 200 {
		description = description[:200] + "..."
	}
	description = strings.ReplaceAll(description, "\n", " ")

	content := fmt.Sprintf(`## %s %s создал новую тему в премиум разделе

**Тема:** %s

**Краткое описание:**
%s

**Теги:** %s

---

💎 **Это анонс платного контента.** Полная тема доступна только подписчикам VIP.

🤖 **Как получить доступ:**
• Оформите VIP подписку через наш Telegram бот: https://t.me/gig_combot
• Доступ к контенту предоставляется автоматически после оплаты подписки
• В подписке доступны все премиум разделы форума

🔗 **[Перейти к оригинальной теме](%s)** (требуется VIP подписка)

💬 **Есть вопросы?** Задавайте их в комментариях к этому анонсу!`,
		roleEmoji,
		processed.Author,
		processed.TopicTitle,
		description,
		tagsStr,
		processed.URL,
	)

	return content
}

// IsAnnouncementCategory проверяет, является ли категория категорией анонсов
func (s *AnnouncementService) IsAnnouncementCategory(categoryID int) bool {
	return s.config.AnnouncementCategoryID == categoryID
}