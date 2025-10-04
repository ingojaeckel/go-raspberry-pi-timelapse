#include "object_detector.hpp"
#include <fstream>
#include <iostream>
#include <algorithm>

// Maximum distance (in pixels) an object can move between frames to be considered the same object
// This assumes objects don't teleport across large portions of the frame
// For 720p video at typical frame rates (5 fps), this allows for reasonable movement
constexpr float MAX_MOVEMENT_DISTANCE = 100.0f;

ObjectDetector::ObjectDetector(const std::string& model_path,
                              const std::string& config_path,
                              const std::string& classes_path,
                              double confidence_threshold,
                              std::shared_ptr<Logger> logger,
                              DetectionModelFactory::ModelType model_type)
    : model_path_(model_path), config_path_(config_path), classes_path_(classes_path),
      confidence_threshold_(confidence_threshold), logger_(logger), model_type_(model_type),
      initialized_(false) {
}

ObjectDetector::~ObjectDetector() = default;

bool ObjectDetector::initialize() {
    if (initialized_) {
        return true;
    }

    logger_->info("Initializing object detector with model abstraction...");
    
    // Create the detection model using the factory
    try {
        detection_model_ = DetectionModelFactory::createModel(model_type_, logger_);
        if (!detection_model_) {
            logger_->error("Failed to create detection model");
            return false;
        }
        
        // Initialize the model
        if (!detection_model_->initialize(model_path_, config_path_, classes_path_, confidence_threshold_)) {
            logger_->error("Failed to initialize detection model");
            return false;
        }
        
        // Warm up the model for accurate performance measurements
        detection_model_->warmUp();
        
        initialized_ = true;
        logger_->info("Object detector initialized successfully with " + detection_model_->getModelName());
        
        // Log model performance characteristics
        auto metrics = detection_model_->getMetrics();
        logger_->info("Model: " + metrics.model_name + " - " + metrics.description);
        logger_->info("Expected inference time: ~" + std::to_string(metrics.avg_inference_time_ms) + "ms");
        logger_->info("Model accuracy: " + std::to_string(static_cast<int>(metrics.accuracy_score * 100)) + "%");
        
        return true;
        
    } catch (const std::exception& e) {
        logger_->error("Exception during model initialization: " + std::string(e.what()));
        return false;
    }
}

std::vector<Detection> ObjectDetector::detectObjects(const cv::Mat& frame) {
    if (!initialized_ || !detection_model_ || frame.empty()) {
        return {};
    }

    return detection_model_->detect(frame);
}

void ObjectDetector::processFrame(const cv::Mat& frame) {
    if (!initialized_ || !detection_model_) {
        return;
    }

    // Detect objects in the current frame
    auto detections = detectObjects(frame);
    
    // Filter for target classes only
    std::vector<Detection> target_detections;
    for (const auto& detection : detections) {
        if (isTargetClass(detection.class_name)) {
            target_detections.push_back(detection);
        }
    }

    // Update tracking and log events
    updateTrackedObjects(target_detections);
    logObjectEvents(target_detections);
}

std::vector<std::string> ObjectDetector::getTargetClasses() {
    return {
        "person",      // People
        "car",         // Vehicles
        "truck",
        "bus", 
        "motorcycle",
        "bicycle",
        "cat",         // Small animals
        "dog"
        // Note: "fox" is not in COCO dataset, but we can detect it as similar to cat/dog
    };
}

bool ObjectDetector::isTargetClass(const std::string& class_name) const {
    auto target_classes = getTargetClasses();
    return std::find(target_classes.begin(), target_classes.end(), class_name) != target_classes.end();
}

ModelMetrics ObjectDetector::getModelMetrics() const {
    if (!detection_model_) {
        return {"Unknown", "Unknown", 0.0, 0, 0, "Model not initialized"};
    }
    return detection_model_->getMetrics();
}

bool ObjectDetector::switchModel(DetectionModelFactory::ModelType new_model_type) {
    logger_->info("Switching to model type: " + DetectionModelFactory::modelTypeToString(new_model_type));
    
    try {
        // Create new model
        auto new_model = DetectionModelFactory::createModel(new_model_type, logger_);
        if (!new_model) {
            logger_->error("Failed to create new detection model");
            return false;
        }
        
        // Initialize new model
        if (!new_model->initialize(model_path_, config_path_, classes_path_, confidence_threshold_)) {
            logger_->error("Failed to initialize new detection model");
            return false;
        }
        
        // Warm up new model
        new_model->warmUp();
        
        // Replace old model
        detection_model_ = std::move(new_model);
        model_type_ = new_model_type;
        
        logger_->info("Successfully switched to " + detection_model_->getModelName());
        
        // Log new model characteristics
        auto metrics = detection_model_->getMetrics();
        logger_->info("New model performance - Accuracy: " + std::to_string(static_cast<int>(metrics.accuracy_score * 100)) + 
                     "%, Expected inference: ~" + std::to_string(metrics.avg_inference_time_ms) + "ms");
        
        return true;
        
    } catch (const std::exception& e) {
        logger_->error("Exception during model switch: " + std::string(e.what()));
        return false;
    }
}

