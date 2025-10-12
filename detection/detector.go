package detection

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Detection represents a single detected object
type Detection struct {
	ClassName  string  `json:"class_name"`
	Confidence float64 `json:"confidence"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Width      float64 `json:"width"`
	Height     float64 `json:"height"`
}

// DetectionResult contains the full detection result
type DetectionResult struct {
	Detections []Detection `json:"detections"`
	ImagePath  string      `json:"image_path"`
	Summary    string      `json:"summary"`
}

// Detector interface for object detection
type Detector interface {
	Detect(imagePath string) (*DetectionResult, error)
	IsAvailable() bool
}

// YOLODetector uses external YOLO command for detection
type YOLODetector struct {
	commandPath string
	modelPath   string
	enabled     bool
}

// NewYOLODetector creates a new YOLO detector
func NewYOLODetector(enabled bool) *YOLODetector {
	// Try to find yolo_detect script in multiple locations
	possiblePaths := []string{
		"/usr/local/bin/yolo_detect",
		"./scripts/yolo_detect.py",
		"/opt/timelapse/yolo_detect.py",
	}
	
	commandPath := possiblePaths[0]
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			commandPath = path
			break
		}
	}
	
	return &YOLODetector{
		commandPath: commandPath,
		modelPath:   "/usr/local/share/yolo/yolov5s.onnx",
		enabled:     enabled,
	}
}

// IsAvailable checks if the YOLO detector is available
func (d *YOLODetector) IsAvailable() bool {
	if !d.enabled {
		return false
	}
	// Check if the command exists
	if _, err := os.Stat(d.commandPath); err == nil {
		return true
	}
	// Also check if it's in PATH
	_, err := exec.LookPath(d.commandPath)
	return err == nil
}

// Detect performs object detection on an image
func (d *YOLODetector) Detect(imagePath string) (*DetectionResult, error) {
	if !d.enabled {
		return &DetectionResult{
			Detections: []Detection{},
			ImagePath:  imagePath,
			Summary:    "Object detection is disabled",
		}, nil
	}

	if !d.IsAvailable() {
		return nil, fmt.Errorf("YOLO detector not available at %s", d.commandPath)
	}

	// Execute the YOLO detection command
	// The command should output JSON with detections
	cmd := exec.Command(d.commandPath, "--image", imagePath, "--model", d.modelPath, "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error running YOLO detection: %v, output: %s", err, string(output))
		return nil, fmt.Errorf("failed to run detection: %w", err)
	}

	// Parse the JSON output
	var result DetectionResult
	if err := json.Unmarshal(output, &result); err != nil {
		log.Printf("Error parsing YOLO output: %v, output: %s", err, string(output))
		return nil, fmt.Errorf("failed to parse detection output: %w", err)
	}

	result.ImagePath = imagePath
	result.Summary = generateSummary(result.Detections)

	return &result, nil
}

// generateSummary creates a human-readable summary of detections
func generateSummary(detections []Detection) string {
	if len(detections) == 0 {
		return "No objects detected"
	}

	// Count objects by class
	classCounts := make(map[string]int)
	for _, det := range detections {
		classCounts[det.ClassName]++
	}

	// Build summary
	var parts []string
	for className, count := range classCounts {
		if count == 1 {
			parts = append(parts, fmt.Sprintf("one %s", className))
		} else {
			parts = append(parts, fmt.Sprintf("%d %ss", count, className))
		}
	}

	summary := "The photo includes: " + strings.Join(parts, ", ")
	
	// Add day/night detection (simple heuristic based on common classes)
	// This is a placeholder - a real implementation would analyze image brightness
	hasPersonOrAnimal := false
	for _, det := range detections {
		if det.ClassName == "person" || strings.Contains(det.ClassName, "bird") || 
		   strings.Contains(det.ClassName, "cat") || strings.Contains(det.ClassName, "dog") {
			hasPersonOrAnimal = true
			break
		}
	}
	
	if hasPersonOrAnimal {
		summary = "It's day time. " + summary
	}

	return summary
}

// MockDetector is a simple detector for testing when YOLO is not available
type MockDetector struct {
	enabled bool
}

// NewMockDetector creates a mock detector
func NewMockDetector(enabled bool) *MockDetector {
	return &MockDetector{enabled: enabled}
}

// IsAvailable always returns true for mock detector
func (d *MockDetector) IsAvailable() bool {
	return d.enabled
}

// Detect returns empty results for mock detector
func (d *MockDetector) Detect(imagePath string) (*DetectionResult, error) {
	if !d.enabled {
		return &DetectionResult{
			Detections: []Detection{},
			ImagePath:  imagePath,
			Summary:    "Object detection is disabled",
		}, nil
	}

	log.Printf("Mock detector: would detect objects in %s (YOLO not available)", imagePath)
	return &DetectionResult{
		Detections: []Detection{},
		ImagePath:  imagePath,
		Summary:    "Mock detection: YOLO model not available on this system",
	}, nil
}
