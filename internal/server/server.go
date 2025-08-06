package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"webhook_tg_bot/internal/bot"
	"webhook_tg_bot/internal/config"
	"webhook_tg_bot/internal/models"
	"webhook_tg_bot/internal/storage"

	"github.com/gorilla/mux"
)

type Server struct {
	config  *config.Config
	bot     *bot.TelegramBot
	router  *mux.Router
	storage *storage.MemoryStorage
}

func New(cfg *config.Config, bot *bot.TelegramBot) *Server {
	s := &Server{
		config:  cfg,
		bot:     bot,
		router:  mux.NewRouter(),
		storage: storage.NewMemoryStorage(),
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.HandleFunc(s.config.WebhookPath, s.handleWebhook).Methods("POST")
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")
}

func (s *Server) Start() error {
	return http.ListenAndServe(":"+s.config.WebhookPort, s.router)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Проверяем подпись вебхука
	if !s.verifyWebhookSignature(r, body) {
		log.Printf("Invalid webhook signature")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Определяем тип вебхука и обрабатываем
	if err := s.processWebhook(body); err != nil {
		log.Printf("Error processing webhook: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) verifyWebhookSignature(r *http.Request, body []byte) bool {
	// Получаем подпись из заголовка
	signature := r.Header.Get("X-Discourse-Event-Signature")
	if signature == "" {
		// Fallback: проверяем через query parameter
		secret := r.URL.Query().Get("secret")
		return secret == s.config.WebhookSecret
	}

	// Вычисляем ожидаемую подпись
	mac := hmac.New(sha256.New, []byte(s.config.WebhookSecret))
	mac.Write(body)
	expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func (s *Server) processWebhook(body []byte) error {
	// Пытаемся определить тип вебхука по содержимому
	var topicWebhook models.WebhookTopic
	var postWebhook models.WebhookPost

	// Сначала пробуем парсить как топик
	if err := json.Unmarshal(body, &topicWebhook); err == nil && topicWebhook.Topic.ID != 0 {
		return s.processTopic(&topicWebhook.Topic)
	}

	// Затем пробуем парсить как пост
	if err := json.Unmarshal(body, &postWebhook); err == nil && postWebhook.Post.ID != 0 {
		return s.processPost(&postWebhook.Post)
	}

	log.Printf("Unknown webhook format: %s", string(body))
	return fmt.Errorf("unknown webhook format")
}

func (s *Server) processTopic(topic *models.Topic) error {
	log.Printf("Processing topic: %s (ID: %d, Category: %d, User: %d)", topic.Title, topic.ID, topic.CategoryID, topic.UserID)

	// Проверяем, нужно ли игнорировать этого пользователя
	if s.config.ShouldIgnoreUser(topic.UserID) {
		log.Printf("Skipping topic %d - user %d is ignored", topic.ID, topic.UserID)
		return nil
	}

	// Проверяем, нужно ли отслеживать эту категорию
	if !s.config.ShouldMonitorCategory(topic.CategoryID) {
		log.Printf("Skipping topic %d - category %d is not monitored or is ignored", topic.ID, topic.CategoryID)
		return nil
	}

	// Добавляем топик в хранилище
	s.storage.AddTopic(topic)

	// Проверяем, есть ли полные данные для отправки
	if data, complete := s.storage.GetCompleteData(topic.ID); complete {
		return s.sendCompleteNotification(data)
	}

	return nil
}

func (s *Server) processPost(post *models.Post) error {
	log.Printf("Processing post: %d in topic %d (Category: %d)", post.ID, post.TopicID, post.CategoryID)

	// Обрабатываем только первый пост в теме (создание темы)
	if post.PostNumber != 1 {
		log.Printf("Skipping post %d - not the first post in topic", post.ID)
		return nil
	}

	// Проверяем, нужно ли игнорировать этого пользователя
	if s.config.ShouldIgnoreUser(post.UserID) {
		log.Printf("Skipping post %d - user %d is ignored", post.ID, post.UserID)
		return nil
	}

	// Проверяем, нужно ли отслеживать эту категорию
	if !s.config.ShouldMonitorCategory(post.CategoryID) {
		log.Printf("Skipping post %d - category %d is not monitored or is ignored", post.ID, post.CategoryID)
		return nil
	}

	// Добавляем пост в хранилище
	s.storage.AddPost(post)

	// Проверяем, есть ли полные данные для отправки
	if data, complete := s.storage.GetCompleteData(post.TopicID); complete {
		return s.sendCompleteNotification(data)
	}

	return nil
}

func (s *Server) getCategoryName(data *storage.TopicData) string {
	// Пытаемся получить имя категории из данных поста
	if data.Post != nil && data.Post.CategorySlug != "" {
		// Преобразуем slug в более читаемый вид
		categoryName := strings.ReplaceAll(data.Post.CategorySlug, "-", " ")
		categoryName = strings.ReplaceAll(categoryName, "_", " ")
		// Делаем первую букву заглавной
		if len(categoryName) > 0 {
			categoryName = strings.ToUpper(categoryName[:1]) + categoryName[1:]
		}
		return categoryName
	}

	// Если slug недоступен, показываем ID с более понятным форматом
	if data.Topic != nil {
		return fmt.Sprintf("Раздел %d", data.Topic.CategoryID)
	}

	return "Основной раздел"
} // getUserRole определяет роль пользователя
func (s *Server) getUserRole(user models.User, post *models.Post) string {
	// Приоритет: данные из Post (более полные в webhook'ах)
	if post != nil {
		if post.Admin {
			return "admin"
		}
		if post.Moderator {
			return "moderator"
		}
		if post.Staff {
			return "staff"
		}

		// Определяем по trust level
		if post.TrustLevel >= 4 {
			return "leader"
		}
	}

	// Fallback: проверяем данные из User (могут быть неполными)
	if user.Admin {
		return "admin"
	}
	if user.Moderator {
		return "moderator"
	}
	if user.Staff {
		return "staff"
	}

	// Определяем по trust level из User
	if user.TrustLevel >= 4 {
		return "leader"
	}

	return "user"
}

func (s *Server) sendCompleteNotification(data *storage.TopicData) error {
	// Определяем роль автора
	authorRole := s.getUserRole(data.Topic.CreatedBy, data.Post)
	log.Printf("Author: %s, Role: %s (Admin: %v, Moderator: %v, Staff: %v, TrustLevel: %d)",
		data.Topic.CreatedBy.Username, authorRole,
		data.Post.Admin, data.Post.Moderator, data.Post.Staff, data.Post.TrustLevel)

	// Создаем объединенные данные для отправки
	processed := &models.ProcessedWebhook{
		Type:       "complete",
		TopicID:    data.Topic.ID,
		TopicTitle: data.Topic.Title,
		Category:   s.getCategoryName(data),
		CategoryID: data.Topic.CategoryID,
		Author:     data.Topic.CreatedBy.Username,
		AuthorRole: authorRole,
		Content:    data.Post.Raw,
		Tags:       data.Topic.Tags,
		URL:        fmt.Sprintf("%s/t/%s/%d", s.config.BaseURL, data.Topic.Slug, data.Topic.ID),
	}

	// Отправляем уведомление
	err := s.bot.SendCompleteNotification(processed, s.config.IsPremiumCategory(data.Topic.CategoryID))

	// Удаляем данные из хранилища после отправки
	s.storage.RemoveTopic(data.Topic.ID)

	return err
}