std::vector<ModelMetrics> ObjectDetector::getAvailableModels() {
    return DetectionModelFactory::getAvailableModels();
}

void ObjectDetector::updateTrackedObjects(const std::vector<Detection>& detections) {
    // Object tracking and permanence model:
    // - Track objects frame-to-frame based on (x, y) position and object type
    // - Determine if detected object is "new" (entered frame) or "moved" (was near this position before)
    // - Use MAX_MOVEMENT_DISTANCE threshold to decide: if distance > threshold, consider it a new object
    
    // Mark all current objects as not seen this frame
    for (auto& tracked : tracked_objects_) {
        tracked.was_present_last_frame = false;
        tracked.frames_since_detection++;
    }
    
    // Update with current detections
    for (const auto& detection : detections) {
        bool found_existing = false;
        
        cv::Point2f detection_center(
            detection.bbox.x + detection.bbox.width / 2.0f,
            detection.bbox.y + detection.bbox.height / 2.0f
        );
        
        // Try to match with existing tracked objects of the same type
        for (auto& tracked : tracked_objects_) {
            if (tracked.object_type == detection.class_name) {
                // Calculate distance from previously tracked position
                float distance = cv::norm(tracked.center - detection_center);
                
                // Check if this detection is close enough to be the same object
                // If distance is within threshold, assume it's the same object that moved
                if (distance < MAX_MOVEMENT_DISTANCE) {
                    // Store previous position before updating (for movement logging)
                    tracked.previous_center = tracked.center;
                    tracked.center = detection_center;
                    tracked.was_present_last_frame = true;
                    tracked.frames_since_detection = 0;
                    tracked.is_new = false;  // Not new, it's been tracked
                    found_existing = true;
                    break;
                }
            }
        }
        
        // Add new object if not found within threshold distance
        // This means either:
        // 1. First time seeing this object type, OR
        // 2. Object of this type is too far from any previously tracked position (likely a different object)
        if (!found_existing) {
            ObjectTracker new_tracker;
            new_tracker.object_type = detection.class_name;
            new_tracker.center = detection_center;
            new_tracker.previous_center = detection_center;  // Same as current for new object
            new_tracker.was_present_last_frame = true;
            new_tracker.frames_since_detection = 0;
            new_tracker.is_new = true;  // Mark as newly entered
            tracked_objects_.push_back(new_tracker);
        }
    }
    
    // Remove objects that haven't been seen for too long
    tracked_objects_.erase(
        std::remove_if(tracked_objects_.begin(), tracked_objects_.end(),
                      [](const ObjectTracker& tracker) {
                          return tracker.frames_since_detection > 30; // 30 frames threshold
                      }),
        tracked_objects_.end());
}


void ObjectDetector::logObjectEvents(const std::vector<Detection>& current_detections) {
    // Log enter/movement events based on tracking
    for (const auto& tracked : tracked_objects_) {
        // Find the current detection for this tracked object
        auto detection_it = std::find_if(current_detections.begin(), current_detections.end(),
                                        [&tracked](const Detection& det) {
                                            return det.class_name == tracked.object_type;
                                        });
        
        // Check if object is currently present in this frame
        bool currently_present = (detection_it != current_detections.end());
        
        if (currently_present && tracked.frames_since_detection == 0) {
            // Object is present in this frame and was just updated
            
            if (tracked.is_new) {
                // New object entered the frame
                logger_->logObjectEntry(
                    tracked.object_type,
                    tracked.center.x,
                    tracked.center.y,
                    detection_it->confidence
                );
            } else {
                // Object was seen before - check if it moved
                float distance = cv::norm(tracked.center - tracked.previous_center);
                
                // Only log movement if the object actually moved a meaningful distance
                // (avoid logging tiny movements due to detection jitter)
                if (distance > 5.0f) {  // 5 pixel threshold to avoid logging noise
                    logger_->logObjectMovement(
                        tracked.object_type,
                        tracked.previous_center.x,
                        tracked.previous_center.y,
                        tracked.center.x,
                        tracked.center.y,
                        detection_it->confidence
                    );
                }
            }
        }
    }
    
    // Note: Exit detection would require more sophisticated tracking
    // For now, we focus on entry and movement detection which are more reliable
}
