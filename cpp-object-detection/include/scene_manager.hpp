#pragma once

#include <string>
#include <vector>
#include <memory>
#include <chrono>
#include <map>
#include <opencv2/opencv.hpp>
#include "logger.hpp"
#include "object_detector.hpp"

// Forward declaration for SQLite
struct sqlite3;

/**
 * Scene representation with stationary objects and their properties
 */
struct SceneObject {
    std::string object_type;
    cv::Point2f position;       // Center position in frame
    double orientation;         // Orientation angle in degrees
    cv::Scalar dominant_color;  // Dominant color (BGR)
    std::vector<std::string> sub_parts;  // Detected sub-parts
    cv::Rect2f bounding_box;    // Bounding box for the object
};

/**
 * Complete scene with all stationary objects and their relationships
 */
struct Scene {
    int id = -1;  // Database ID (-1 means not yet persisted)
    std::chrono::system_clock::time_point created_at;
    std::string description;
    std::vector<SceneObject> objects;
    
    // Spatial relationships between objects (distances and angles)
    std::map<std::pair<int, int>, double> object_distances;  // Distance between object pairs
    std::map<std::pair<int, int>, double> object_angles;     // Angle between object pairs
};

/**
 * Configuration for scene matching
 */
struct SceneMatchConfig {
    double position_tolerance = 0.15;      // Max relative position difference (15% of frame size)
    double object_count_tolerance = 0.2;   // Max difference in object counts (20%)
    double distance_tolerance = 0.2;       // Max difference in inter-object distances (20%)
    double angle_tolerance = 15.0;         // Max angle difference in degrees
    double min_match_score = 0.7;          // Minimum score to consider scenes matching (0-1)
    int min_observation_seconds = 60;      // Minimum observation time before creating scene
};

/**
 * Scene Manager - Analyzes, persists, and matches scenes
 * 
 * This class identifies stationary objects in the camera view, analyzes their
 * properties and spatial relationships, and stores scene information in SQLite.
 * It can recognize when the camera returns to a previously seen scene using
 * fuzzy matching that accounts for slight variations in camera angle and object positions.
 */
class SceneManager {
public:
    /**
     * Constructor
     * @param db_path Path to SQLite database file
     * @param logger Logger instance
     * @param config Matching configuration
     */
    SceneManager(const std::string& db_path, 
                 std::shared_ptr<Logger> logger,
                 const SceneMatchConfig& config = SceneMatchConfig());
    
    ~SceneManager();
    
    /**
     * Initialize the scene manager and create database tables
     */
    bool initialize();
    
    /**
     * Update the current observation with new detections
     * This builds up information about the current scene over time
     * @param detections Current frame detections
     * @param frame Current frame for color analysis
     */
    void updateObservation(const std::vector<ObjectDetector::ObjectTracker>& tracked_objects,
                          const cv::Mat& frame);
    
    /**
     * Check if enough observation time has passed to analyze the scene
     */
    bool isReadyToAnalyzeScene() const;
    
    /**
     * Analyze the current observation and create/match a scene
     * Returns the scene ID (existing or new)
     * Also logs whether this is a new scene or a recognized return to an earlier scene
     */
    int analyzeAndPersistScene();
    
    /**
     * Get all stored scenes from database
     */
    std::vector<Scene> getAllScenes() const;
    
    /**
     * Get a specific scene by ID
     */
    Scene getScene(int scene_id) const;
    
    /**
     * Clear the current observation (e.g., when camera view changes significantly)
     */
    void resetObservation();

private:
    std::string db_path_;
    std::shared_ptr<Logger> logger_;
    SceneMatchConfig config_;
    sqlite3* db_;
    
    // Current observation state
    std::chrono::system_clock::time_point observation_start_;
    std::vector<SceneObject> current_objects_;
    bool observation_active_;
    
    /**
     * Create database tables if they don't exist
     */
    bool createTables();
    
    /**
     * Analyze a stationary object to extract features
     */
    SceneObject analyzeObject(const ObjectDetector::ObjectTracker& tracker, 
                             const cv::Mat& frame);
    
    /**
     * Extract dominant color from a region of interest
     */
    cv::Scalar extractDominantColor(const cv::Mat& frame, const cv::Rect& roi);
    
    /**
     * Estimate orientation of an object
     */
    double estimateOrientation(const cv::Mat& frame, const cv::Rect& roi);
    
    /**
     * Build a scene from current observations
     */
    Scene buildScene();
    
    /**
     * Calculate spatial relationships (distances and angles) between objects
     */
    void calculateSpatialRelationships(Scene& scene);
    
    /**
     * Find the best matching scene from the database
     * Returns the scene ID and match score
     */
    std::pair<int, double> findMatchingScene(const Scene& current_scene);
    
    /**
     * Calculate similarity score between two scenes (0-1, higher is more similar)
     */
    double calculateSceneSimilarity(const Scene& scene1, const Scene& scene2);
    
    /**
     * Persist a new scene to the database
     * Returns the new scene ID
     */
    int persistScene(const Scene& scene);
    
    /**
     * Generate a human-readable description of the scene
     */
    std::string generateSceneDescription(const Scene& scene);
    
    /**
     * Load all scenes from database
     */
    std::vector<Scene> loadScenesFromDatabase() const;
};
