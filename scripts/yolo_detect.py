#!/usr/bin/env python3
"""
YOLO Object Detection Wrapper Script
This script provides a command-line interface to run YOLO object detection on images
and return results as JSON.

Requirements:
- OpenCV (cv2)
- NumPy
- A YOLO model file (e.g., yolov5s.onnx)

Usage:
    ./yolo_detect.py --image path/to/image.jpg --model path/to/model.onnx --json
"""

import argparse
import json
import sys
import os

def detect_objects(image_path, model_path, confidence_threshold=0.5):
    """
    Perform object detection on an image using YOLO.
    
    Args:
        image_path: Path to input image
        model_path: Path to YOLO model file (.onnx)
        confidence_threshold: Minimum confidence for detections
        
    Returns:
        Dictionary with detections and summary
    """
    try:
        import cv2
        import numpy as np
    except ImportError as e:
        return {
            "detections": [],
            "image_path": image_path,
            "summary": f"Missing required library: {e}. Install with: pip install opencv-python numpy"
        }
    
    # Check if files exist
    if not os.path.exists(image_path):
        return {
            "detections": [],
            "image_path": image_path,
            "summary": f"Error: Image file not found: {image_path}"
        }
    
    if not os.path.exists(model_path):
        return {
            "detections": [],
            "image_path": image_path,
            "summary": f"Error: Model file not found: {model_path}"
        }
    
    # COCO class names (80 classes)
    class_names = [
        "person", "bicycle", "car", "motorcycle", "airplane", "bus", "train", "truck", "boat",
        "traffic light", "fire hydrant", "stop sign", "parking meter", "bench", "bird", "cat",
        "dog", "horse", "sheep", "cow", "elephant", "bear", "zebra", "giraffe", "backpack",
        "umbrella", "handbag", "tie", "suitcase", "frisbee", "skis", "snowboard", "sports ball",
        "kite", "baseball bat", "baseball glove", "skateboard", "surfboard", "tennis racket",
        "bottle", "wine glass", "cup", "fork", "knife", "spoon", "bowl", "banana", "apple",
        "sandwich", "orange", "broccoli", "carrot", "hot dog", "pizza", "donut", "cake", "chair",
        "couch", "potted plant", "bed", "dining table", "toilet", "tv", "laptop", "mouse",
        "remote", "keyboard", "cell phone", "microwave", "oven", "toaster", "sink", "refrigerator",
        "book", "clock", "vase", "scissors", "teddy bear", "hair drier", "toothbrush"
    ]
    
    try:
        # Load the image
        image = cv2.imread(image_path)
        if image is None:
            return {
                "detections": [],
                "image_path": image_path,
                "summary": "Error: Failed to load image"
            }
        
        height, width = image.shape[:2]
        
        # Load the YOLO model
        net = cv2.dnn.readNetFromONNX(model_path)
        
        # Prepare the image for YOLO (640x640 input)
        blob = cv2.dnn.blobFromImage(image, 1/255.0, (640, 640), swapRB=True, crop=False)
        net.setInput(blob)
        
        # Run inference
        outputs = net.forward()
        
        # Process detections
        detections = []
        
        # YOLOv5 output format: [batch, num_detections, 85]
        # 85 = x, y, w, h, confidence, 80 class scores
        for detection in outputs[0]:
            scores = detection[5:]
            class_id = np.argmax(scores)
            confidence = scores[class_id] * detection[4]  # objectness * class score
            
            if confidence > confidence_threshold:
                # Scale bounding box back to original image size
                center_x = int(detection[0] * width)
                center_y = int(detection[1] * height)
                w = int(detection[2] * width)
                h = int(detection[3] * height)
                
                x = int(center_x - w / 2)
                y = int(center_y - h / 2)
                
                detections.append({
                    "class_name": class_names[class_id] if class_id < len(class_names) else f"class_{class_id}",
                    "confidence": float(confidence),
                    "x": float(x),
                    "y": float(y),
                    "width": float(w),
                    "height": float(h)
                })
        
        # Generate summary
        if not detections:
            summary = "No objects detected"
        else:
            # Count objects by class
            class_counts = {}
            for det in detections:
                class_name = det["class_name"]
                class_counts[class_name] = class_counts.get(class_name, 0) + 1
            
            # Build summary text
            parts = []
            for class_name, count in class_counts.items():
                if count == 1:
                    parts.append(f"one {class_name}")
                else:
                    parts.append(f"{count} {class_name}s")
            
            summary = "The photo includes: " + ", ".join(parts)
            
            # Add day/night heuristic
            has_person_or_animal = any(
                d["class_name"] in ["person", "bird", "cat", "dog", "horse", "sheep", "cow"] 
                for d in detections
            )
            if has_person_or_animal:
                summary = "It's day time. " + summary
        
        return {
            "detections": detections,
            "image_path": image_path,
            "summary": summary
        }
        
    except Exception as e:
        return {
            "detections": [],
            "image_path": image_path,
            "summary": f"Error during detection: {str(e)}"
        }


def main():
    parser = argparse.ArgumentParser(description="YOLO Object Detection for Timelapse")
    parser.add_argument("--image", required=True, help="Path to input image")
    parser.add_argument("--model", required=True, help="Path to YOLO model (.onnx)")
    parser.add_argument("--confidence", type=float, default=0.5, help="Confidence threshold (default: 0.5)")
    parser.add_argument("--json", action="store_true", help="Output results as JSON")
    
    args = parser.parse_args()
    
    result = detect_objects(args.image, args.model, args.confidence)
    
    if args.json:
        print(json.dumps(result, indent=2))
    else:
        print(f"Image: {result['image_path']}")
        print(f"Summary: {result['summary']}")
        print(f"Detections: {len(result['detections'])}")
        for i, det in enumerate(result['detections'], 1):
            print(f"  {i}. {det['class_name']}: {det['confidence']:.2f} at ({det['x']:.0f}, {det['y']:.0f})")
    
    return 0


if __name__ == "__main__":
    sys.exit(main())
