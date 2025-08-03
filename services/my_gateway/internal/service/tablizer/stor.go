package tablizer

import (
	"sync"
	"time"
)

type TempStorage struct {
	mu   sync.RWMutex
	data map[string]*TableData
}
type TableData struct {
	CSV       []byte
	CreatedAt time.Time
}

func NewTempStorage() *TempStorage {
	storage := &TempStorage{
		data: make(map[string]*TableData),
	}
	go storage.cleanup()
	return storage
}

func (s *TempStorage) Store(requestID string, csv []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[requestID] = &TableData{
		CSV:       csv,
		CreatedAt: time.Now(),
	}
}

func (s *TempStorage) Get(requestID string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, exists := s.data[requestID]
	if !exists {
		return nil, false
	}
	delete(s.data, requestID)
	return data.CSV, true
}

func (s *TempStorage) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		for id, data := range s.data {
			if time.Since(data.CreatedAt) > 5*time.Minute {
				delete(s.data, id)
			}
		}
		s.mu.Unlock()
	}
}
