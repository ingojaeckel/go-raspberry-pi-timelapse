//go:build opencv
// +build opencv

package detection

import (
	"fmt"
	"image"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gocv.io/x/gocv"
)

// OpenCVDetector provides native Go OpenCV object detection capabilities
type OpenCVDetector struct {
	net         *gocv.Net
	classes     []string
	initialized bool
}

// DetectionBox represents a detected object with its bounding box and confidence
type DetectionBox struct {
	ClassID    int
	Class      string
	Confidence float32
	X          int
	Y          int
	Width      int
	Height     int
}

// NewOpenCVDetector creates a new OpenCV detector instance
func NewOpenCVDetector() *OpenCVDetector {
	detector := &OpenCVDetector{}
	detector.initialize()
	return detector
}

// initialize sets up the YOLO detector with pre-trained weights
func (d *OpenCVDetector) initialize() {
	// Try to load YOLO model files in order of preference
	modelPaths := []struct {
		weights string
		config  string
		name    string
	}{
		{"/opt/yolo/yolov3-tiny.weights", "/opt/yolo/yolov3-tiny.cfg", "YOLOv3-tiny"},
		{"/opt/yolo/yolov3.weights", "/opt/yolo/yolov3.cfg", "YOLOv3"},
		{"/opt/yolo/yolov4-tiny.weights", "/opt/yolo/yolov4-tiny.cfg", "YOLOv4-tiny"},
	}

	classesPath := "/opt/yolo/coco.names"

	// Load class names
	if err := d.loadClassNames(classesPath); err != nil {
		log.Printf("Failed to load class names from %s: %v", classesPath, err)
		// Use default COCO class names as fallback
		d.classes = cocoClassNames
	}

	// Try each model configuration
	for _, model := range modelPaths {
		if d.tryLoadModel(model.weights, model.config, model.name) {
			d.initialized = true
			return
		}
	}

	log.Printf("Warning: No YOLO models found. OpenCV detection will not be available.")
	d.initialized = false
}

// tryLoadModel attempts to load a specific YOLO model
func (d *OpenCVDetector) tryLoadModel(weightsPath, configPath, modelName string) bool {
	// Check if both files exist
	if _, err := os.Stat(weightsPath); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return false
	}

	// Try to load the network
	net := gocv.ReadNetFromDarknet(configPath, weightsPath)
	if net.Empty() {
		log.Printf("Failed to load %s model from %s", modelName, weightsPath)
		return false
	}

	// Set backend and target for optimal performance on Raspberry Pi
	net.SetPreferableBackend(gocv.NetBackendOpenCV)
	net.SetPreferableTarget(gocv.NetTargetCPU)

	d.net = &net
	log.Printf("Successfully loaded %s model for object detection", modelName)
	return true
}

// loadClassNames loads COCO class names from file
func (d *OpenCVDetector) loadClassNames(classesPath string) error {
	data, err := os.ReadFile(classesPath)
	if err != nil {
		return err
	}

	classes := strings.Split(strings.TrimSpace(string(data)), "\n")
	d.classes = classes
	return nil
}

// DetectObjects performs object detection on an image file
func (d *OpenCVDetector) DetectObjects(imagePath string) ([]DetectionBox, error) {
	if !d.initialized || d.net == nil {
		return nil, fmt.Errorf("OpenCV detector not properly initialized")
	}

	// Load the image
	img := gocv.IMRead(imagePath, gocv.IMReadColor)
	if img.Empty() {
		return nil, fmt.Errorf("failed to load image: %s", imagePath)
	}
	defer img.Close()

	return d.detectObjectsFromMat(img)
}

// detectObjectsFromMat performs detection on a gocv.Mat
func (d *OpenCVDetector) detectObjectsFromMat(img gocv.Mat) ([]DetectionBox, error) {
	// Create blob from image
	blob := gocv.BlobFromImage(img, 1.0/255.0, image.Pt(416, 416), gocv.NewScalar(0, 0, 0, 0), true, false, gocv.MatTypeCV32F)
	defer blob.Close()

	// Set input to the network
	d.net.SetInput(blob, "")

	// Get output layer names
	outputNames := d.net.GetUnconnectedOutLayersNames()

	// Run forward pass
	outputs := d.net.ForwardLayers(outputNames)
	defer func() {
		for i := range outputs {
			outputs[i].Close()
		}
	}()

	// Process detections
	detections := d.processDetections(outputs, img.Cols(), img.Rows())

	// Apply Non-Maximum Suppression (NMS)
	filteredDetections := d.applyNMS(detections, 0.5, 0.4)

	return filteredDetections, nil
}

