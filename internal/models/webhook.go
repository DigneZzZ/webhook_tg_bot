package models

import (
	"encoding/json"
	"time"
)

// WebhookTopic представляет данные о созданной теме
type WebhookTopic struct {
	Topic Topic `json:"topic"`
}

// WebhookPost представляет данные о созданном посте
type WebhookPost struct {
	Post Post `json:"post"`
}

// Tag представляет тег в формате объекта
type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// FlexibleTags тип для тегов, который может парсить как массив строк, так и массив объектов
type FlexibleTags []string

// UnmarshalJSON кастомный парсер для тегов
func (ft *FlexibleTags) UnmarshalJSON(data []byte) error {
	// Пробуем сначала как массив строк
	var stringTags []string
	if err := json.Unmarshal(data, &stringTags); err == nil {
		*ft = stringTags
		return nil
	}

	// Если не получилось, пробуем как массив объектов
	var objectTags []Tag
	if err := json.Unmarshal(data, &objectTags); err == nil {
		result := make([]string, len(objectTags))
		for i, tag := range objectTags {
			result[i] = tag.Name
		}
		*ft = result
		return nil
	}

	// Если ничего не получилось, возвращаем пустой массив
	*ft = []string{}
	return nil
}

// Topic структура темы из вебхука
type Topic struct {
	ID         int          `json:"id"`
	Title      string       `json:"title"`
	FancyTitle string       `json:"fancy_title"`
	PostsCount int          `json:"posts_count"`
	CreatedAt  time.Time    `json:"created_at"`
	Views      int          `json:"views"`
	ReplyCount int          `json:"reply_count"`
	LikeCount  int          `json:"like_count"`
	CategoryID int          `json:"category_id"`
	WordCount  int          `json:"word_count"`
	UserID     int          `json:"user_id"`
	Tags       FlexibleTags `json:"tags"`
	Slug       string       `json:"slug"`
	CreatedBy  User         `json:"created_by"`
	LastPoster User         `json:"last_poster"`
}

// Post структура поста из вебхука
type Post struct {
	ID              int       `json:"id"`
	Username        string    `json:"username"`
	CreatedAt       time.Time `json:"created_at"`
	Cooked          string    `json:"cooked"`
	PostNumber      int       `json:"post_number"`
	PostType        int       `json:"post_type"`
	PostsCount      int       `json:"posts_count"`
	TopicID         int       `json:"topic_id"`
	TopicSlug       string    `json:"topic_slug"`
	TopicTitle      string    `json:"topic_title"`
	CategoryID      int       `json:"category_id"`
	CategorySlug    string    `json:"category_slug"`
	Raw             string    `json:"raw"`
	UserID          int       `json:"user_id"`
	TrustLevel      int       `json:"trust_level"`
	TopicPostsCount int       `json:"topic_posts_count"`
	Admin           bool      `json:"admin"`
	Moderator       bool      `json:"moderator"`
	Staff           bool      `json:"staff"`
}

// User структура пользователя
type User struct {
	ID             int    `json:"id"`
	Username       string `json:"username"`
	Name           string `json:"name"`
	AvatarTemplate string `json:"avatar_template"`
	Admin          bool   `json:"admin"`
	Moderator      bool   `json:"moderator"`
	Staff          bool   `json:"staff"`
	TrustLevel     int    `json:"trust_level"`
}

// ProcessedWebhook обработанные данные для отправки в Telegram
type ProcessedWebhook struct {
	Type            string // "topic" или "post"
	TopicID         int
	TopicTitle      string
	Category        string
	CategoryID      int // ID категории для маппинга на thread
	Author          string
	AuthorRole      string // роль автора (admin, moderator, staff, user)
	Content         string
	Tags            []string
	Summary         string
	URL             string
	AnnouncementURL string // URL анонса, если был создан
	ImageURL        string // первая контентная картинка поста для превью в Telegram
}
