package detection

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// DetectionResult represents the results of object detection analysis
type DetectionResult struct {
	IsDay             bool     `json:"isDay"`
	Objects           []string `json:"objects"`
	Summary           string   `json:"summary"`
	PhotoPath         string   `json:"photoPath"`
	Details           []ObjectDetail `json:"details,omitempty"`
	LatencyMs         int64    `json:"latencyMs"`         // Detection time in milliseconds
	OverallConfidence float32  `json:"overallConfidence"` // Overall confidence score (0.0-1.0)
}

// ObjectDetail provides detailed information about detected objects
type ObjectDetail struct {
	Class      string    `json:"class"`
	Confidence float32   `json:"confidence"`
	Category   string    `json:"category"`
	BBox       *BoundingBox `json:"bbox,omitempty"` // Bounding box coordinates for visual display
}

// BoundingBox represents the coordinates of a detected object
type BoundingBox struct {
	X      int `json:"x"`      // Left coordinate
	Y      int `json:"y"`      // Top coordinate  
	Width  int `json:"width"`  // Width of the box
	Height int `json:"height"` // Height of the box
}

// COCO dataset class names for object detection
var cocoClassNames = []string{
	"person", "bicycle", "car", "motorcycle", "airplane", "bus", "train", "truck",
	"boat", "traffic light", "fire hydrant", "stop sign", "parking meter", "bench",
	"bird", "cat", "dog", "horse", "sheep", "cow", "elephant", "bear", "zebra", "giraffe",
	"backpack", "umbrella", "handbag", "tie", "suitcase", "frisbee", "skis", "snowboard",
	"sports ball", "kite", "baseball bat", "baseball glove", "skateboard", "surfboard",
	"tennis racket", "bottle", "wine glass", "cup", "fork", "knife", "spoon", "bowl",
	"banana", "apple", "sandwich", "orange", "broccoli", "carrot", "hot dog", "pizza",
	"donut", "cake", "chair", "couch", "potted plant", "bed", "dining table", "toilet",
	"tv", "laptop", "mouse", "remote", "keyboard", "cell phone", "microwave", "oven",
	"toaster", "sink", "refrigerator", "book", "clock", "vase", "scissors", "teddy bear",
	"hair drier", "toothbrush",
}

// AnalyzePhoto performs advanced object detection using OpenCV when available,
// falling back to enhanced image analysis
func AnalyzePhoto(photoPath string) (*DetectionResult, error) {
	return AnalyzePhotoWithConfig(photoPath, nil)
}

// AnalyzePhotoWithConfig performs object detection with configuration options
func AnalyzePhotoWithConfig(photoPath string, config *DetectionConfig) (*DetectionResult, error) {
	startTime := time.Now()
	
	if photoPath == "" {
		return nil, fmt.Errorf("photo path cannot be empty")
	}

	// Check if file exists
	if _, err := os.Stat(photoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("photo file does not exist: %s", photoPath)
	}

	// Use default config if none provided
	if config == nil {
		config = &DetectionConfig{
			UseOpenCV: true,
			Timeout:   5 * time.Minute,
		}
	}

	var result *DetectionResult
	var err error
	
	// Try OpenCV-based detection first if enabled
	if config.UseOpenCV {
		if result, err = analyzeWithOpenCV(photoPath, config.Timeout); err == nil {
			log.Printf("Using OpenCV for high-accuracy object detection")
		} else {
			log.Printf("OpenCV detection failed (%v), falling back to enhanced analysis", err)
			// Fallback to original enhanced detection
			result, err = analyzeWithEnhancedDetection(photoPath)
		}
	} else {
		// Use enhanced detection directly
		result, err = analyzeWithEnhancedDetection(photoPath)
	}
	
	if err != nil {
		return nil, err
	}
	
	// Calculate detection latency
	latency := time.Since(startTime)
	result.LatencyMs = latency.Milliseconds()
	
	// Calculate overall confidence from individual detection confidence scores
	result.OverallConfidence = calculateOverallConfidence(result.Details)
	
	// Log detection performance metrics
	log.Printf("Object detection completed in %dms with overall confidence %.2f", 
		result.LatencyMs, result.OverallConfidence)
	
	return result, nil
}

