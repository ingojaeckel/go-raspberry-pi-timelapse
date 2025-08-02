package detection

import (
	"fmt"
	"image"
	"log"
	"os"
	"strings"
)

// DetectionResult represents the results of object detection analysis
type DetectionResult struct {
	IsDay     bool     `json:"isDay"`
	Objects   []string `json:"objects"`
	Summary   string   `json:"summary"`
	PhotoPath string   `json:"photoPath"`
}

// AnalyzePhoto performs basic object detection on the given photo
func AnalyzePhoto(photoPath string) (*DetectionResult, error) {
	if photoPath == "" {
		return nil, fmt.Errorf("photo path cannot be empty")
	}

	// Open and decode the image
	file, err := os.Open(photoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open photo: %v", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	if format != "jpeg" && format != "png" && format != "gif" {
		log.Printf("Warning: image format %s may not be fully supported, continuing with basic analysis", format)
	}

	// Perform basic analysis
	result := &DetectionResult{
		PhotoPath: photoPath,
		Objects:   []string{},
	}

	// Analyze brightness for day/night detection
	result.IsDay = analyzeTimeOfDay(img)

	// Basic object detection based on image characteristics
	objects := analyzeBasicObjects(img)
	result.Objects = objects

	// Generate summary
	result.Summary = generateSummary(result.IsDay, objects)

	return result, nil
}

// analyzeTimeOfDay determines if it's day or night based on image brightness
func analyzeTimeOfDay(img image.Image) bool {
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y
	
	if width == 0 || height == 0 {
		return true // default to day if we can't analyze
	}

	// Sample pixels to calculate average brightness
	sampleSize := 100 // Sample every nth pixel to avoid performance issues
	totalBrightness := uint64(0)
	sampleCount := 0

	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleSize {
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleSize {
			r, g, b, _ := img.At(x, y).RGBA()
			// Convert to 8-bit and calculate brightness using standard formula
			brightness := (0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8))
			totalBrightness += uint64(brightness)
			sampleCount++
		}
	}

	if sampleCount == 0 {
		return true // default to day
	}

	avgBrightness := float64(totalBrightness) / float64(sampleCount)
	
	// Threshold for day/night (adjustable based on camera and environment)
	dayThreshold := 80.0
	
	return avgBrightness > dayThreshold
}

// analyzeBasicObjects performs simple object detection based on image characteristics
func analyzeBasicObjects(img image.Image) []string {
	objects := []string{}
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	if width == 0 || height == 0 {
		return objects
	}

	// Sample colors to detect different types of objects
	colorCounts := make(map[string]int)
	sampleSize := 50 // Sample every nth pixel
	totalSamples := 0

	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleSize {
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleSize {
			r, g, b, _ := img.At(x, y).RGBA()
			// Convert to 8-bit
			red := r >> 8
			green := g >> 8
			blue := b >> 8

			// Categorize colors
			if isGreenish(red, green, blue) {
				colorCounts["vegetation"]++
			} else if isBluish(red, green, blue) {
				colorCounts["sky_water"]++
			} else if isBrownish(red, green, blue) {
				colorCounts["earth_wood"]++
			} else if isGrayish(red, green, blue) {
				colorCounts["structures"]++
			}
			totalSamples++
		}
	}

	// Determine objects based on color distribution
	if totalSamples > 0 {
		vegetationPercent := float64(colorCounts["vegetation"]) / float64(totalSamples) * 100
		skyWaterPercent := float64(colorCounts["sky_water"]) / float64(totalSamples) * 100
		structuresPercent := float64(colorCounts["structures"]) / float64(totalSamples) * 100

		if vegetationPercent > 15 {
			objects = append(objects, "vegetation")
		}
		if skyWaterPercent > 10 {
			objects = append(objects, "sky")
		}
		if structuresPercent > 20 {
			objects = append(objects, "structures")
		}
		
		// If we don't detect much of anything specific, it might be a general scene
		if len(objects) == 0 {
			objects = append(objects, "general scene")
		}
	}

	return objects
}

// Color detection helper functions
func isGreenish(r, g, b uint32) bool {
	return g > r && g > b && g > 60
}

func isBluish(r, g, b uint32) bool {
	return b > r && b > g && b > 60
}

func isBrownish(r, g, b uint32) bool {
	return r > 80 && g > 40 && g < r && b < r && b < g
}

func isGrayish(r, g, b uint32) bool {
	diff := func(a, b uint32) uint32 {
		if a > b {
			return a - b
		}
		return b - a
	}
	return diff(r, g) < 30 && diff(g, b) < 30 && diff(r, b) < 30 && r > 40
}

// generateSummary creates a human-readable summary of the detection results
func generateSummary(isDay bool, objects []string) string {
	timeOfDay := "night"
	if isDay {
		timeOfDay = "day"
	}

	if len(objects) == 0 {
		return fmt.Sprintf("It's %s time. No specific objects detected.", timeOfDay)
	}

	objectsText := strings.Join(objects, ", ")
	return fmt.Sprintf("It's %s time. The photo includes: %s", timeOfDay, objectsText)
}