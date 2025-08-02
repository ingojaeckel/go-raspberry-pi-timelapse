#!/usr/bin/env python3
"""
Advanced object detection using OpenCV and pre-trained models.
This script provides OpenCV-level object detection capabilities for the timelapse camera.
"""

import cv2
import numpy as np
import json
import sys
import os
import argparse
from typing import List, Dict, Tuple, Optional

class OpenCVObjectDetector:
    """High-accuracy object detection using OpenCV and YOLO."""
    
    def __init__(self):
        self.net = None
        self.classes = []
        self.colors = []
        self.output_layers = []
        self.confidence_threshold = 0.5
        self.nms_threshold = 0.4
        
        # Initialize the detector
        self._initialize_detector()
    
    def _initialize_detector(self):
        """Initialize YOLO detector with pre-trained weights."""
        try:
            # Try to load YOLO model files
            # These would typically be downloaded during setup
            weights_path = "/opt/yolo/yolov3.weights"
            config_path = "/opt/yolo/yolov3.cfg"
            classes_path = "/opt/yolo/coco.names"
            
            # Fallback to a smaller model if available
            if not os.path.exists(weights_path):
                weights_path = "/opt/yolo/yolov3-tiny.weights"
                config_path = "/opt/yolo/yolov3-tiny.cfg"
            
            if os.path.exists(weights_path) and os.path.exists(config_path):
                self.net = cv2.dnn.readNetFromDarknet(config_path, weights_path)
                self.net.setPreferableBackend(cv2.dnn.DNN_BACKEND_OPENCV)
                self.net.setPreferableTarget(cv2.dnn.DNN_TARGET_CPU)
                
                layer_names = self.net.getLayerNames()
                self.output_layers = [layer_names[i[0] - 1] for i in self.net.getUnconnectedOutLayers()]
            
            # Load class names
            if os.path.exists(classes_path):
                with open(classes_path, 'r') as f:
                    self.classes = [line.strip() for line in f.readlines()]
            else:
                # COCO dataset classes
                self.classes = [
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
                    "hair drier", "toothbrush"
                ]
            
            # Generate colors for each class
            self.colors = np.random.uniform(0, 255, size=(len(self.classes), 3))
            
        except Exception as e:
            print(f"Warning: Could not initialize YOLO detector: {e}", file=sys.stderr)
            self.net = None
    
    def _analyze_day_night(self, image: np.ndarray) -> bool:
        """Enhanced day/night detection using OpenCV."""
        # Convert to grayscale
        gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
        
        # Calculate average brightness
        avg_brightness = np.mean(gray)
        
        # Calculate histogram to understand brightness distribution
        hist = cv2.calcHist([gray], [0], None, [256], [0, 256])
        
        # Check if most pixels are in the brighter range
        bright_pixels = np.sum(hist[128:])
        dark_pixels = np.sum(hist[:128])
        
        # Enhanced analysis considering both average brightness and distribution
        is_day = avg_brightness > 80 or (avg_brightness > 50 and bright_pixels > dark_pixels * 1.5)
        
        return bool(is_day)
    
    def _detect_with_yolo(self, image: np.ndarray) -> List[Dict]:
        """Perform object detection using YOLO."""
        if self.net is None:
            return []
        
        height, width = image.shape[:2]
        
        # Create blob from image
        blob = cv2.dnn.blobFromImage(image, 1/255.0, (416, 416), swapRB=True, crop=False)
        self.net.setInput(blob)
        
        # Run inference
        outputs = self.net.forward(self.output_layers)
        
        # Parse detections
        boxes = []
        confidences = []
        class_ids = []
        
        for output in outputs:
            for detection in output:
                scores = detection[5:]
                class_id = np.argmax(scores)
                confidence = scores[class_id]
                
                if confidence > self.confidence_threshold:
                    # Get bounding box coordinates
                    center_x = int(detection[0] * width)
                    center_y = int(detection[1] * height)
                    w = int(detection[2] * width)
                    h = int(detection[3] * height)
                    
                    x = int(center_x - w/2)
                    y = int(center_y - h/2)
                    
                    boxes.append([x, y, w, h])
                    confidences.append(float(confidence))
                    class_ids.append(class_id)
        
        # Apply non-maximum suppression
        indices = cv2.dnn.NMSBoxes(boxes, confidences, self.confidence_threshold, self.nms_threshold)
        
        detections = []
        if len(indices) > 0:
            for i in indices.flatten():
                class_name = self.classes[class_ids[i]] if class_ids[i] < len(self.classes) else "unknown"
                detections.append({
                    "class": class_name,
                    "confidence": confidences[i],
                    "category": self._categorize_object(class_name),
                    "bbox": boxes[i]
                })
        
        return detections
    
    def _detect_with_opencv_features(self, image: np.ndarray) -> List[Dict]:
        """Advanced feature-based detection using OpenCV algorithms."""
        detections = []
        
        # Convert to different color spaces for analysis
        gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
        hsv = cv2.cvtColor(image, cv2.COLOR_BGR2HSV)
        
        # 1. Human detection using HOG
        detections.extend(self._detect_humans_hog(image, gray))
        
        # 2. Vehicle detection using contours and features
        detections.extend(self._detect_vehicles_contours(image, gray))
        
        # 3. Animal detection using texture analysis
        detections.extend(self._detect_animals_texture(image, gray, hsv))
        
        # 4. Machinery detection using edge analysis
        detections.extend(self._detect_machinery_edges(image, gray))
        
        return detections
    
    def _detect_humans_hog(self, image: np.ndarray, gray: np.ndarray) -> List[Dict]:
        """Detect humans using HOG (Histogram of Oriented Gradients)."""
        detections = []
        try:
            # Initialize HOG descriptor for human detection
            hog = cv2.HOGDescriptor()
            hog.setSVMDetector(cv2.HOGDescriptor_getDefaultPeopleDetector())
            
            # Detect people
            (rects, weights) = hog.detectMultiScale(gray, winStride=(4, 4), padding=(8, 8), scale=1.05)
            
            for i, (x, y, w, h) in enumerate(rects):
                confidence = float(weights[i][0]) if i < len(weights) else 0.6
                if confidence > 0.3:
                    detections.append({
                        "class": "person",
                        "confidence": min(confidence, 1.0),
                        "category": "human",
                        "bbox": [x, y, w, h]
                    })
        except Exception as e:
            print(f"HOG detection error: {e}", file=sys.stderr)
        
        return detections
    
    def _detect_vehicles_contours(self, image: np.ndarray, gray: np.ndarray) -> List[Dict]:
        """Detect vehicles using contour analysis."""
        detections = []
        try:
            # Edge detection
            edges = cv2.Canny(gray, 50, 150, apertureSize=3)
            
            # Find contours
            contours, _ = cv2.findContours(edges, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
            
            for contour in contours:
                area = cv2.contourArea(contour)
                if area > 5000:  # Minimum area for vehicles
                    # Get bounding rectangle
                    x, y, w, h = cv2.boundingRect(contour)
                    aspect_ratio = w / h
                    
                    # Check if shape is vehicle-like (rectangular)
                    if 1.2 < aspect_ratio < 4.0 and w > 100 and h > 50:
                        # Analyze colors in the region for artificial/metallic colors
                        roi = image[y:y+h, x:x+w]
                        if self._has_artificial_colors(roi):
                            confidence = min(0.7, area / 50000)
                            detections.append({
                                "class": "vehicle",
                                "confidence": confidence,
                                "category": "vehicle",
                                "bbox": [x, y, w, h]
                            })
        except Exception as e:
            print(f"Vehicle contour detection error: {e}", file=sys.stderr)
        
        return detections
    
    def _detect_animals_texture(self, image: np.ndarray, gray: np.ndarray, hsv: np.ndarray) -> List[Dict]:
        """Detect animals using texture and color analysis."""
        detections = []
        try:
            # LBP (Local Binary Pattern) for texture analysis
            # This is a simplified version - in practice you'd train on animal textures
            
            # Look for organic shapes and earth tones
            lower_brown = np.array([5, 50, 50])
            upper_brown = np.array([25, 255, 200])
            brown_mask = cv2.inRange(hsv, lower_brown, upper_brown)
            
            # Find contours in brown regions
            contours, _ = cv2.findContours(brown_mask, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
            
            for contour in contours:
                area = cv2.contourArea(contour)
                if 2000 < area < 50000:  # Animal-sized objects
                    x, y, w, h = cv2.boundingRect(contour)
                    aspect_ratio = w / h
                    
                    # Animals typically have more organic aspect ratios
                    if 0.5 < aspect_ratio < 3.0:
                        # Check for texture variation (animals have fur/skin texture)
                        roi_gray = gray[y:y+h, x:x+w]
                        texture_var = np.var(roi_gray)
                        
                        if texture_var > 100:  # Good texture variation
                            confidence = min(0.6, (texture_var / 1000) + (area / 100000))
                            detections.append({
                                "class": "animal",
                                "confidence": confidence,
                                "category": "animal",
                                "bbox": [x, y, w, h]
                            })
        except Exception as e:
            print(f"Animal texture detection error: {e}", file=sys.stderr)
        
        return detections
    
    def _detect_machinery_edges(self, image: np.ndarray, gray: np.ndarray) -> List[Dict]:
        """Detect machinery using edge pattern analysis."""
        detections = []
        try:
            # Strong edge detection for machinery (metal has sharp edges)
            edges = cv2.Canny(gray, 100, 200, apertureSize=3)
            
            # Find lines (machinery often has straight lines)
            lines = cv2.HoughLinesP(edges, 1, np.pi/180, threshold=100, minLineLength=50, maxLineGap=10)
            
            if lines is not None and len(lines) > 20:  # Many lines suggest machinery
                # Group nearby lines to find machinery regions
                line_regions = []
                for line in lines:
                    x1, y1, x2, y2 = line[0]
                    line_regions.append([(x1, y1), (x2, y2)])
                
                # Simple clustering of line endpoints to find machinery regions
                if len(line_regions) > 10:
                    # Find bounding box of line concentration
                    all_points = []
                    for (x1, y1), (x2, y2) in line_regions:
                        all_points.extend([(x1, y1), (x2, y2)])
                    
                    if all_points:
                        xs, ys = zip(*all_points)
                        x_min, x_max = min(xs), max(xs)
                        y_min, y_max = min(ys), max(ys)
                        
                        w, h = x_max - x_min, y_max - y_min
                        if w > 100 and h > 100:  # Reasonable machinery size
                            confidence = min(0.7, len(lines) / 100)
                            detections.append({
                                "class": "machinery",
                                "confidence": confidence,
                                "category": "machinery",
                                "bbox": [x_min, y_min, w, h]
                            })
        except Exception as e:
            print(f"Machinery edge detection error: {e}", file=sys.stderr)
        
        return detections
    
    def _has_artificial_colors(self, roi: np.ndarray) -> bool:
        """Check if region has artificial/manufactured colors."""
        if roi.size == 0:
            return False
        
        # Convert to HSV for better color analysis
        hsv_roi = cv2.cvtColor(roi, cv2.COLOR_BGR2HSV)
        
        # Check for high saturation (artificial colors are often more saturated)
        saturation = hsv_roi[:,:,1]
        high_sat_ratio = np.sum(saturation > 100) / saturation.size
        
        return high_sat_ratio > 0.3
    
    def _categorize_object(self, class_name: str) -> str:
        """Categorize detected objects into main categories."""
        animals = ["bird", "cat", "dog", "horse", "sheep", "cow", "elephant", "bear", "zebra", "giraffe"]
        vehicles = ["car", "motorcycle", "bus", "truck", "bicycle", "airplane", "boat"]
        humans = ["person"]
        machinery = ["truck", "bus", "train"]  # Some overlap with vehicles
        
        class_name_lower = class_name.lower()
        
        if class_name_lower in humans:
            return "human"
        elif class_name_lower in animals:
            return "animal"
        elif class_name_lower in machinery:
            return "machinery"
        elif class_name_lower in vehicles:
            return "vehicle"
        else:
            return class_name_lower
    
    def analyze_image(self, image_path: str) -> Dict:
        """Main analysis function."""
        try:
            # Load image
            image = cv2.imread(image_path)
            if image is None:
                raise ValueError(f"Could not load image: {image_path}")
            
            # Analyze day/night
            is_day = self._analyze_day_night(image)
            
            # Perform object detection
            detections = []
            
            # Try YOLO first (if available)
            if self.net is not None:
                detections = self._detect_with_yolo(image)
            
            # If YOLO didn't find much, use OpenCV features
            if len(detections) < 2:
                opencv_detections = self._detect_with_opencv_features(image)
                # Merge detections, avoiding duplicates
                for det in opencv_detections:
                    if not any(d["category"] == det["category"] for d in detections):
                        detections.append(det)
            
            # Extract categories and create summary
            categories = list(set(det["category"] for det in detections))
            if not categories:
                categories = ["general scene"]
            
            # Generate summary
            time_of_day = "day" if is_day else "night"
            objects_text = ", ".join(categories)
            summary = f"It's {time_of_day} time. The photo includes: {objects_text}"
            
            return {
                "isDay": is_day,
                "objects": categories,
                "details": detections,
                "summary": summary,
                "photoPath": image_path
            }
            
        except Exception as e:
            print(f"Error analyzing image: {e}", file=sys.stderr)
            return {
                "isDay": True,
                "objects": ["general scene"],
                "details": [],
                "summary": f"Error analyzing image: {str(e)}",
                "photoPath": image_path
            }

def main():
    parser = argparse.ArgumentParser(description="OpenCV Object Detection for Timelapse Camera")
    parser.add_argument("image_path", help="Path to image file to analyze")
    parser.add_argument("--output-json", action="store_true", help="Output results as JSON")
    
    args = parser.parse_args()
    
    if not os.path.exists(args.image_path):
        print(f"Error: Image file not found: {args.image_path}", file=sys.stderr)
        sys.exit(1)
    
    # Initialize detector
    detector = OpenCVObjectDetector()
    
    # Analyze image
    result = detector.analyze_image(args.image_path)
    
    if args.output_json:
        print(json.dumps(result, indent=2))
    else:
        print(f"Day/Night: {'Day' if result['isDay'] else 'Night'}")
        print(f"Objects detected: {', '.join(result['objects'])}")
        print(f"Summary: {result['summary']}")
        if result['details']:
            print("\nDetailed results:")
            for detail in result['details']:
                print(f"  - {detail['class']}: {detail['confidence']:.2f} confidence ({detail['category']})")

if __name__ == "__main__":
    main()