// DetectionConfig holds configuration for object detection
type DetectionConfig struct {
	UseOpenCV bool          // Whether to try OpenCV detection first
	Timeout   time.Duration // Maximum time for detection
}

// analyzeWithOpenCV performs object detection using the OpenCV Python script
func analyzeWithOpenCV(photoPath string, timeout time.Duration) (*DetectionResult, error) {
	// Get the directory where this Go file is located
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("could not determine current file location")
	}
	detectionDir := filepath.Dir(currentFile)
	scriptPath := filepath.Join(detectionDir, "opencv_detector.py")
	
	// Check if the Python script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("OpenCV detection script not found: %s", scriptPath)
	}
	
	// Run the Python script with configurable timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "python3", scriptPath, photoPath, "--output-json")
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("OpenCV detection script failed: %v", err)
	}
	
	// Parse the JSON output
	var result DetectionResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse OpenCV detection output: %v", err)
	}
	
	log.Printf("OpenCV detection completed: found %d object categories", len(result.Objects))
	return &result, nil
}

// analyzeWithEnhancedDetection performs the original enhanced detection as fallback
func analyzeWithEnhancedDetection(photoPath string) (*DetectionResult, error) {
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
		log.Printf("Warning: image format %s may not be fully supported, continuing with enhanced analysis", format)
	}

	// Initialize result
	result := &DetectionResult{
		PhotoPath:         photoPath,
		Objects:           []string{},
		Details:           []ObjectDetail{},
		LatencyMs:         0, // Will be set by calling function
		OverallConfidence: 0, // Will be calculated by calling function
	}

	// Analyze time of day using brightness
	result.IsDay = analyzeTimeOfDayEnhanced(img)

	// Perform enhanced object detection
	objects, details := detectObjectsEnhanced(img)
	result.Objects = objects
	result.Details = details

	// Generate summary
	result.Summary = generateSummary(result.IsDay, objects)

	return result, nil
}

// analyzeTimeOfDayEnhanced determines if it's day or night using enhanced brightness analysis
func analyzeTimeOfDayEnhanced(img image.Image) bool {
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y
	
	if width == 0 || height == 0 {
		return true // default to day if we can't analyze
	}

	// Enhanced sampling with multiple metrics
	sampleSize := 20 // More intensive sampling for better accuracy
	totalBrightness := uint64(0)
	totalContrast := uint64(0)
	sampleCount := 0

	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleSize {
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleSize {
			r, g, b, _ := img.At(x, y).RGBA()
			// Convert to 8-bit and calculate brightness using standard formula
			brightness := (0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8))
			totalBrightness += uint64(brightness)
			
			// Calculate local contrast by comparing with nearby pixels
			if x+sampleSize < bounds.Max.X && y+sampleSize < bounds.Max.Y {
				r2, g2, b2, _ := img.At(x+sampleSize, y+sampleSize).RGBA()
				brightness2 := (0.299*float64(r2>>8) + 0.587*float64(g2>>8) + 0.114*float64(b2>>8))
				contrast := math.Abs(brightness - brightness2)
				totalContrast += uint64(contrast)
			}
			
			sampleCount++
		}
	}

	if sampleCount == 0 {
		return true // default to day
	}

	avgBrightness := float64(totalBrightness) / float64(sampleCount)
	avgContrast := float64(totalContrast) / float64(sampleCount)
	
	// Enhanced day/night detection considering both brightness and contrast
	// Night photos typically have lower brightness but higher contrast (due to artificial lighting)
	dayThreshold := 70.0
	contrastThreshold := 15.0
	
	isDayByBrightness := avgBrightness > dayThreshold
	hasGoodContrast := avgContrast > contrastThreshold
	
	// If very bright, definitely day. If dark but high contrast, might be night with artificial lighting
	return isDayByBrightness || (avgBrightness > 40 && !hasGoodContrast)
}

