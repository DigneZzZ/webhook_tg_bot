package storage

import (
	"sync"
	"time"
	"webhook_tg_bot/internal/models"
)

// TopicData объединенные данные о теме
type TopicData struct {
	Topic     *models.Topic
	Post      *models.Post
	CreatedAt time.Time
	Complete  bool // есть ли и топик и пост
}

// MemoryStorage простое хранилище в памяти
type MemoryStorage struct {
	topics map[int]*TopicData
	mutex  sync.RWMutex
	ttl    time.Duration
}

// NewMemoryStorage создает новое хранилище
func NewMemoryStorage() *MemoryStorage {
	storage := &MemoryStorage{
		topics: make(map[int]*TopicData),
		ttl:    5 * time.Minute, // TTL для автоочистки
	}

	// Запускаем горутину для очистки устаревших записей
	go storage.cleanup()

	return storage
}

// AddTopic добавляет данные о теме
func (s *MemoryStorage) AddTopic(topic *models.Topic) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	data, exists := s.topics[topic.ID]
	if !exists {
		data = &TopicData{
			CreatedAt: time.Now(),
		}
		s.topics[topic.ID] = data
	}

	data.Topic = topic
	data.Complete = data.Topic != nil && data.Post != nil
}

// AddPost добавляет данные о посте
func (s *MemoryStorage) AddPost(post *models.Post) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	data, exists := s.topics[post.TopicID]
	if !exists {
		data = &TopicData{
			CreatedAt: time.Now(),
		}
		s.topics[post.TopicID] = data
	}

	data.Post = post
	data.Complete = data.Topic != nil && data.Post != nil
}

// GetCompleteData возвращает полные данные о теме, если они есть
func (s *MemoryStorage) GetCompleteData(topicID int) (*TopicData, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	data, exists := s.topics[topicID]
	if !exists || !data.Complete {
		return nil, false
	}

	return data, true
}

// RemoveTopic удаляет данные о теме
func (s *MemoryStorage) RemoveTopic(topicID int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.topics, topicID)
}

// cleanup удаляет устаревшие записи
func (s *MemoryStorage) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.Lock()
		now := time.Now()
		for topicID, data := range s.topics {
			if now.Sub(data.CreatedAt) > s.ttl {
				delete(s.topics, topicID)
			}
		}
		s.mutex.Unlock()
	}
}
