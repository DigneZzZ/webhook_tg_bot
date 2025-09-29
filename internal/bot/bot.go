package bot

import (
	"fmt"
	"log"
	"webhook_tg_bot/internal/ai"
	"webhook_tg_bot/internal/config"
	"webhook_tg_bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramBot struct {
	bot    *tgbotapi.BotAPI
	config *config.Config
	ai     ai.AIProvider
}

func New(cfg *config.Config) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %v", err)
	}

	// Инициализируем AI провайдер
	aiProvider, err := ai.NewProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI provider: %v", err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	return &TelegramBot{
		bot:    bot,
		config: cfg,
		ai:     aiProvider,
	}, nil
}

func (tb *TelegramBot) SendCompleteNotification(processed *models.ProcessedWebhook, isPremium bool) error {
	// Генерируем краткое резюме с помощью AI
	summary, err := tb.ai.GenerateSummary(processed.Content, processed.TopicTitle, processed.AuthorRole, processed.Category)
	if err != nil {
		log.Printf("Failed to generate AI summary: %v", err)
		summary = "Не удалось сгенерировать резюме"
	}

	// Определяем префикс для роли автора
	var rolePrefix string
	switch processed.AuthorRole {
	case "admin":
		rolePrefix = "👑 "
	case "moderator":
		rolePrefix = "🛡️ "
	case "staff":
		rolePrefix = "⭐ "
	case "leader":
		rolePrefix = "🔥 "
	default:
		rolePrefix = ""
	}

	// Формируем сообщение по новому формату
	message := fmt.Sprintf("👤 %s<b>%s</b> создал новый пост: <b>%s</b>\n\n"+
		"📋 %s\n\n"+
		"🔗 <a href=\"%s\">Ссылка на тему</a>\n\n"+
		"🏷 Теги: %s",
		rolePrefix,
		processed.Author,
		processed.TopicTitle,
		summary,
		processed.URL,
		formatTags(processed.Tags))

	// Добавляем информацию о платности, если нужно
	if isPremium {
		message += "\n\n💎 <b>Данный раздел доступен только по подписке.</b>\n" +
			"Оформить VIP можно в тг-боте: https://t.me/gig_combot"
			
		// Если есть ссылка на анонс, добавляем её
		if processed.AnnouncementURL != "" {
			message += "\n\n💬 <a href=\"" + processed.AnnouncementURL + "\">Обсудить в анонсе</a> - задавайте вопросы и делитесь мнением!"
		}
	}

	// Определяем thread ID на основе категории
	threadID := tb.config.GetThreadIDForCategory(processed.CategoryID)

	return tb.sendMessage(message, threadID)
}

func (tb *TelegramBot) sendMessage(text string, threadID int) error {
	msg := tgbotapi.NewMessage(tb.config.TelegramChatID, text)
	msg.ParseMode = "HTML"

	// Используем переданный thread ID, если он не равен 0
	if threadID != 0 {
		msg.ReplyToMessageID = threadID
	}

	_, err := tb.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %v", err)
	}

	return nil
}

func formatTags(tags []string) string {
	if len(tags) == 0 {
		return "нет"
	}

	result := ""
	for i, tag := range tags {
		if i > 0 {
			result += ", "
		}
		result += "#" + tag
	}
	return result
}
