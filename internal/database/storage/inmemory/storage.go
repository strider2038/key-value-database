package inmemory

import (
	"sync"

	"github.com/strider2038/key-value-database/internal/database/storage"
)

type MapStorage struct {
	mu     sync.RWMutex
	values map[string]string
}

func NewMapStorage() *MapStorage {
	return &MapStorage{values: make(map[string]string)}
}

func (s *MapStorage) Get(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if value, exists := s.values[key]; exists {
		return value, nil
	}

	return "", storage.ErrNotFound
}

func (s *MapStorage) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.values[key] = value

	return nil
}

func (s *MapStorage) Del(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.values, key)

	return nil
}
