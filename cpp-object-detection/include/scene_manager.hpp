#pragma once

#include <string>
#include <vector>
#include <memory>
#include <chrono>
#include <opencv2/opencv.hpp>
#include "object_detector.hpp"
#include "logger.hpp"

struct sqlite3;

/**
 * Represents a stationary object in a scene with detailed properties
 */
struct SceneObject {
    std::string object_type;
    cv::Point2f position;        // Position in frame
    cv::Scalar color;            // Average color of object
    cv::Size size;               // Bounding box size
    float orientation;           // Orientation angle if detectable
};

/**
 * Represents spatial relationships between objects
 */
struct ObjectRelationship {
    int object1_idx;
    int object2_idx;
    float distance;              // Pixel distance between centers
    float angle;                 // Angle between objects (radians)
};

/**
 * Scene fingerprint for fuzzy matching
 */
struct SceneFingerprint {
    std::map<std::string, int> object_counts;  // Count of each object type
    std::vector<ObjectRelationship> relationships;  // Spatial relationships
    cv::Mat spatial_histogram;   // Spatial distribution of objects
};

/**
 * Represents a complete scene
 */
struct Scene {
    int id;
    std::chrono::system_clock::time_point created_at;
    std::string description;
    std::vector<SceneObject> objects;
    SceneFingerprint fingerprint;
};

/**
 * Manages scene recognition, persistence, and fuzzy matching
 */
class SceneManager {
public:
    SceneManager(std::shared_ptr<Logger> logger, const std::string& db_path = "scenes.db");
    ~SceneManager();

    /**
     * Initialize the scene database
     */
    bool initialize();

    /**
     * Analyze current frame and stationary objects to identify scene
     * Returns scene ID if matched, -1 if new scene
     */
    int analyzeAndMatchScene(const cv::Mat& frame, 
                            const std::vector<ObjectDetector::ObjectTracker>& tracked_objects);

    /**
     * Check if enough time has passed with stationary objects to trigger scene analysis
     */
    bool shouldAnalyzeScene(const std::vector<ObjectDetector::ObjectTracker>& tracked_objects) const;

    /**
     * Get description of a scene by ID
     */
    std::string getSceneDescription(int scene_id) const;

private:
    std::shared_ptr<Logger> logger_;
    std::string db_path_;
    sqlite3* db_;
    bool initialized_;
    
    // Timing for scene analysis
    std::chrono::steady_clock::time_point last_analysis_time_;
    static constexpr int SCENE_ANALYSIS_INTERVAL_SECONDS = 60;  // Analyze after 1 minute
    static constexpr float MATCH_THRESHOLD = 0.75f;  // Similarity threshold for matching

    /**
     * Create database schema
     */
    bool createSchema();

    /**
     * Extract detailed properties from stationary objects
     */
    std::vector<SceneObject> analyzeStationaryObjects(
        const cv::Mat& frame,
        const std::vector<ObjectDetector::ObjectTracker>& tracked_objects);

    /**
     * Calculate spatial relationships between objects
     */
    std::vector<ObjectRelationship> calculateRelationships(
        const std::vector<SceneObject>& objects);

    /**
     * Create scene fingerprint for matching
     */
    SceneFingerprint createFingerprint(const std::vector<SceneObject>& objects);

    /**
     * Find matching scene in database (fuzzy matching)
     * Returns scene ID if match found, -1 otherwise
     */
    int findMatchingScene(const SceneFingerprint& fingerprint);

    /**
     * Save new scene to database
     */
    int saveScene(const Scene& scene);

    /**
     * Calculate similarity between two fingerprints (0.0 - 1.0)
     */
    float calculateSimilarity(const SceneFingerprint& fp1, const SceneFingerprint& fp2);

    /**
     * Generate human-readable description of scene
     */
    std::string generateDescription(const std::vector<SceneObject>& objects);

    /**
     * Extract color from region of interest
     */
    cv::Scalar extractColor(const cv::Mat& frame, const cv::Rect& roi);

    /**
     * Load scene fingerprint from database
     */
    SceneFingerprint loadFingerprint(int scene_id);
};
