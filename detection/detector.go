package detection

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gocv.io/x/gocv"
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

// YOLODetector uses GoCV for native YOLO detection
type YOLODetector struct {
	net              *gocv.Net
	modelPath        string
	classNames       []string
	enabled          bool
	confidenceThresh float32
	nmsThreshold     float32
	inputWidth       int
	inputHeight      int
}

// COCO class names (80 classes)
var cocoClassNames = []string{
	"person", "bicycle", "car", "motorcycle", "airplane", "bus", "train", "truck", "boat",
	"traffic light", "fire hydrant", "stop sign", "parking meter", "bench", "bird", "cat",
	"dog", "horse", "sheep", "cow", "elephant", "bear", "zebra", "giraffe", "backpack",
	"umbrella", "handbag", "tie", "suitcase", "frisbee", "skis", "snowboard", "sports ball",
	"kite", "baseball bat", "baseball glove", "skateboard", "surfboard", "tennis racket",
	"bottle", "wine glass", "cup", "fork", "knife", "spoon", "bowl", "banana", "apple",
	"sandwich", "orange", "broccoli", "carrot", "hot dog", "pizza", "donut", "cake", "chair",
	"couch", "potted plant", "bed", "dining table", "toilet", "tv", "laptop", "mouse",
	"remote", "keyboard", "cell phone", "microwave", "oven", "toaster", "sink", "refrigerator",
	"book", "clock", "vase", "scissors", "teddy bear", "hair drier", "toothbrush",
}

// NewYOLODetector creates a new YOLO detector using GoCV
func NewYOLODetector(enabled bool) *YOLODetector {
	detector := &YOLODetector{
		modelPath:        "/usr/local/share/yolo/yolov5s.onnx",
		classNames:       cocoClassNames,
		enabled:          enabled,
		confidenceThresh: 0.5,
		nmsThreshold:     0.4,
		inputWidth:       640,
		inputHeight:      640,
	}
	
	// Try to load the model if enabled
	if enabled {
		if err := detector.loadModel(); err != nil {
			log.Printf("Warning: Failed to load YOLO model: %v", err)
		}
	}
	
	return detector
}

// loadModel loads the YOLO ONNX model
func (d *YOLODetector) loadModel() error {
	// Check if model file exists
	if _, err := os.Stat(d.modelPath); err != nil {
		return fmt.Errorf("model file not found: %s", d.modelPath)
	}
	
	// Load the network from ONNX file
	net := gocv.ReadNet(d.modelPath, "")
	if net.Empty() {
		return fmt.Errorf("failed to load model from %s", d.modelPath)
	}
	
	// Set backend and target (CPU by default, can be configured for GPU)
	net.SetPreferableBackend(gocv.NetBackendDefault)
	net.SetPreferableTarget(gocv.NetTargetCPU)
	
	d.net = &net
	log.Printf("YOLO model loaded successfully from %s", d.modelPath)
	return nil
}

// IsAvailable checks if the YOLO detector is available
func (d *YOLODetector) IsAvailable() bool {
	if !d.enabled {
		return false
	}
	return d.net != nil && !d.net.Empty()
}

// Close releases resources used by the detector
func (d *YOLODetector) Close() error {
	if d.net != nil && !d.net.Empty() {
		if err := d.net.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Detect performs object detection on an image using GoCV
func (d *YOLODetector) Detect(imagePath string) (*DetectionResult, error) {
	if !d.enabled {
		return &DetectionResult{
			Detections: []Detection{},
			ImagePath:  imagePath,
			Summary:    "Object detection is disabled",
		}, nil
	}

	if !d.IsAvailable() {
		return nil, fmt.Errorf("YOLO detector not available (model not loaded)")
	}

	// Read the image
	img := gocv.IMRead(imagePath, gocv.IMReadColor)
	if img.Empty() {
		return nil, fmt.Errorf("failed to read image: %s", imagePath)
	}
	defer img.Close()

	// Create blob from image
	blob := gocv.BlobFromImage(img, 1.0/255.0, 
		gocv.NewSize(d.inputWidth, d.inputHeight),
		gocv.NewScalar(0, 0, 0, 0), true, false)
	defer blob.Close()

	// Set input to the network
	d.net.SetInput(blob, "")

	// Forward pass
	probs := d.net.Forward("")
	defer probs.Close()

	// Process detections
	detections := d.postProcess(img, &probs)

	result := &DetectionResult{
		Detections: detections,
		ImagePath:  imagePath,
		Summary:    generateSummary(detections),
	}

	return result, nil
}

// postProcess processes the network output and returns detections
func (d *YOLODetector) postProcess(img gocv.Mat, output *gocv.Mat) []Detection {
	var detections []Detection
	
	// Get image dimensions
	imgHeight := float32(img.Rows())
	imgWidth := float32(img.Cols())
	
	// YOLOv5 output format: [1, 25200, 85] for 640x640 input
	// 85 = x, y, w, h, confidence, 80 class scores
	data := output.DataPtrFloat32()
	rows := output.Size()[1] // 25200 for YOLOv5s
	cols := output.Size()[2] // 85
	
	type detection struct {
		classID    int
		confidence float32
		box        [4]float32
	}
	
	var candidates []detection
	
	// Iterate through detections
	for i := 0; i < rows; i++ {
		offset := i * cols
		
		// Get confidence (5th element)
		objectness := data[offset+4]
		
		if objectness < d.confidenceThresh {
			continue
		}
		
		// Find class with highest score
		var maxClassScore float32
		var maxClassID int
		for j := 5; j < cols; j++ {
			classScore := data[offset+j]
			if classScore > maxClassScore {
				maxClassScore = classScore
				maxClassID = j - 5
			}
		}
		
		// Calculate final confidence
		confidence := objectness * maxClassScore
		
		if confidence < d.confidenceThresh {
			continue
		}
		
		// Get bounding box (first 4 elements are center_x, center_y, width, height)
		centerX := data[offset+0]
		centerY := data[offset+1]
		width := data[offset+2]
		height := data[offset+3]
		
		candidates = append(candidates, detection{
			classID:    maxClassID,
			confidence: confidence,
			box:        [4]float32{centerX, centerY, width, height},
		})
	}
	
	// Apply Non-Maximum Suppression (NMS)
	// For simplicity, we'll keep all detections above threshold
	// A full NMS implementation would remove overlapping boxes
	
	// Scale coordinates back to original image size
	xRatio := imgWidth / float32(d.inputWidth)
	yRatio := imgHeight / float32(d.inputHeight)
	
	for _, det := range candidates {
		if det.classID >= len(d.classNames) {
			continue
		}
		
		// Convert from center coordinates to corner coordinates
		x := (det.box[0] - det.box[2]/2) * xRatio
		y := (det.box[1] - det.box[3]/2) * yRatio
		w := det.box[2] * xRatio
		h := det.box[3] * yRatio
		
		detections = append(detections, Detection{
			ClassName:  d.classNames[det.classID],
			Confidence: float64(det.confidence),
			X:          float64(x),
			Y:          float64(y),
			Width:      float64(w),
			Height:     float64(h),
		})
	}
	
	return detections
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