// detectObjectsEnhanced performs sophisticated object detection using enhanced image analysis
func detectObjectsEnhanced(img image.Image) ([]string, []ObjectDetail) {
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	if width == 0 || height == 0 {
		return []string{"general scene"}, []ObjectDetail{}
	}

	objects := []string{}
	details := []ObjectDetail{}
	
	// Enhanced analysis using multiple detection algorithms
	results := make(map[string]float64)
	
	// Color pattern analysis
	results = analyzeColorPatternsEnhanced(img, results)
	
	// Edge and texture analysis  
	results = analyzeEdgesAndTextures(img, results)
	
	// Shape and pattern analysis
	results = analyzeShapesAndPatterns(img, results)
	
	// Motion/blur analysis (can indicate living objects)
	results = analyzeMotionBlur(img, results)
	
	// Convert results to objects and details
	confidenceThreshold := 0.3
	
	for category, confidence := range results {
		if confidence > confidenceThreshold {
			detail := ObjectDetail{
				Class:      category,
				Confidence: float32(confidence),
				Category:   categorizeObject(category),
				BBox:       generateFakeBBox(bounds), // Generate approximate bounding box for enhanced detection
			}
			details = append(details, detail)
			
			finalCategory := categorizeObject(category)
			if !contains(objects, finalCategory) {
				objects = append(objects, finalCategory)
			}
		}
	}

	if len(objects) == 0 {
		objects = append(objects, "general scene")
	}

	return objects, details
}

// analyzeColorPatternsEnhanced performs enhanced color pattern analysis
func analyzeColorPatternsEnhanced(img image.Image, results map[string]float64) map[string]float64 {
	bounds := img.Bounds()
	sampleSize := 10 // Dense sampling for better accuracy
	
	// Color category counters
	vegetation := 0
	sky := 0
	earth := 0
	artificial := 0
	skin := 0
	metal := 0
	totalSamples := 0
	
	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleSize {
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleSize {
			r, g, b, _ := img.At(x, y).RGBA()
			red := uint8(r >> 8)
			green := uint8(g >> 8)
			blue := uint8(b >> 8)

			// Enhanced color classification
			if isVegetationColor(red, green, blue) {
				vegetation++
			} else if isSkyColor(red, green, blue) {
				sky++
			} else if isEarthColor(red, green, blue) {
				earth++
			} else if isSkinColor(red, green, blue) {
				skin++
			} else if isMetalColor(red, green, blue) {
				metal++
			} else if isArtificialColor(red, green, blue) {
				artificial++
			}
			totalSamples++
		}
	}
	
	if totalSamples > 0 {
		vegPercent := float64(vegetation) / float64(totalSamples)
		skyPercent := float64(sky) / float64(totalSamples)
		earthPercent := float64(earth) / float64(totalSamples)
		skinPercent := float64(skin) / float64(totalSamples)
		metalPercent := float64(metal) / float64(totalSamples)
		artificialPercent := float64(artificial) / float64(totalSamples)
		
		if vegPercent > 0.15 {
			results["vegetation"] = vegPercent * 2.0
		}
		if skyPercent > 0.1 {
			results["sky"] = skyPercent * 1.5
		}
		if earthPercent > 0.1 && earthPercent < 0.4 {
			results["animal"] = earthPercent * 2.5 // Earth tones might indicate animals
		}
		if skinPercent > 0.02 {
			results["human"] = skinPercent * 5.0 // Even small skin areas indicate humans
		}
		if metalPercent > 0.1 {
			results["machinery"] = metalPercent * 3.0
		}
		if artificialPercent > 0.15 {
			results["vehicle"] = artificialPercent * 2.0
		}
	}
	
	return results
}

