#include "scene_model.hpp"
#include <algorithm>
#include <cmath>
#include <sstream>
#include <iomanip>

SceneModel::SceneModel() {
}

void SceneModel::addOrUpdateObject(const std::string& object_type,
                                   const cv::Point2f& position,
                                   const cv::Rect2f& bounding_box,
                                   float confidence,
                                   bool is_stationary) {
    // Find if object already exists at similar position
    auto* existing = findExistingObject(object_type, position);
    
    if (existing != nullptr) {
        // Update existing object
        existing->position = position;
        existing->bounding_box = bounding_box;
        existing->is_stationary = is_stationary;
        existing->last_seen = std::chrono::steady_clock::now();
        existing->detection_count++;
        
        // Update rolling average confidence
        existing->average_confidence = 
            (existing->average_confidence * (existing->detection_count - 1) + confidence) / 
            existing->detection_count;
        
        existing->estimated_depth = estimateDepth(bounding_box);
    } else {
        // Add new object
        if (objects_.size() >= MAX_SCENE_OBJECTS) {
            // Remove oldest object to prevent unbounded growth
            cleanupOldObjects(60);  // Remove objects older than 60 seconds
        }
        
        SceneObject new_object;
        new_object.object_type = object_type;
        new_object.position = position;
        new_object.bounding_box = bounding_box;
        new_object.estimated_depth = estimateDepth(bounding_box);
        new_object.is_stationary = is_stationary || isStationaryObjectType(object_type);
        new_object.first_detected = std::chrono::steady_clock::now();
        new_object.last_seen = new_object.first_detected;
        new_object.detection_count = 1;
        new_object.average_confidence = confidence;
        
        objects_.push_back(new_object);
    }
}

void SceneModel::removeObject(const std::string& object_type, const cv::Point2f& position) {
    objects_.erase(
        std::remove_if(objects_.begin(), objects_.end(),
                      [&](const SceneObject& obj) {
                          return obj.object_type == object_type &&
                                 calculateDistance(obj.position, position) < SAME_OBJECT_DISTANCE_THRESHOLD;
                      }),
        objects_.end());
}

std::vector<SceneObject> SceneModel::getStationaryObjects() const {
    std::vector<SceneObject> result;
    std::copy_if(objects_.begin(), objects_.end(), std::back_inserter(result),
                [](const SceneObject& obj) { return obj.is_stationary; });
    return result;
}

std::vector<SceneObject> SceneModel::getDynamicObjects() const {
    std::vector<SceneObject> result;
    std::copy_if(objects_.begin(), objects_.end(), std::back_inserter(result),
                [](const SceneObject& obj) { return !obj.is_stationary; });
    return result;
}

std::vector<SceneObject> SceneModel::getObjectsByType(const std::string& type) const {
    std::vector<SceneObject> result;
    std::copy_if(objects_.begin(), objects_.end(), std::back_inserter(result),
                [&type](const SceneObject& obj) { return obj.object_type == type; });
    return result;
}

std::vector<SpatialRelationship> SceneModel::getSpatialRelationships() const {
    std::vector<SpatialRelationship> relationships;
    
    // Calculate relationships between all pairs of objects
    for (size_t i = 0; i < objects_.size(); ++i) {
        for (size_t j = i + 1; j < objects_.size(); ++j) {
            const auto& obj1 = objects_[i];
            const auto& obj2 = objects_[j];
            
            SpatialRelationship rel;
            rel.object1_type = obj1.object_type;
            rel.object2_type = obj2.object_type;
            rel.object1_position = obj1.position;
            rel.object2_position = obj2.position;
            rel.distance = calculateDistance(obj1.position, obj2.position);
            rel.angle = calculateAngle(obj1.position, obj2.position);
            rel.relative_position = describeRelativePosition(rel.angle, rel.distance);
            
            relationships.push_back(rel);
        }
    }
    
    return relationships;
}

std::vector<SpatialRelationship> SceneModel::getRelationshipsForObject(const std::string& object_type) const {
    auto all_relationships = getSpatialRelationships();
    std::vector<SpatialRelationship> result;
    
    std::copy_if(all_relationships.begin(), all_relationships.end(), std::back_inserter(result),
                [&object_type](const SpatialRelationship& rel) {
                    return rel.object1_type == object_type || rel.object2_type == object_type;
                });
    
    return result;
}

SceneObject* SceneModel::findNearestObject(const std::string& object_type, const cv::Point2f& position) {
    SceneObject* nearest = nullptr;
    float min_distance = std::numeric_limits<float>::max();
    
    for (auto& obj : objects_) {
        if (obj.object_type == object_type) {
            float dist = calculateDistance(obj.position, position);
            if (dist < min_distance) {
                min_distance = dist;
                nearest = &obj;
            }
        }
    }
    
    return nearest;
}

SceneModel::SceneStats SceneModel::getSceneStats() const {
    SceneStats stats;
    stats.total_objects = objects_.size();
    stats.stationary_objects = 0;
    stats.dynamic_objects = 0;
    
    for (const auto& obj : objects_) {
        if (obj.is_stationary) {
            stats.stationary_objects++;
        } else {
            stats.dynamic_objects++;
        }
        stats.objects_by_type[obj.object_type]++;
    }
    
    return stats;
}

