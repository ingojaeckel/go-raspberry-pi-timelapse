package detection

import (
	"image"
	"image/color"
	"testing"
)

func TestAnalyzePhoto_InvalidPath(t *testing.T) {
	_, err := AnalyzePhoto("")
	if err == nil {
		t.Error("Expected error for empty path")
	}

	_, err = AnalyzePhoto("/nonexistent/path.jpg")
	if err == nil {
		t.Error("Expected error for nonexistent path")
	}
}

func TestAnalyzeTimeOfDayEnhanced(t *testing.T) {
	// Create a bright test image (day)
	brightImg := createTestImage(200, 200, color.RGBA{200, 200, 200, 255})
	isDayBright := analyzeTimeOfDayEnhanced(brightImg)
	if !isDayBright {
		t.Error("Expected bright image to be detected as day")
	}

	// Create a dark test image (night)
	darkImg := createTestImage(200, 200, color.RGBA{20, 20, 20, 255})
	isDayDark := analyzeTimeOfDayEnhanced(darkImg)
	if isDayDark {
		t.Error("Expected dark image to be detected as night")
	}
}

func TestDetectObjectsEnhanced(t *testing.T) {
	// Create a green test image (vegetation)
	greenImg := createTestImage(200, 200, color.RGBA{50, 150, 50, 255})
	objects, details := detectObjectsEnhanced(greenImg)
	
	found := false
	for _, obj := range objects {
		if obj == "vegetation" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to detect vegetation in green image")
	}
	
	// Check that details are provided
	if len(details) == 0 {
		t.Error("Expected detection details to be provided")
	}
}

func TestEnhancedColorDetection(t *testing.T) {
	// Test skin color detection
	if !isSkinColor(150, 100, 80) {
		t.Error("Expected skin color to be detected")
	}
	
	// Test vegetation color detection
	if !isVegetationColor(50, 150, 50) {
		t.Error("Expected vegetation color to be detected")
	}
	
	// Test sky color detection
	if !isSkyColor(100, 150, 255) {
		t.Error("Expected sky color to be detected")
	}
	
	// Test metal color detection
	if !isMetalColor(120, 120, 120) {
		t.Error("Expected metal color to be detected")
	}
}

func TestCategorizeObject(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"human", "human"},
		{"person", "human"},
		{"animal", "animal"},
		{"cat", "animal"},
		{"vehicle", "vehicle"},
		{"car", "vehicle"},
		{"machinery", "machinery"},
		{"truck", "machinery"},
		{"unknown", "unknown"},
	}
	
	for _, test := range tests {
		result := categorizeObject(test.input)
		if result != test.expected {
			t.Errorf("categorizeObject(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestGenerateSummary(t *testing.T) {
	// Test day with objects
	summary := generateSummary(true, []string{"vegetation", "sky"})
	expected := "It's day time. The photo includes: vegetation, sky"
	if summary != expected {
		t.Errorf("Expected '%s', got '%s'", expected, summary)
	}

	// Test night with no objects
	summary = generateSummary(false, []string{})
	expected = "It's night time. No specific objects detected."
	if summary != expected {
		t.Errorf("Expected '%s', got '%s'", expected, summary)
	}
}

func TestBoundingBoxGeneration(t *testing.T) {
	// Test generateFakeBBox function
	bounds := image.Rect(0, 0, 600, 400)
	bbox := generateFakeBBox(bounds)
	
	if bbox == nil {
		t.Error("Expected bounding box to be generated")
	}
	
	// Check that the bounding box is within the image bounds
	if bbox.X < 0 || bbox.Y < 0 {
		t.Error("Bounding box coordinates should be positive")
	}
	
	if bbox.X+bbox.Width > 600 || bbox.Y+bbox.Height > 400 {
		t.Error("Bounding box should be within image bounds")
	}
	
	// Check that the box has reasonable dimensions (should be about 1/3 of image)
	expectedWidth := 600 / 3
	expectedHeight := 400 / 3
	
	if bbox.Width != expectedWidth || bbox.Height != expectedHeight {
		t.Errorf("Expected dimensions %dx%d, got %dx%d", expectedWidth, expectedHeight, bbox.Width, bbox.Height)
	}
}

func TestDetectionWithBoundingBoxes(t *testing.T) {
	// Create a green test image (should detect vegetation)
	greenImg := createTestImage(300, 200, color.RGBA{50, 150, 50, 255})
	objects, details := detectObjectsEnhanced(greenImg)
	
	// Check that details include bounding boxes
	for _, detail := range details {
		if detail.BBox == nil {
			t.Error("Expected all detection details to include bounding boxes")
		}
		
		// Check that bounding box coordinates are valid
		if detail.BBox.X < 0 || detail.BBox.Y < 0 {
			t.Error("Bounding box coordinates should be non-negative")
		}
		
		if detail.BBox.Width <= 0 || detail.BBox.Height <= 0 {
			t.Error("Bounding box dimensions should be positive")
		}
	}
	
	if len(objects) == 0 {
		t.Error("Expected to detect some objects in green image")
	}
}

// Helper function to create test images
func createTestImage(width, height int, c color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}