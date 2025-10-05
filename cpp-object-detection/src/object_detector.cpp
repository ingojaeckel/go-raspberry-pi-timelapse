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
                              DetectionModelFactory::ModelType model_type,
                              double detection_scale_factor)
    : model_path_(model_path), config_path_(config_path), classes_path_(classes_path),
      confidence_threshold_(confidence_threshold), detection_scale_factor_(detection_scale_factor),
      logger_(logger), model_type_(model_type),
      initialized_(false), total_objects_detected_(0) {
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
        if (!detection_model_->initialize(model_path_, config_path_, classes_path_, confidence_threshold_, detection_scale_factor_)) {
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
        "cat",         // Animals
        "dog",
        "bird",
        "bear",        // Wild animals (closest to fox in COCO dataset)
        "chair",       // Furniture and objects
        "book"
        // Note: "fox" is not in COCO dataset, using "bear" as closest wild animal equivalent
        // Note: "painting" is not in COCO dataset
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
        if (!new_model->initialize(model_path_, config_path_, classes_path_, confidence_threshold_, detection_scale_factor_)) {
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
    // - Maintain position history for better movement analysis
    
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
        
        logger_->debug("Processing detection: " + detection.class_name + 
                      " at (" + std::to_string(detection_center.x) + ", " + 
                      std::to_string(detection_center.y) + ")");
        
        // Try to match with existing tracked objects of the same type
        float min_distance = MAX_MOVEMENT_DISTANCE;
        ObjectTracker* best_match = nullptr;
        
        for (auto& tracked : tracked_objects_) {
            if (tracked.object_type == detection.class_name) {
                // Calculate distance from previously tracked position
                float distance = cv::norm(tracked.center - detection_center);
                
                logger_->debug("  Distance to existing " + tracked.object_type + 
                              " at (" + std::to_string(tracked.center.x) + ", " + 
                              std::to_string(tracked.center.y) + "): " + 
                              std::to_string(distance) + " pixels");
                
                // Find the closest matching object within threshold
                if (distance < min_distance) {
                    min_distance = distance;
                    best_match = &tracked;
                }
            }
        }
        
        // Check if we found a match within threshold
        if (best_match != nullptr) {
            logger_->debug("  Matched to existing " + best_match->object_type + 
                          " (distance: " + std::to_string(min_distance) + " pixels)");
            
            // Store previous position before updating (for movement logging)
            best_match->previous_center = best_match->center;
            
            // Add current position to history before updating
            best_match->position_history.push_back(best_match->center);
            if (best_match->position_history.size() > ObjectTracker::MAX_POSITION_HISTORY) {
                best_match->position_history.pop_front();
            }
            
            // Update position
            best_match->center = detection_center;
            best_match->was_present_last_frame = true;
            best_match->frames_since_detection = 0;
            best_match->is_new = false;  // Not new, it's been tracked
            found_existing = true;
            
            // Update stationary status based on movement
            updateStationaryStatus(*best_match);
            
            // Log movement pattern if we have enough history
            if (best_match->position_history.size() >= 3) {
                float total_path_length = 0.0f;
                for (size_t i = 1; i < best_match->position_history.size(); ++i) {
                    total_path_length += cv::norm(best_match->position_history[i] - 
                                                  best_match->position_history[i-1]);
                }
                logger_->debug("  Movement pattern: " + std::to_string(best_match->position_history.size()) + 
                              " positions tracked, total path length: " + 
                              std::to_string(total_path_length) + " pixels");
            }
        }
        
        // Add new object if not found within threshold distance
        // This means either:
        // 1. First time seeing this object type, OR
        // 2. Object of this type is too far from any previously tracked position (likely a different object)
        if (!found_existing) {
            // Check if we're at the tracking limit
            if (tracked_objects_.size() >= MAX_TRACKED_OBJECTS) {
                logger_->warning("Maximum tracked objects limit (" + std::to_string(MAX_TRACKED_OBJECTS) + 
                                ") reached. Cleaning up oldest objects.");
                cleanupOldTrackedObjects();
            }

            logger_->debug("  Creating new tracker for " + detection.class_name + 
                          " (no existing object within " + std::to_string(MAX_MOVEMENT_DISTANCE) + 
                          " pixel threshold)");
            
            ObjectTracker new_tracker;
            new_tracker.object_type = detection.class_name;
            new_tracker.center = detection_center;
            new_tracker.previous_center = detection_center;  // Same as current for new object
            new_tracker.position_history.push_back(detection_center);  // Initialize history
            new_tracker.was_present_last_frame = true;
            new_tracker.frames_since_detection = 0;
            new_tracker.is_new = true;  // Mark as newly entered
            new_tracker.is_stationary = false;  // New objects are not yet stationary
            new_tracker.stationary_since = std::chrono::steady_clock::now();
            tracked_objects_.push_back(new_tracker);
            
            // Update statistics with bounded growth protection
            total_objects_detected_++;
            object_type_counts_[detection.class_name]++;
            
            // Limit object type counts map size
            if (object_type_counts_.size() > MAX_OBJECT_TYPE_ENTRIES) {
                limitObjectTypeCounts();
            }
        }
    }
    
    // Remove objects that haven't been seen for too long
    // First, log the objects that will be removed
    for (const auto& tracker : tracked_objects_) {
        if (tracker.frames_since_detection > 30) {
            logger_->debug("Removing " + tracker.object_type + 
                          " tracker (not seen for " + 
                          std::to_string(tracker.frames_since_detection) + " frames)");
        }
    }
    
    // Count how many will be removed before erasing
    auto removed_count = std::count_if(tracked_objects_.begin(), tracked_objects_.end(),
                                       [](const ObjectTracker& tracker) {
                                           return tracker.frames_since_detection > 30; // 30 frames threshold
                                       });
    
    // Now remove them
    tracked_objects_.erase(
        std::remove_if(tracked_objects_.begin(), tracked_objects_.end(),
                      [](const ObjectTracker& tracker) {
                          return tracker.frames_since_detection > 30; // 30 frames threshold
                      }),
        tracked_objects_.end());
    
    if (removed_count > 0) {
        logger_->debug("Removed " + std::to_string(removed_count) + " stale tracker(s)");
    }
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
                logger_->debug("New object entered: " + tracked.object_type + 
                             " at (" + std::to_string(tracked.center.x) + ", " + 
                             std::to_string(tracked.center.y) + ")");
                logger_->logObjectEntry(
                    tracked.object_type,
                    tracked.center.x,
                    tracked.center.y,
                    detection_it->confidence
                );
                // Record for hourly summary (new objects are typically moving/dynamic)
                logger_->recordDetection(tracked.object_type, false);
            } else {
                // Object was seen before - check if it moved
                float distance = cv::norm(tracked.center - tracked.previous_center);
                
                logger_->debug("Checking movement for " + tracked.object_type + 
                             ": distance = " + std::to_string(distance) + " pixels, " +
                             "from (" + std::to_string(tracked.previous_center.x) + ", " + 
                             std::to_string(tracked.previous_center.y) + ") to (" +
                             std::to_string(tracked.center.x) + ", " + 
                             std::to_string(tracked.center.y) + ")");
                
                // Only log movement if the object actually moved a meaningful distance
                // (avoid logging tiny movements due to detection jitter)
                if (distance > 5.0f) {  // 5 pixel threshold to avoid logging noise
                    // Calculate movement characteristics from position history
                    std::string movement_info = "";
                    if (tracked.position_history.size() >= 2) {
                        // Calculate average movement over recent history
                        float total_distance = 0.0f;
                        for (size_t i = 1; i < tracked.position_history.size(); ++i) {
                            total_distance += cv::norm(tracked.position_history[i] - 
                                                      tracked.position_history[i-1]);
                        }
                        float avg_distance = total_distance / (tracked.position_history.size() - 1);
                        
                        // Determine movement direction from history
                        cv::Point2f overall_direction = tracked.center - tracked.position_history.front();
                        float overall_distance = cv::norm(overall_direction);
                        
                        movement_info = " [avg step: " + std::to_string(avg_distance) + 
                                      " px, overall path: " + std::to_string(overall_distance) + " px]";
                        
                        logger_->debug("Movement analysis for " + tracked.object_type + 
                                     ": " + std::to_string(tracked.position_history.size()) + 
                                     " positions in history, average step size: " + 
                                     std::to_string(avg_distance) + " pixels, " +
                                     "overall displacement: " + std::to_string(overall_distance) + 
                                     " pixels");
                    }
                    
                    logger_->debug("Logging movement: " + tracked.object_type + 
                                 " moved " + std::to_string(distance) + " pixels" + movement_info);
                    
                    logger_->logObjectMovement(
                        tracked.object_type,
                        tracked.previous_center.x,
                        tracked.previous_center.y,
                        tracked.center.x,
                        tracked.center.y,
                        detection_it->confidence
                    );
                    // Record as dynamic object
                    logger_->recordDetection(tracked.object_type, false);
                } else {
                    logger_->debug("Movement below threshold (" + std::to_string(distance) + 
                                 " < 5.0 pixels) - not logging");
                }
            }
        }
    }
    
    // Note: Exit detection would require more sophisticated tracking
    // For now, we focus on entry and movement detection which are more reliable
}