// analyzeEdgesAndTextures analyzes edges and textures to detect objects
func analyzeEdgesAndTextures(img image.Image, results map[string]float64) map[string]float64 {
	bounds := img.Bounds()
	
	// Simple edge detection using brightness gradients
	edgeCount := 0
	highFreqVariation := 0.0
	totalSamples := 0
	sampleSize := 15
	
	for y := bounds.Min.Y + sampleSize; y < bounds.Max.Y - sampleSize; y += sampleSize {
		for x := bounds.Min.X + sampleSize; x < bounds.Max.X - sampleSize; x += sampleSize {
			// Get current pixel brightness
			r, g, b, _ := img.At(x, y).RGBA()
			brightness := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
			
			// Check gradients in 4 directions
			directions := [][]int{{0, sampleSize}, {sampleSize, 0}, {0, -sampleSize}, {-sampleSize, 0}}
			maxGradient := 0.0
			totalGradient := 0.0
			
			for _, dir := range directions {
				nx, ny := x + dir[0], y + dir[1]
				if nx >= bounds.Min.X && nx < bounds.Max.X && ny >= bounds.Min.Y && ny < bounds.Max.Y {
					r2, g2, b2, _ := img.At(nx, ny).RGBA()
					brightness2 := 0.299*float64(r2>>8) + 0.587*float64(g2>>8) + 0.114*float64(b2>>8)
					gradient := math.Abs(brightness - brightness2)
					totalGradient += gradient
					if gradient > maxGradient {
						maxGradient = gradient
					}
				}
			}
			
			avgGradient := totalGradient / 4.0
			if maxGradient > 30 {
				edgeCount++
			}
			
			highFreqVariation += avgGradient
			totalSamples++
		}
	}
	
	if totalSamples > 0 {
		edgeRatio := float64(edgeCount) / float64(totalSamples)
		avgVariation := highFreqVariation / float64(totalSamples)
		
		// High edge density suggests machinery/structures
		if edgeRatio > 0.2 {
			results["machinery"] = edgeRatio * 2.0
		} else if edgeRatio > 0.1 {
			results["vehicle"] = edgeRatio * 1.5
		}
		
		// High variation suggests organic/textured surfaces
		if avgVariation > 15 && edgeRatio < 0.3 {
			current := results["animal"]
			results["animal"] = current + (avgVariation / 50.0)
		}
	}
	
	return results
}

// analyzeShapesAndPatterns analyzes geometric shapes and patterns
func analyzeShapesAndPatterns(img image.Image, results map[string]float64) map[string]float64 {
	bounds := img.Bounds()
	
	// Look for regular patterns that might indicate artificial objects
	horizontalLines := 0
	verticalLines := 0
	diagonalPatterns := 0
	totalSamples := 0
	
	sampleSize := 20
	lineThreshold := 5 // minimum consecutive similar pixels to count as a line
	
	for y := bounds.Min.Y; y < bounds.Max.Y - sampleSize*lineThreshold; y += sampleSize {
		for x := bounds.Min.X; x < bounds.Max.X - sampleSize*lineThreshold; x += sampleSize {
			// Check for horizontal lines
			horizontalConsistency := checkLineConsistency(img, x, y, sampleSize, 0, lineThreshold)
			if horizontalConsistency > 0.7 {
				horizontalLines++
			}
			
			// Check for vertical lines
			verticalConsistency := checkLineConsistency(img, x, y, 0, sampleSize, lineThreshold)
			if verticalConsistency > 0.7 {
				verticalLines++
			}
			
			// Check for diagonal patterns
			diagonalConsistency := checkLineConsistency(img, x, y, sampleSize, sampleSize, lineThreshold)
			if diagonalConsistency > 0.6 {
				diagonalPatterns++
			}
			
			totalSamples++
		}
	}
	
	if totalSamples > 0 {
		hLineRatio := float64(horizontalLines) / float64(totalSamples)
		vLineRatio := float64(verticalLines) / float64(totalSamples)
		
		// Regular geometric patterns suggest artificial objects
		if hLineRatio > 0.05 || vLineRatio > 0.05 {
			results["structures"] = (hLineRatio + vLineRatio) * 3.0
		}
		
		if hLineRatio > 0.02 && vLineRatio > 0.02 {
			// Both horizontal and vertical lines suggest machinery/vehicles
			current := results["machinery"]
			results["machinery"] = current + ((hLineRatio + vLineRatio) * 2.0)
		}
	}
	
	return results
}

