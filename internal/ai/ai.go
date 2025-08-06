package ai

import (
	"context"
	"fmt"
	"strings"
	"time"
	"webhook_tg_bot/internal/config"

	"github.com/sashabaranov/go-openai"
)

// AIProvider интерфейс для работы с AI
type AIProvider interface {
	GenerateSummary(content, title, authorRole, category string) (string, error)
}

// OpenAIProvider реализация для OpenAI
type OpenAIProvider struct {
	client *openai.Client
	model  string
}

// NewProvider создает провайдер AI в зависимости от конфигурации
func NewProvider(cfg *config.Config) (AIProvider, error) {
	if cfg.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}
	return &OpenAIProvider{
		client: openai.NewClient(cfg.OpenAIAPIKey),
		model:  cfg.OpenAIModel,
	}, nil
}

// GenerateSummary генерирует краткое резюме с помощью OpenAI
func (p *OpenAIProvider) GenerateSummary(content, title, authorRole, category string) (string, error) {
	// Очищаем содержимое от HTML тегов для лучшего анализа
	cleanContent := strings.ReplaceAll(content, "<p>", "")
	cleanContent = strings.ReplaceAll(cleanContent, "</p>", "")
	cleanContent = strings.ReplaceAll(cleanContent, "<br>", " ")
	cleanContent = strings.TrimSpace(cleanContent)

	prompt := fmt.Sprintf(`Ты - эксперт по анализу контента технических форумов. Создай краткое описание КОНКРЕТНОГО ПОСТА.

КОНТЕКСТ:
Тема форума: "%s"
Категория: %s
Роль автора: %s

СОДЕРЖАНИЕ КОНКРЕТНОГО ПОСТА:
%s

ПРАВИЛА ОПИСАНИЯ ПОСТА:

1. ОСМЫСЛЕННЫЙ КОНТЕНТ:
   • Опиши что конкретно написал автор в этом посте (НЕ в теме в целом)
   • Укажи суть сообщения в 1-2 предложениях
   • НЕ описывай тему форума, а именно содержание поста
   • Фокусируйся на том, что автор хотел сказать

2. ТЕСТОВЫЙ/БЕССМЫСЛЕННЫЙ КОНТЕНТ:
   • "Автор оставил тестовое сообщение"
   • "Пост содержит бессмысленный набор символов"

3. ВОПРОСЫ В ПОСТЕ:
   • "Автор задает вопрос о..."
   • "Пользователь просит помощи с..."

4. ОТВЕТЫ/РЕШЕНИЯ В ПОСТЕ:
   • "Автор предлагает решение..."
   • "Пользователь объясняет как..."

5. КОММЕНТАРИИ/МНЕНИЯ В ПОСТЕ:
   • "Автор высказывает мнение о..."
   • "Пользователь комментирует..."

ВАЖНО:
- Описывай именно СОДЕРЖАНИЕ ПОСТА, а не тему форума
- Максимум 2 предложения на русском языке
- Не упоминай название темы и роль автора в описании
- Фокусируйся на том, что написано в самом сообщении

Описание поста:`, title, category, authorRole, cleanContent)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := p.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: p.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   100,
			Temperature: 0.2,
		},
	)

	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}