int ObjectDetector::getTotalObjectsDetected() const {
    return total_objects_detected_;
}

std::vector<std::pair<std::string, int>> ObjectDetector::getTopDetectedObjects(int top_n) const {
    // Convert map to vector for sorting
    std::vector<std::pair<std::string, int>> sorted_objects(object_type_counts_.begin(), object_type_counts_.end());
    
    // Sort by count in descending order
    std::sort(sorted_objects.begin(), sorted_objects.end(),
              [](const auto& a, const auto& b) {
                  return a.second > b.second;
              });
    
    // Return top N objects
    if (sorted_objects.size() > static_cast<size_t>(top_n)) {
        sorted_objects.resize(top_n);
    }
    
    return sorted_objects;
}

void ObjectDetector::cleanupOldTrackedObjects() {
    // Remove objects that haven't been seen recently, prioritizing older ones
    // This is called when we hit the MAX_TRACKED_OBJECTS limit
    if (tracked_objects_.empty()) {
        return;
    }
    
    // Sort by frames_since_detection (descending) to remove oldest first
    std::sort(tracked_objects_.begin(), tracked_objects_.end(),
              [](const ObjectTracker& a, const ObjectTracker& b) {
                  return a.frames_since_detection > b.frames_since_detection;
              });
    
    // Remove the oldest 20% or at least 10 objects
    size_t to_remove = std::max(static_cast<size_t>(10), tracked_objects_.size() / 5);
    to_remove = std::min(to_remove, tracked_objects_.size());
    
    logger_->debug("Cleaning up " + std::to_string(to_remove) + " old tracked objects");
    tracked_objects_.erase(tracked_objects_.begin(), tracked_objects_.begin() + to_remove);
}

