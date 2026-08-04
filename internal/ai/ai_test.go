package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"webhook_tg_bot/internal/config"
)

func newTestProvider(t *testing.T, baseURL string) AIProvider {
	t.Helper()
	provider, err := NewProvider(&config.Config{
		OpenAIAPIKey:  "test-key",
		OpenAIModel:   "gpt-5-nano",
		OpenAIBaseURL: baseURL,
	})
	if err != nil {
		t.Fatalf("NewProvider: %v", err)
	}
	return provider
}

func completionJSON(text string) []byte {
	resp := map[string]any{
		"id":      "chatcmpl-test",
		"object":  "chat.completion",
		"created": 0,
		"model":   "gpt-5-nano",
		"choices": []map[string]any{
			{
				"index":         0,
				"finish_reason": "stop",
				"message":       map[string]any{"role": "assistant", "content": text},
			},
		},
	}
	b, _ := json.Marshal(resp)
	return b
}

// Временная ошибка (429 rate_limit) должна ретраиться и в итоге завершиться успехом
func TestGenerateSummaryRetriesOnRateLimit(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n < 3 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":{"message":"Rate limit reached","type":"rate_limit_exceeded","param":null,"code":"rate_limit_exceeded"}}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(completionJSON("Автор делится скриптом бэкапа."))
	}))
	defer srv.Close()

	provider := newTestProvider(t, srv.URL)
	summary, err := provider.GenerateSummary("Тестовый пост про бэкапы Docker.", "Тема", "user", "Docker")
	if err != nil {
		t.Fatalf("expected success after retries, got error: %v", err)
	}
	if summary != "Автор делится скриптом бэкапа." {
		t.Fatalf("unexpected summary: %q", summary)
	}
	if got := calls.Load(); got != 3 {
		t.Fatalf("expected 3 attempts, got %d", got)
	}
}

// Постоянная ошибка (insufficient_quota) не должна ретраиться
func TestGenerateSummaryNoRetryOnInsufficientQuota(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"message":"You have no credits remaining.","type":"insufficient_quota","param":null,"code":"credit_balance_exhausted"}}`))
	}))
	defer srv.Close()

	provider := newTestProvider(t, srv.URL)
	_, err := provider.GenerateSummary("Тестовый пост.", "Тема", "user", "Docker")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("expected exactly 1 attempt (no retries), got %d", got)
	}
}
