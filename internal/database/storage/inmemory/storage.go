package inmemory

import (
	"context"

	"github.com/strider2038/key-value-database/internal/database/storage"
)

type MapStorage struct {
	values map[string]string
}

func NewMapStorage() *MapStorage {
	return &MapStorage{values: make(map[string]string)}
}

func (s *MapStorage) Get(ctx context.Context, key string) (string, error) {
	if value, exists := s.values[key]; exists {
		return value, nil
	}

	return "", storage.ErrNotFound
}

func (s *MapStorage) Set(ctx context.Context, key, value string) error {
	s.values[key] = value

	return nil
}

func (s *MapStorage) Del(ctx context.Context, key string) error {
	delete(s.values, key)

	return nil
}
