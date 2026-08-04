package ai

import (
	"context"
	"errors"
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

// maxAttempts — всего попыток запроса (1 основная + 2 ретрая с бэкоффом)
const maxAttempts = 3

// NewProvider создает провайдер AI с официальным OpenAI SDK
func NewProvider(cfg *config.Config) (AIProvider, error) {
	if cfg.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required (set OPENAI_API_KEY)")
	}

	// Ретраи делаем сами в GenerateSummary, чтобы логировать каждую попытку
	opts := []option.RequestOption{
		option.WithAPIKey(cfg.OpenAIAPIKey),
		option.WithMaxRetries(0),
	}
	// OPENAI_BASE_URL позволяет ходить через прокси/совместимый шлюз,
	// если api.openai.com недоступен напрямую (например, региональный блок)
	if cfg.OpenAIBaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.OpenAIBaseURL))
	}
	client := openai.NewClient(opts...)

	// Имя модели передаётся как есть: так работают и новые модели
	// (gpt-5.1, gpt-5.1-mini и т.д.) без изменения кода
	modelName := cfg.OpenAIModel
	if modelName == "" {
		modelName = "gpt-5-nano"
	}
	log.Printf("AI: provider initialized, model=%s", modelName)

	return &OpenAIProvider{
		client: client,
		model:  shared.ChatModel(modelName),
		debug:  cfg.Debug,
	}, nil
}

// isGPT5Family — вся линейка gpt-5*: gpt-5, gpt-5-mini/nano, gpt-5.1 и новее
func (p *OpenAIProvider) isGPT5Family() bool {
	return strings.HasPrefix(string(p.model), "gpt-5")
}

// isGPT51OrNewer — модели gpt-5.1+ поддерживают reasoning_effort=none вместо minimal
func (p *OpenAIProvider) isGPT51OrNewer() bool {
	return strings.HasPrefix(string(p.model), "gpt-5.")
}

// isReasoningModel проверяет, является ли модель reasoning model
func (p *OpenAIProvider) isReasoningModel() bool {
	modelStr := string(p.model)
	return strings.HasPrefix(modelStr, "o1") ||
		strings.HasPrefix(modelStr, "o3") ||
		strings.HasPrefix(modelStr, "o4") ||
		p.isGPT5Family()
}

// debugLog выводит логи только если включен debug режим
func (p *OpenAIProvider) debugLog(format string, args ...interface{}) {
	if p.debug {
		log.Printf(format, args...)
	}
}

// isRetryableError — временные ошибки, которые имеет смысл повторить:
// таймаут/сеть, 5xx и rate-limit 429. Ошибка исчерпанной квоты
// (insufficient_quota) постоянная — ретраи только тратят время.
func isRetryableError(err error) bool {
	var apierr *openai.Error
	if errors.As(err, &apierr) {
		if apierr.StatusCode == 429 {
			return apierr.Type != "insufficient_quota"
		}
		return apierr.StatusCode >= 500
	}
	// Не-API ошибка: таймаут контекста или сетевой сбой
	return true
}

// logAPIError пишет в лог реальную причину сбоя: HTTP-код, тип и сообщение API
func logAPIError(attempt int, err error) {
	var apierr *openai.Error
	if errors.As(err, &apierr) {
		log.Printf("AI: OpenAI API error (attempt %d/%d): status=%d type=%s code=%v message=%s",
			attempt, maxAttempts, apierr.StatusCode, apierr.Type, apierr.Code, apierr.Message)
		return
	}
	log.Printf("AI: request failed (attempt %d/%d): %v", attempt, maxAttempts, err)
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

	var lastErr error
	backoff := 2 * time.Second
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err := p.complete(prompt)
		if err == nil {
			// Если результат пустой, возвращаем дефолтное сообщение
			if len(result) == 0 {
				return "Автор оставил краткое сообщение", nil
			}
			return result, nil
		}

		lastErr = err
		logAPIError(attempt, err)

		if !isRetryableError(err) {
			break
		}
		if attempt < maxAttempts {
			p.debugLog("AI: retrying in %s...", backoff)
			time.Sleep(backoff)
			backoff *= 2
		}
	}

	return "", fmt.Errorf("OpenAI API error: %w", lastErr)
}

// complete выполняет один запрос к Chat Completions API с таймаутом 30с
func (p *OpenAIProvider) complete(prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Создаем параметры запроса с GPT-5 поддержкой
	params := openai.ChatCompletionNewParams{
		Model: p.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		MaxCompletionTokens: openai.Int(200), // Уменьшаем лимит токенов для более кратких ответов
	}

	// Настраиваем специфичные для GPT-5 параметры
	if p.isGPT5Family() {
		if p.isGPT51OrNewer() {
			// gpt-5.1+ вместо minimal используют none; verbosity в Chat Completions
			// у новых моделей не поддерживается — длину ограничиваем промптом
			// и max_completion_tokens
			params.ReasoningEffort = shared.ReasoningEffort("none")
		} else {
			params.Verbosity = openai.ChatCompletionNewParamsVerbosityLow
			params.ReasoningEffort = shared.ReasoningEffortMinimal
		}
		p.debugLog("AI: Using GPT-5 parameters (reasoning=%s)", params.ReasoningEffort)
	} else if p.isReasoningModel() {
		p.debugLog("AI: Using reasoning model parameters")
		params.ReasoningEffort = shared.ReasoningEffortMedium
	} else {
		p.debugLog("AI: Using standard model parameters")
		params.Temperature = openai.Float(0.2)
	}

	p.debugLog("AI: Making OpenAI API call...")
	completion, err := p.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", err
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	result := strings.TrimSpace(completion.Choices[0].Message.Content)
	p.debugLog("AI: Generated summary successfully, length: %d, content: '%s'", len(result), result)
	return result, nil
}