// processDetections extracts detection boxes from network outputs
func (d *OpenCVDetector) processDetections(outputs []gocv.Mat, imgWidth, imgHeight int) []DetectionBox {
	var detections []DetectionBox
	confidenceThreshold := float32(0.5)

	for _, output := range outputs {
		data, err := output.DataPtrFloat32()
		if err != nil {
			continue
		}

		rows := output.Rows()
		cols := output.Cols()

		for i := 0; i < rows; i++ {
			// Each detection is: [center_x, center_y, width, height, confidence, class_probs...]
			confidence := data[i*cols+4]
			
			if confidence > confidenceThreshold {
				// Find the class with highest probability
				maxClassProb := float32(0)
				classID := -1
				for j := 5; j < cols; j++ {
					classProb := data[i*cols+j]
					if classProb > maxClassProb {
						maxClassProb = classProb
						classID = j - 5
					}
				}

				// Final confidence is objectness * class probability
				finalConfidence := confidence * maxClassProb
				
				if finalConfidence > confidenceThreshold && classID >= 0 && classID < len(d.classes) {
					// Convert normalized coordinates to pixel coordinates
					centerX := int(data[i*cols+0] * float32(imgWidth))
					centerY := int(data[i*cols+1] * float32(imgHeight))
					width := int(data[i*cols+2] * float32(imgWidth))
					height := int(data[i*cols+3] * float32(imgHeight))

					// Convert center coordinates to top-left coordinates
					x := centerX - width/2
					y := centerY - height/2

					// Ensure coordinates are within image bounds
					if x < 0 {
						x = 0
					}
					if y < 0 {
						y = 0
					}
					if x+width > imgWidth {
						width = imgWidth - x
					}
					if y+height > imgHeight {
						height = imgHeight - y
					}

					detection := DetectionBox{
						ClassID:    classID,
						Class:      d.classes[classID],
						Confidence: finalConfidence,
						X:          x,
						Y:          y,
						Width:      width,
						Height:     height,
					}
					detections = append(detections, detection)
				}
			}
		}
	}

	return detections
}

// applyNMS applies Non-Maximum Suppression to remove overlapping detections
func (d *OpenCVDetector) applyNMS(detections []DetectionBox, scoreThreshold, nmsThreshold float32) []DetectionBox {
	if len(detections) == 0 {
		return detections
	}

	// Sort detections by confidence (highest first)
	sort.Slice(detections, func(i, j int) bool {
		return detections[i].Confidence > detections[j].Confidence
	})

	var filteredDetections []DetectionBox
	used := make([]bool, len(detections))

	for i := 0; i < len(detections); i++ {
		if used[i] || detections[i].Confidence < scoreThreshold {
			continue
		}

		filteredDetections = append(filteredDetections, detections[i])
		used[i] = true

		// Suppress overlapping detections
		for j := i + 1; j < len(detections); j++ {
			if used[j] {
				continue
			}

			// Calculate IoU (Intersection over Union)
			iou := d.calculateIoU(detections[i], detections[j])
			if iou > nmsThreshold {
				used[j] = true
			}
		}
	}

	return filteredDetections
}

// calculateIoU calculates the Intersection over Union of two bounding boxes
func (d *OpenCVDetector) calculateIoU(box1, box2 DetectionBox) float32 {
	// Calculate intersection coordinates
	x1 := int(math.Max(float64(box1.X), float64(box2.X)))
	y1 := int(math.Max(float64(box1.Y), float64(box2.Y)))
	x2 := int(math.Min(float64(box1.X+box1.Width), float64(box2.X+box2.Width)))
	y2 := int(math.Min(float64(box1.Y+box1.Height), float64(box2.Y+box2.Height)))

	// Check if there's no intersection
	if x2 <= x1 || y2 <= y1 {
		return 0.0
	}

	// Calculate intersection area
	intersectionArea := float32((x2 - x1) * (y2 - y1))

	// Calculate areas of both boxes
	area1 := float32(box1.Width * box1.Height)
	area2 := float32(box2.Width * box2.Height)

	// Calculate union area
	unionArea := area1 + area2 - intersectionArea

	if unionArea == 0 {
		return 0.0
	}

	return intersectionArea / unionArea
}

// IsInitialized returns whether the detector was successfully initialized
func (d *OpenCVDetector) IsInitialized() bool {
	return d.initialized
}

// Close releases resources used by the detector
func (d *OpenCVDetector) Close() {
	if d.net != nil {
		d.net.Close()
		d.net = nil
	}
	d.initialized = false
}

// analyzeWithNativeOpenCV performs object detection using native Go OpenCV bindings
func analyzeWithNativeOpenCV(photoPath string, timeout_unused interface{}) (*DetectionResult, error) {
	detector := NewOpenCVDetector()
	defer detector.Close()

	if !detector.IsInitialized() {
		return nil, fmt.Errorf("failed to initialize OpenCV detector")
	}

	// Perform object detection
	detections, err := detector.DetectObjects(photoPath)
	if err != nil {
		return nil, fmt.Errorf("object detection failed: %v", err)
	}

	// Load image to determine if it's day or night
	file, err := os.Open(photoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open photo for analysis: %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	// Initialize result
	result := &DetectionResult{
		PhotoPath:         photoPath,
		Objects:           []string{},
		Details:           []ObjectDetail{},
		LatencyMs:         0, // Will be set by calling function
		OverallConfidence: 0, // Will be calculated by calling function
	}

	// Analyze time of day
	result.IsDay = analyzeTimeOfDayEnhanced(img)

	// Convert detections to our format
	objectCategories := make(map[string]bool)
	
	for _, detection := range detections {
		category := categorizeObject(detection.Class)
		objectCategories[category] = true

		detail := ObjectDetail{
			Class:      detection.Class,
			Confidence: detection.Confidence,
			Category:   category,
			BBox: &BoundingBox{
				X:      detection.X,
				Y:      detection.Y,
				Width:  detection.Width,
				Height: detection.Height,
			},
		}
		result.Details = append(result.Details, detail)
	}

	// Create objects list from unique categories
	for category := range objectCategories {
		result.Objects = append(result.Objects, category)
	}

	if len(result.Objects) == 0 {
		result.Objects = append(result.Objects, "general scene")
	}

	// Generate summary
	result.Summary = generateSummary(result.IsDay, result.Objects)

	log.Printf("Native OpenCV detection completed: found %d objects in %d categories", 
		len(result.Details), len(result.Objects))

	return result, nil
}