package service

import (
	"context"
	"sync"
)

type InMemoryCheckPointStore struct {
	mu    sync.RWMutex
	store map[string][]byte
}

func NewInMemoryCheckPointStore() *InMemoryCheckPointStore {
	return &InMemoryCheckPointStore{
		store: make(map[string][]byte),
	}
}

func (s *InMemoryCheckPointStore) Get(_ context.Context, checkPointID string) ([]byte, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, ok := s.store[checkPointID]
	if !ok {
		return nil, false, nil
	}
	return data, true, nil
}

func (s *InMemoryCheckPointStore) Set(_ context.Context, checkPointID string, checkPoint []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[checkPointID] = checkPoint
	return nil
}