// analyzeMotionBlur analyzes motion blur that might indicate moving objects
func analyzeMotionBlur(img image.Image, results map[string]float64) map[string]float64 {
	bounds := img.Bounds()
	
	// Detect motion blur by analyzing directional gradients
	horizontalBlur := 0.0
	verticalBlur := 0.0
	totalSamples := 0
	sampleSize := 25
	
	for y := bounds.Min.Y + sampleSize; y < bounds.Max.Y - sampleSize; y += sampleSize {
		for x := bounds.Min.X + sampleSize; x < bounds.Max.X - sampleSize; x += sampleSize {
			// Check horizontal blur (compare with pixels to left and right)
			r1, g1, b1, _ := img.At(x-sampleSize, y).RGBA()
			r2, g2, b2, _ := img.At(x+sampleSize, y).RGBA()
			brightness1 := 0.299*float64(r1>>8) + 0.587*float64(g1>>8) + 0.114*float64(b1>>8)
			brightness2 := 0.299*float64(r2>>8) + 0.587*float64(g2>>8) + 0.114*float64(b2>>8)
			
			hBlur := math.Abs(brightness1 - brightness2)
			horizontalBlur += hBlur
			
			// Check vertical blur
			r3, g3, b3, _ := img.At(x, y-sampleSize).RGBA()
			r4, g4, b4, _ := img.At(x, y+sampleSize).RGBA()
			brightness3 := 0.299*float64(r3>>8) + 0.587*float64(g3>>8) + 0.114*float64(b3>>8)
			brightness4 := 0.299*float64(r4>>8) + 0.587*float64(g4>>8) + 0.114*float64(b4>>8)
			
			vBlur := math.Abs(brightness3 - brightness4)
			verticalBlur += vBlur
			
			totalSamples++
		}
	}
	
	if totalSamples > 0 {
		avgHBlur := horizontalBlur / float64(totalSamples)
		avgVBlur := verticalBlur / float64(totalSamples)
		
		// Motion blur might indicate moving animals or vehicles
		if avgHBlur > 20 || avgVBlur > 20 {
			current := results["animal"]
			results["animal"] = current + 0.3
			
			currentVehicle := results["vehicle"]
			results["vehicle"] = currentVehicle + 0.2
		}
	}
	
	return results
}

// Helper function to check line consistency for pattern detection
func checkLineConsistency(img image.Image, startX, startY, deltaX, deltaY, length int) float64 {
	bounds := img.Bounds()
	if startX + deltaX*length >= bounds.Max.X || startY + deltaY*length >= bounds.Max.Y {
		return 0.0
	}
	
	// Get the first pixel as reference
	r1, g1, b1, _ := img.At(startX, startY).RGBA()
	refBrightness := 0.299*float64(r1>>8) + 0.587*float64(g1>>8) + 0.114*float64(b1>>8)
	
	consistentPixels := 1 // Count the first pixel
	tolerance := 30.0 // Brightness tolerance for considering pixels similar
	
	for i := 1; i < length; i++ {
		x := startX + deltaX*i
		y := startY + deltaY*i
		
		if x >= bounds.Max.X || y >= bounds.Max.Y {
			break
		}
		
		r, g, b, _ := img.At(x, y).RGBA()
		brightness := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
		
		if math.Abs(brightness - refBrightness) < tolerance {
			consistentPixels++
		}
	}
	
	return float64(consistentPixels) / float64(length)
}

// Enhanced color detection functions
func isVegetationColor(r, g, b uint8) bool {
	// Various shades of green for vegetation
	return (g > r && g > b && g > 60) || // Standard green
		   (g > 80 && r < g-20 && b < g-20) || // Dark green
		   (g > 100 && r > 40 && r < g && b < g-30) // Yellow-green
}

func isSkyColor(r, g, b uint8) bool {
	// Blue sky colors
	return (b > r && b > g && b > 80) || // Standard blue
		   (b > 120 && r < b-30 && g < b-20) || // Deep blue
		   (r > 200 && g > 200 && b > 220) // Light sky/clouds
}