void ObjectDetector::limitObjectTypeCounts() {
    // Keep only the most frequently detected object types
    // This prevents the map from growing unbounded with rare detections
    if (object_type_counts_.size() <= MAX_OBJECT_TYPE_ENTRIES) {
        return;
    }
    
    // Convert to vector and sort by count
    std::vector<std::pair<std::string, int>> sorted_counts(object_type_counts_.begin(), 
                                                            object_type_counts_.end());
    std::sort(sorted_counts.begin(), sorted_counts.end(),
              [](const std::pair<std::string, int>& a, const std::pair<std::string, int>& b) {
                  return a.second > b.second;
              });
    
    // Keep only top MAX_OBJECT_TYPE_ENTRIES
    object_type_counts_.clear();
    for (size_t i = 0; i < MAX_OBJECT_TYPE_ENTRIES && i < sorted_counts.size(); ++i) {
        object_type_counts_[sorted_counts[i].first] = sorted_counts[i].second;
    }
    
    logger_->debug("Limited object type counts to top " + std::to_string(MAX_OBJECT_TYPE_ENTRIES) + " types");
}

void ObjectDetector::updateStationaryStatus(ObjectTracker& tracker) {
    // Need at least 3 positions to determine if stationary
    if (tracker.position_history.size() < 3) {
        tracker.is_stationary = false;
        tracker.stationary_since = std::chrono::steady_clock::now();
        return;
    }
    
    // Calculate average movement over recent history
    float total_distance = 0.0f;
    for (size_t i = 1; i < tracker.position_history.size(); ++i) {
        total_distance += cv::norm(tracker.position_history[i] - tracker.position_history[i-1]);
    }
    float avg_distance = total_distance / (tracker.position_history.size() - 1);
    
    // Check if object is stationary (avg movement below threshold)
    bool currently_stationary = avg_distance <= ObjectTracker::STATIONARY_MOVEMENT_THRESHOLD;
    
    if (currently_stationary && !tracker.is_stationary) {
        // Object just became stationary
        tracker.is_stationary = true;
        tracker.stationary_since = std::chrono::steady_clock::now();
        logger_->debug("Object " + tracker.object_type + " is now stationary (avg movement: " + 
                      std::to_string(avg_distance) + " pixels)");
    } else if (!currently_stationary && tracker.is_stationary) {
        // Object started moving again
        tracker.is_stationary = false;
        logger_->debug("Object " + tracker.object_type + " started moving again (avg movement: " + 
                      std::to_string(avg_distance) + " pixels)");
    } else if (currently_stationary) {
        // Still stationary
        auto now = std::chrono::steady_clock::now();
        auto stationary_duration = std::chrono::duration_cast<std::chrono::seconds>(now - tracker.stationary_since);
        logger_->debug("Object " + tracker.object_type + " stationary for " + 
                      std::to_string(stationary_duration.count()) + " seconds (avg movement: " + 
                      std::to_string(avg_distance) + " pixels)");
    }
}

bool ObjectDetector::isStationaryPastTimeout(const ObjectTracker& tracker, int stationary_timeout_seconds) const {
    if (!tracker.is_stationary) {
        return false;
    }
    
    auto now = std::chrono::steady_clock::now();
    auto stationary_duration = std::chrono::duration_cast<std::chrono::seconds>(now - tracker.stationary_since);
    
    return stationary_duration.count() >= stationary_timeout_seconds;
}
