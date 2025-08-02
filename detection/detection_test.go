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

func TestAnalyzeTimeOfDay(t *testing.T) {
	// Create a bright test image (day)
	brightImg := createTestImage(200, 200, color.RGBA{200, 200, 200, 255})
	isDayBright := analyzeTimeOfDay(brightImg)
	if !isDayBright {
		t.Error("Expected bright image to be detected as day")
	}

	// Create a dark test image (night)
	darkImg := createTestImage(200, 200, color.RGBA{20, 20, 20, 255})
	isDayDark := analyzeTimeOfDay(darkImg)
	if isDayDark {
		t.Error("Expected dark image to be detected as night")
	}
}

func TestAnalyzeBasicObjects(t *testing.T) {
	// Create a green test image (vegetation)
	greenImg := createTestImage(200, 200, color.RGBA{50, 150, 50, 255})
	objects := analyzeBasicObjects(greenImg)
	
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