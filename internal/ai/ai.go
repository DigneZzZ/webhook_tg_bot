package ai

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	"webhook_tg_bot/internal/config"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/openai/openai-go/v2/shared"
)

// AIProvider интерфейс для работы с AI
type AIProvider interface {
	GenerateSummary(content, title, authorRole, category string) (string, error)
}

// OpenAIProvider реализация для OpenAI с официальным SDK
type OpenAIProvider struct {
	client openai.Client
	model  shared.ChatModel
	debug  bool
}

// NewProvider создает провайдер AI с официальным OpenAI SDK
func NewProvider(cfg *config.Config) (AIProvider, error) {
	if cfg.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	// Создаем клиента с API ключом
	client := openai.NewClient(
		option.WithAPIKey(cfg.OpenAIAPIKey),
	)

	// Определяем модель
	var model shared.ChatModel
	switch cfg.OpenAIModel {
	case "gpt-5-mini":
		model = shared.ChatModelGPT5Mini
	case "gpt-5-nano":
		model = shared.ChatModelGPT5Nano
	case "gpt-5":
		model = shared.ChatModelGPT5
	case "gpt-4o":
		model = shared.ChatModelGPT4o
	case "gpt-4o-mini":
		model = shared.ChatModelGPT4oMini
	default:
		model = shared.ChatModelGPT5Nano // По умолчанию используем стабильную модель для тестирования
	}

	return &OpenAIProvider{
		client: client,
		model:  model,
		debug:  cfg.Debug,
	}, nil
}

// isGPT5Model проверяет, является ли модель GPT-5 серии
func (p *OpenAIProvider) isGPT5Model() bool {
	return p.model == shared.ChatModelGPT5 ||
		p.model == shared.ChatModelGPT5Mini ||
		p.model == shared.ChatModelGPT5Nano
}

// isReasoningModel проверяет, является ли модель reasoning model
func (p *OpenAIProvider) isReasoningModel() bool {
	// Официальный SDK автоматически обрабатывает reasoning модели
	// Но мы можем проверить модель для настройки параметров
	modelStr := string(p.model)
	return strings.HasPrefix(modelStr, "o1") ||
		strings.HasPrefix(modelStr, "o3") ||
		strings.HasPrefix(modelStr, "o4") ||
		p.isGPT5Model()
}

// debugLog выводит логи только если включен debug режим
func (p *OpenAIProvider) debugLog(format string, args ...interface{}) {
	if p.debug {
		log.Printf(format, args...)
	}
}

// GenerateSummary генерирует краткое резюме с использованием GPT-5 параметров
func (p *OpenAIProvider) GenerateSummary(content, title, authorRole, category string) (string, error) {
	p.debugLog("AI: Starting summary generation for content length: %d", len(content))

	// Очищаем содержимое от HTML тегов для лучшего анализа
	cleanContent := strings.ReplaceAll(content, "<p>", "")
	cleanContent = strings.ReplaceAll(cleanContent, "</p>", "")
	cleanContent = strings.ReplaceAll(cleanContent, "<br>", " ")
	cleanContent = strings.TrimSpace(cleanContent)

	p.debugLog("AI: Using model: %s", p.model)
	p.debugLog("AI: Clean content: '%s'", cleanContent)

	// Проверяем, что контент не пустой
	if len(cleanContent) == 0 {
		return "Автор оставил пустое сообщение", nil
	}

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

	p.debugLog("AI: Creating request parameters...")

	// Создаем параметры запроса с GPT-5 поддержкой
	params := openai.ChatCompletionNewParams{
		Model: p.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		MaxCompletionTokens: openai.Int(200), // Уменьшаем лимит токенов для более кратких ответов
	}

	// Настраиваем специфичные для GPT-5 параметры
	if p.isGPT5Model() {
		p.debugLog("AI: Using GPT-5 parameters")
		// Используем минимальные значения для надежности
		params.Verbosity = openai.ChatCompletionNewParamsVerbosityLow
		params.ReasoningEffort = shared.ReasoningEffortMinimal
		p.debugLog("AI: Set verbosity to low and reasoning effort to minimal for reliability")
	} else if p.isReasoningModel() {
		p.debugLog("AI: Using reasoning model parameters")
		// Для других reasoning моделей используем только ReasoningEffort
		params.ReasoningEffort = shared.ReasoningEffortMedium
	} else {
		p.debugLog("AI: Using standard model parameters")
		// Для стандартных моделей используем классические параметры
		params.Temperature = openai.Float(0.2)
	}

	p.debugLog("AI: Making OpenAI API call...")
	// Выполняем запрос
	completion, err := p.client.Chat.Completions.New(ctx, params)
	if err != nil {
		log.Printf("AI: OpenAI API error: %v", err)
		return "", fmt.Errorf("OpenAI API error: %v", err)
	}

	p.debugLog("AI: Received response, processing...")
	if len(completion.Choices) == 0 {
		p.debugLog("AI: No choices in response")
		return "", fmt.Errorf("no response from OpenAI")
	}

	result := strings.TrimSpace(completion.Choices[0].Message.Content)
	p.debugLog("AI: Raw response: '%s'", completion.Choices[0].Message.Content)
	p.debugLog("AI: Generated summary successfully, length: %d, content: '%s'", len(result), result)

	// Если результат пустой, возвращаем дефолтное сообщение
	if len(result) == 0 {
		return "Автор оставил краткое сообщение", nil
	}

	return result, nil
}