void SceneModel::clear() {
    objects_.clear();
}

void SceneModel::cleanupOldObjects(int max_age_seconds) {
    auto now = std::chrono::steady_clock::now();
    
    objects_.erase(
        std::remove_if(objects_.begin(), objects_.end(),
                      [&](const SceneObject& obj) {
                          auto age = std::chrono::duration_cast<std::chrono::seconds>(
                              now - obj.last_seen);
                          return age.count() >= max_age_seconds;
                      }),
        objects_.end());
}

std::string SceneModel::getSceneDescription() const {
    std::ostringstream desc;
    auto stats = getSceneStats();
    
    desc << "Scene contains " << stats.total_objects << " objects";
    if (stats.total_objects > 0) {
        desc << " (" << stats.stationary_objects << " stationary, " 
             << stats.dynamic_objects << " dynamic)";
    }
    
    if (!stats.objects_by_type.empty()) {
        desc << ": ";
        bool first = true;
        for (const auto& [type, count] : stats.objects_by_type) {
            if (!first) desc << ", ";
            desc << count << " " << type;
            if (count > 1) desc << "s";
            first = false;
        }
    }
    
    return desc.str();
}

bool SceneModel::isStationaryObjectType(const std::string& object_type) {
    auto stationary_types = getStationaryObjectTypes();
    return std::find(stationary_types.begin(), stationary_types.end(), object_type) 
           != stationary_types.end();
}

std::vector<std::string> SceneModel::getStationaryObjectTypes() {
    return {
        // Natural stationary objects
        "potted plant",   // COCO class for plants/trees in pots
        
        // Structural stationary objects
        "bench",          // Park benches, outdoor seating
        "traffic light",  // Street fixtures
        "fire hydrant",   // Street fixtures
        "stop sign",      // Street signs
        "parking meter",  // Street fixtures
        
        // Furniture (typically stationary outdoors)
        "chair",          // Outdoor chairs
        "couch",          // Outdoor furniture
        "dining table",   // Outdoor tables
        
        // Note: COCO dataset doesn't include "tree", "hedge", "bush", or "house"
        // as separate classes. These would need a custom trained model.
        // Using closest available COCO classes as proxies.
    };
}

float SceneModel::calculateDistance(const cv::Point2f& p1, const cv::Point2f& p2) const {
    return cv::norm(p1 - p2);
}

float SceneModel::calculateAngle(const cv::Point2f& from, const cv::Point2f& to) const {
    float dx = to.x - from.x;
    float dy = to.y - from.y;
    return std::atan2(dy, dx);
}

std::string SceneModel::describeRelativePosition(float angle, float distance) const {
    // Convert angle to degrees for easier understanding
    float degrees = angle * 180.0f / M_PI;
    
    // Normalize to 0-360 range
    if (degrees < 0) degrees += 360.0f;
    
    // Determine relative position based on angle
    std::string direction;
    if (degrees >= 337.5 || degrees < 22.5) {
        direction = "to the right of";
    } else if (degrees >= 22.5 && degrees < 67.5) {
        direction = "below and to the right of";
    } else if (degrees >= 67.5 && degrees < 112.5) {
        direction = "below";
    } else if (degrees >= 112.5 && degrees < 157.5) {
        direction = "below and to the left of";
    } else if (degrees >= 157.5 && degrees < 202.5) {
        direction = "to the left of";
    } else if (degrees >= 202.5 && degrees < 247.5) {
        direction = "above and to the left of";
    } else if (degrees >= 247.5 && degrees < 292.5) {
        direction = "above";
    } else {
        direction = "above and to the right of";
    }
    
    // Add distance context
    std::ostringstream desc;
    desc << direction;
    
    if (distance < 100) {
        desc << " (very close)";
    } else if (distance < 300) {
        desc << " (nearby)";
    } else if (distance < 600) {
        desc << " (moderate distance)";
    } else {
        desc << " (far away)";
    }
    
    return desc.str();
}

float SceneModel::estimateDepth(const cv::Rect2f& bbox) const {
    // Simple depth estimation based on object size
    // Larger objects in frame are typically closer
    // Returns value between 0.0 (far) and 1.0 (near)
    
    float area = bbox.width * bbox.height;
    float frame_area = 1280.0f * 720.0f;  // Assuming 720p default
    float size_ratio = area / frame_area;
    
    // Clamp between 0 and 1
    float depth = std::min(1.0f, std::max(0.0f, size_ratio * 10.0f));
    
    return depth;
}

SceneObject* SceneModel::findExistingObject(const std::string& type, const cv::Point2f& position) {
    for (auto& obj : objects_) {
        if (obj.object_type == type && 
            calculateDistance(obj.position, position) < SAME_OBJECT_DISTANCE_THRESHOLD) {
            return &obj;
        }
    }
    return nullptr;
}
