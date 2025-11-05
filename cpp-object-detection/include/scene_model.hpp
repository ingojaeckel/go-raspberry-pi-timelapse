#pragma once

#include <opencv2/opencv.hpp>
#include <vector>
#include <string>
#include <map>
#include <memory>
#include <chrono>

/**
 * Scene Model - Represents a 3D scene with stationary and dynamic objects
 * 
 * This model tracks:
 * - Stationary objects (trees, hedges, bushes, houses, benches, etc.)
 * - Dynamic objects (people, vehicles, animals)
 * - Relative positions between objects
 * - Scene composition and spatial relationships
 */

/**
 * Represents a detected object in the scene with 3D spatial information
 */
struct SceneObject {
    std::string object_type;           // Type of object (e.g., "tree", "house", "person")
    cv::Point2f position;              // 2D position in frame (center point)
    cv::Rect2f bounding_box;           // Bounding box in frame
    float estimated_depth;             // Estimated depth/distance (0.0-1.0, based on size heuristics)
    bool is_stationary;                // Whether object is stationary
    std::chrono::steady_clock::time_point first_detected;  // When object was first detected
    std::chrono::steady_clock::time_point last_seen;       // Last time object was detected
    int detection_count;               // Number of times detected
    float average_confidence;          // Average detection confidence
    
    SceneObject() : estimated_depth(0.0f), is_stationary(false), 
                   detection_count(0), average_confidence(0.0f) {}
};

/**
 * Represents the spatial relationship between two objects in the scene
 */
struct SpatialRelationship {
    std::string object1_type;          // Type of first object
    std::string object2_type;          // Type of second object
    cv::Point2f object1_position;      // Position of first object
    cv::Point2f object2_position;      // Position of second object
    float distance;                    // 2D Euclidean distance between objects
    float angle;                       // Angle from object1 to object2 (radians)
    std::string relative_position;     // Textual description (e.g., "left of", "above", "behind")
    
    SpatialRelationship() : distance(0.0f), angle(0.0f) {}
};

/**
 * Scene Model - Comprehensive representation of the monitored environment
 */
class SceneModel {
public:
    SceneModel();
    ~SceneModel() = default;
    
    /**
     * Add or update an object in the scene
     */
    void addOrUpdateObject(const std::string& object_type, 
                          const cv::Point2f& position,
                          const cv::Rect2f& bounding_box,
                          float confidence,
                          bool is_stationary);
    
    /**
     * Remove an object from the scene (when no longer detected)
     */
    void removeObject(const std::string& object_type, const cv::Point2f& position);
    
    /**
     * Get all objects currently in the scene
     */
    const std::vector<SceneObject>& getObjects() const { return objects_; }
    
    /**
     * Get all stationary objects in the scene
     */
    std::vector<SceneObject> getStationaryObjects() const;
    
    /**
     * Get all dynamic (moving) objects in the scene
     */
    std::vector<SceneObject> getDynamicObjects() const;
    
    /**
     * Get objects of a specific type
     */
    std::vector<SceneObject> getObjectsByType(const std::string& type) const;
    
    /**
     * Calculate and get spatial relationships between objects in the scene
     */
    std::vector<SpatialRelationship> getSpatialRelationships() const;
    
    /**
     * Get spatial relationships involving a specific object type
     */
    std::vector<SpatialRelationship> getRelationshipsForObject(const std::string& object_type) const;
    
    /**
     * Find nearest object of a given type to a position
     */
    SceneObject* findNearestObject(const std::string& object_type, const cv::Point2f& position);
    
    /**
     * Get scene statistics
     */
    struct SceneStats {
        int total_objects;
        int stationary_objects;
        int dynamic_objects;
        std::map<std::string, int> objects_by_type;
    };
    SceneStats getSceneStats() const;
    
    /**
     * Clear all objects from the scene
     */
    void clear();
    
    /**
     * Clean up objects that haven't been seen recently
     */
    void cleanupOldObjects(int max_age_seconds = 300);
    
    /**
     * Get a textual description of the scene composition
     */
    std::string getSceneDescription() const;
    
    /**
     * Check if an object type is considered stationary by default
     */
    static bool isStationaryObjectType(const std::string& object_type);
    
    /**
     * Get list of stationary object types to detect
     */
    static std::vector<std::string> getStationaryObjectTypes();

private:
    std::vector<SceneObject> objects_;
    
    // Constants for spatial relationship calculations
    static constexpr float SAME_OBJECT_DISTANCE_THRESHOLD = 50.0f;  // pixels
    static constexpr int MAX_SCENE_OBJECTS = 200;  // Prevent unbounded growth
    
    // Helper methods
    float calculateDistance(const cv::Point2f& p1, const cv::Point2f& p2) const;
    float calculateAngle(const cv::Point2f& from, const cv::Point2f& to) const;
    std::string describeRelativePosition(float angle, float distance) const;
    float estimateDepth(const cv::Rect2f& bbox) const;
    SceneObject* findExistingObject(const std::string& type, const cv::Point2f& position);
};
