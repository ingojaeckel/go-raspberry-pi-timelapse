//go:build !opencv
// +build !opencv

package detection

import (
	"fmt"
	"log"
)

// analyzeWithNativeOpenCV provides a fallback when OpenCV is not available
func analyzeWithNativeOpenCV(photoPath string, timeout_unused interface{}) (*DetectionResult, error) {
	log.Printf("Native OpenCV not available (not compiled with opencv build tag), falling back to enhanced detection")
	return nil, fmt.Errorf("OpenCV not available - compile with -tags opencv to enable native Go OpenCV support")
}