package detection

import (
	"testing"
)

func TestNewYOLODetector(t *testing.T) {
	detector := NewYOLODetector(true)
	if detector == nil {
		t.Error("Expected detector to be created")
	}
	if detector.enabled != true {
		t.Error("Expected detector to be enabled")
	}
}

func TestNewMockDetector(t *testing.T) {
	detector := NewMockDetector(true)
	if detector == nil {
		t.Error("Expected mock detector to be created")
	}
	if !detector.IsAvailable() {
		t.Error("Expected mock detector to be available")
	}
}

func TestMockDetectorDisabled(t *testing.T) {
	detector := NewMockDetector(false)
	if detector.IsAvailable() {
		t.Error("Expected disabled mock detector to not be available")
	}
}

func TestMockDetectorDetect(t *testing.T) {
	detector := NewMockDetector(true)
	result, err := detector.Detect("/tmp/test.jpg")
	
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Error("Expected result to be non-nil")
	}
	if result.ImagePath != "/tmp/test.jpg" {
		t.Errorf("Expected image path to be /tmp/test.jpg, got: %s", result.ImagePath)
	}
}

func TestGenerateSummaryEmpty(t *testing.T) {
	detections := []Detection{}
	summary := generateSummary(detections)
	if summary != "No objects detected" {
		t.Errorf("Expected 'No objects detected', got: %s", summary)
	}
}

func TestGenerateSummarySingle(t *testing.T) {
	detections := []Detection{
		{ClassName: "person", Confidence: 0.9, X: 100, Y: 100, Width: 50, Height: 100},
	}
	summary := generateSummary(detections)
	if summary != "It's day time. The photo includes: one person" {
		t.Errorf("Expected day time summary with one person, got: %s", summary)
	}
}

func TestGenerateSummaryMultiple(t *testing.T) {
	detections := []Detection{
		{ClassName: "bird", Confidence: 0.8, X: 100, Y: 100, Width: 30, Height: 30},
		{ClassName: "bird", Confidence: 0.85, X: 200, Y: 150, Width: 30, Height: 30},
	}
	summary := generateSummary(detections)
	// Check that it mentions birds (the exact format may vary due to map iteration)
	if !contains(summary, "bird") {
		t.Errorf("Expected summary to mention birds, got: %s", summary)
	}
}

func TestYOLODetectorDisabled(t *testing.T) {
	detector := NewYOLODetector(false)
	result, err := detector.Detect("/tmp/test.jpg")
	
	if err != nil {
		t.Errorf("Expected no error for disabled detector, got: %v", err)
	}
	if result == nil {
		t.Error("Expected result to be non-nil")
	}
	if result.Summary != "Object detection is disabled" {
		t.Errorf("Expected disabled message, got: %s", result.Summary)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
