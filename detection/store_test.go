package detection

import (
	"testing"
)

func TestNewResultStore(t *testing.T) {
	store := NewResultStore()
	if store == nil {
		t.Error("Expected store to be created")
	}
}

func TestStoreAndGet(t *testing.T) {
	store := NewResultStore()
	result := &DetectionResult{
		Detections: []Detection{},
		ImagePath:  "/tmp/test.jpg",
		Summary:    "Test summary",
	}
	
	store.Store("/tmp/test.jpg", result)
	retrieved, exists := store.Get("/tmp/test.jpg")
	
	if !exists {
		t.Error("Expected result to exist")
	}
	if retrieved == nil {
		t.Error("Expected retrieved result to be non-nil")
	}
	if retrieved.Summary != "Test summary" {
		t.Errorf("Expected summary 'Test summary', got: %s", retrieved.Summary)
	}
}

func TestGetNonExistent(t *testing.T) {
	store := NewResultStore()
	_, exists := store.Get("/tmp/nonexistent.jpg")
	
	if exists {
		t.Error("Expected result to not exist")
	}
}

func TestGetLatest(t *testing.T) {
	store := NewResultStore()
	
	result1 := &DetectionResult{
		ImagePath: "/tmp/test1.jpg",
		Summary:   "First",
	}
	result2 := &DetectionResult{
		ImagePath: "/tmp/test2.jpg",
		Summary:   "Second",
	}
	
	store.Store("/tmp/test1.jpg", result1)
	store.Store("/tmp/test2.jpg", result2)
	
	latest := store.GetLatest()
	if latest == nil {
		t.Error("Expected latest result to be non-nil")
	}
	// Latest should be test2 since the path is lexicographically greater
	if latest.ImagePath != "/tmp/test2.jpg" {
		t.Errorf("Expected latest to be test2.jpg, got: %s", latest.ImagePath)
	}
}

func TestGetLatestEmpty(t *testing.T) {
	store := NewResultStore()
	latest := store.GetLatest()
	
	if latest != nil {
		t.Error("Expected latest to be nil for empty store")
	}
}

func TestClear(t *testing.T) {
	store := NewResultStore()
	result := &DetectionResult{
		ImagePath: "/tmp/test.jpg",
		Summary:   "Test",
	}
	
	store.Store("/tmp/test.jpg", result)
	store.Clear()
	
	_, exists := store.Get("/tmp/test.jpg")
	if exists {
		t.Error("Expected result to not exist after clear")
	}
}
