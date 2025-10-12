package detection

import (
	"sync"
)

// ResultStore stores detection results by image path
type ResultStore struct {
	results map[string]*DetectionResult
	mu      sync.RWMutex
}

// NewResultStore creates a new result store
func NewResultStore() *ResultStore {
	return &ResultStore{
		results: make(map[string]*DetectionResult),
	}
}

// Store saves a detection result
func (s *ResultStore) Store(imagePath string, result *DetectionResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results[imagePath] = result
}

// Get retrieves a detection result by image path
func (s *ResultStore) Get(imagePath string) (*DetectionResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result, exists := s.results[imagePath]
	return result, exists
}

// GetLatest returns the most recently stored detection result
func (s *ResultStore) GetLatest() *DetectionResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var latest *DetectionResult
	for _, result := range s.results {
		if latest == nil || result.ImagePath > latest.ImagePath {
			latest = result
		}
	}
	return latest
}

// Clear removes all stored results
func (s *ResultStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results = make(map[string]*DetectionResult)
}