func isEarthColor(r, g, b uint8) bool {
	// Brown, tan, earth tones that might indicate animals or ground
	return (r > 60 && g > 40 && b > 20 && r > g && g > b) || // Brown
		   (r > 80 && g > 60 && b < 50 && r-g < 40) || // Tan
		   (r > 50 && g > 30 && b < 30 && r > b+20) // Reddish brown
}

func isSkinColor(r, g, b uint8) bool {
	// Human skin color detection (various ethnicities)
	return (r > 95 && g > 40 && b > 20 && r > g && g > b && 
			r-g > 15 && r-b > 15) || // Light skin
		   (r > 60 && g > 30 && b > 15 && r > g && g >= b &&
			r-g < 45 && r-b > 10) // Darker skin
}

func isMetalColor(r, g, b uint8) bool {
	// Metallic surfaces (gray/silver colors)
	diff := func(a, b uint8) uint8 {
		if a > b {
			return a - b
		}
		return b - a
	}
	
	avgColor := (uint16(r) + uint16(g) + uint16(b)) / 3
	
	return diff(r, uint8(avgColor)) < 15 && 
		   diff(g, uint8(avgColor)) < 15 && 
		   diff(b, uint8(avgColor)) < 15 && 
		   avgColor > 80 && avgColor < 200 // Mid-range gray
}

func isArtificialColor(r, g, b uint8) bool {
	// Bright artificial colors often found on vehicles/machinery
	maxColor := r
	if g > maxColor {
		maxColor = g
	}
	if b > maxColor {
		maxColor = b
	}
	
	minColor := r
	if g < minColor {
		minColor = g
	}
	if b < minColor {
		minColor = b
	}
	
	// High saturation (difference between max and min) suggests artificial colors
	saturation := maxColor - minColor
	
	return saturation > 60 && maxColor > 100
}
func categorizeObject(className string) string {
	animals := []string{"bird", "cat", "dog", "horse", "sheep", "cow", "elephant", "bear", "zebra", "giraffe", "animal"}
	vehicles := []string{"car", "motorcycle", "bus", "train", "truck", "bicycle", "airplane", "boat", "vehicle"}
	humans := []string{"person", "human"}
	machinery := []string{"truck", "tractor", "bus", "train", "machinery"}
	
	className = strings.ToLower(className)
	
	if contains(humans, className) {
		return "human"
	}
	if contains(animals, className) {
		return "animal"
	}
	if contains(machinery, className) {
		return "machinery"
	}
	if contains(vehicles, className) {
		return "vehicle"
	}
	
	return className
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// generateFakeBBox creates a generic bounding box for enhanced detection (non-OpenCV)
// Since enhanced detection doesn't provide precise location data, we create a representative box
func generateFakeBBox(bounds image.Rectangle) *BoundingBox {
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y
	
	// Create a centered box that's about 1/3 of the image size
	boxWidth := width / 3
	boxHeight := height / 3
	x := width/2 - boxWidth/2
	y := height/2 - boxHeight/2
	
	return &BoundingBox{
		X:      x,
		Y:      y,
		Width:  boxWidth,
		Height: boxHeight,
	}
}

// calculateOverallConfidence computes an overall confidence score from individual detections
func calculateOverallConfidence(details []ObjectDetail) float32 {
	if len(details) == 0 {
		return 0.5 // Default confidence when no specific detections
	}
	
	// Calculate weighted average confidence based on detection quality
	totalWeight := float32(0)
	weightedSum := float32(0)
	
	for _, detail := range details {
		// Weight higher confidence detections more heavily
		weight := detail.Confidence * detail.Confidence // Square for emphasis
		weightedSum += detail.Confidence * weight
		totalWeight += weight
	}
	
	if totalWeight == 0 {
		return 0.5
	}
	
	overall := weightedSum / totalWeight
	
	// Ensure confidence is within valid range
	if overall > 1.0 {
		overall = 1.0
	} else if overall < 0.0 {
		overall = 0.0
	}
	
	return overall
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