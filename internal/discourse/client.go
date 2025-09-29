package discourse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"webhook_tg_bot/internal/config"
)

// Client клиент для работы с Discourse API
type Client struct {
	baseURL    string
	apiKey     string
	apiUsername string
	httpClient *http.Client
}

// NewClient создает новый клиент Discourse API
func NewClient(cfg *config.Config) (*Client, error) {
	if cfg.DiscourseAPIKey == "" {
		return nil, fmt.Errorf("DISCOURSE_API_KEY is required")
	}
	if cfg.DiscourseAPIUsername == "" {
		return nil, fmt.Errorf("DISCOURSE_API_USERNAME is required")
	}
	if cfg.DiscourseBaseURL == "" {
		return nil, fmt.Errorf("DISCOURSE_BASE_URL is required")
	}

	return &Client{
		baseURL:     cfg.DiscourseBaseURL,
		apiKey:      cfg.DiscourseAPIKey,
		apiUsername: cfg.DiscourseAPIUsername,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// CreateTopicRequest структура для создания новой темы
type CreateTopicRequest struct {
	Title      string `json:"title"`
	Raw        string `json:"raw"`
	CategoryID int    `json:"category,omitempty"`
	Tags       string `json:"tags,omitempty"`
}

// CreateTopicResponse ответ при создании темы
type CreateTopicResponse struct {
	TopicSlug string `json:"topic_slug"`
	TopicID   int    `json:"topic_id"`
	PostID    int    `json:"id"`
}

// Category структура категории Discourse
type Category struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	ParentID    *int   `json:"parent_category_id,omitempty"`
}

// CategoriesResponse ответ со списком категорий
type CategoriesResponse struct {
	CategoryList struct {
		Categories []Category `json:"categories"`
	} `json:"category_list"`
}

// CreateTopic создает новую тему в указанной категории
func (c *Client) CreateTopic(categoryID int, title, content string, tags []string) (*CreateTopicResponse, error) {
	url := fmt.Sprintf("%s/posts.json", c.baseURL)
	
	// Формируем теги
	tagsStr := ""
	if len(tags) > 0 {
		tagsBytes, _ := json.Marshal(tags)
		tagsStr = string(tagsBytes)
	}

	payload := CreateTopicRequest{
		Title:      title,
		Raw:        content,
		CategoryID: categoryID,
		Tags:       tagsStr,
	}

	return c.makePostRequest(url, payload)
}

// GetCategories получает список всех категорий
func (c *Client) GetCategories() ([]Category, error) {
	url := fmt.Sprintf("%s/categories.json", c.baseURL)
	
	resp, err := c.makeGetRequest(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var categoriesResp CategoriesResponse
	if err := json.Unmarshal(body, &categoriesResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal categories response: %v", err)
	}

	return categoriesResp.CategoryList.Categories, nil
}

// makePostRequest выполняет POST запрос к Discourse API
func (c *Client) makePostRequest(url string, payload interface{}) (*CreateTopicResponse, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	c.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result CreateTopicResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return &result, nil
}

// makeGetRequest выполняет GET запрос к Discourse API
func (c *Client) makeGetRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	return resp, nil
}

// setHeaders устанавливает необходимые заголовки для аутентификации
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Api-Key", c.apiKey)
	req.Header.Set("Api-Username", c.apiUsername)
	req.Header.Set("User-Agent", "webhook_tg_bot/1.0")
}

// TestConnection проверяет подключение к Discourse API
func (c *Client) TestConnection() error {
	_, err := c.GetCategories()
	if err != nil {
		return fmt.Errorf("failed to connect to Discourse API: %v", err)
	}
	return nil
}