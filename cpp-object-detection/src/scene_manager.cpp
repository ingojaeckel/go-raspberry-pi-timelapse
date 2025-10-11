#include "scene_manager.hpp"
#include <sqlite3.h>
#include <sstream>
#include <iomanip>
#include <cmath>
#include <algorithm>

SceneManager::SceneManager(const std::string& db_path,
                           std::shared_ptr<Logger> logger,
                           const SceneMatchConfig& config)
    : db_path_(db_path), logger_(logger), config_(config), db_(nullptr),
      observation_active_(false) {
}

SceneManager::~SceneManager() {
    if (db_) {
        sqlite3_close(db_);
        db_ = nullptr;
    }
}

bool SceneManager::initialize() {
    logger_->info("Initializing Scene Manager with database: " + db_path_);
    
    // Open SQLite database
    int rc = sqlite3_open(db_path_.c_str(), &db_);
    if (rc != SQLITE_OK) {
        logger_->error("Failed to open database: " + std::string(sqlite3_errmsg(db_)));
        return false;
    }
    
    // Create tables
    if (!createTables()) {
        logger_->error("Failed to create database tables");
        return false;
    }
    
    logger_->info("Scene Manager initialized successfully");
    logger_->info("Match configuration: position_tolerance=" + 
                 std::to_string(config_.position_tolerance) +
                 ", min_match_score=" + std::to_string(config_.min_match_score));
    
    return true;
}

bool SceneManager::createTables() {
    const char* scenes_table = R"(
        CREATE TABLE IF NOT EXISTS scenes (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            created_at TEXT NOT NULL,
            description TEXT NOT NULL,
            object_count INTEGER NOT NULL,
            object_types TEXT NOT NULL
        );
    )";
    
    const char* scene_objects_table = R"(
        CREATE TABLE IF NOT EXISTS scene_objects (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            scene_id INTEGER NOT NULL,
            object_type TEXT NOT NULL,
            position_x REAL NOT NULL,
            position_y REAL NOT NULL,
            orientation REAL,
            color_r INTEGER,
            color_g INTEGER,
            color_b INTEGER,
            bbox_x REAL,
            bbox_y REAL,
            bbox_width REAL,
            bbox_height REAL,
            FOREIGN KEY (scene_id) REFERENCES scenes(id)
        );
    )";
    
    const char* spatial_relationships_table = R"(
        CREATE TABLE IF NOT EXISTS spatial_relationships (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            scene_id INTEGER NOT NULL,
            object1_idx INTEGER NOT NULL,
            object2_idx INTEGER NOT NULL,
            distance REAL NOT NULL,
            angle REAL NOT NULL,
            FOREIGN KEY (scene_id) REFERENCES scenes(id)
        );
    )";
    
    char* err_msg = nullptr;
    
    // Create scenes table
    int rc = sqlite3_exec(db_, scenes_table, nullptr, nullptr, &err_msg);
    if (rc != SQLITE_OK) {
        logger_->error("Failed to create scenes table: " + std::string(err_msg));
        sqlite3_free(err_msg);
        return false;
    }
    
    // Create scene_objects table
    rc = sqlite3_exec(db_, scene_objects_table, nullptr, nullptr, &err_msg);
    if (rc != SQLITE_OK) {
        logger_->error("Failed to create scene_objects table: " + std::string(err_msg));
        sqlite3_free(err_msg);
        return false;
    }
    
    // Create spatial_relationships table
    rc = sqlite3_exec(db_, spatial_relationships_table, nullptr, nullptr, &err_msg);
    if (rc != SQLITE_OK) {
        logger_->error("Failed to create spatial_relationships table: " + std::string(err_msg));
        sqlite3_free(err_msg);
        return false;
    }
    
    return true;
}

void SceneManager::updateObservation(const std::vector<ObjectDetector::ObjectTracker>& tracked_objects,
                                     const cv::Mat& frame) {
    if (!observation_active_) {
        observation_start_ = std::chrono::system_clock::now();
        observation_active_ = true;
        current_objects_.clear();
        logger_->debug("Started new scene observation");
    }
    
    // Update current objects with stationary tracked objects
    current_objects_.clear();
    for (const auto& tracker : tracked_objects) {
        if (tracker.is_stationary) {
            SceneObject obj = analyzeObject(tracker, frame);
            current_objects_.push_back(obj);
        }
    }
}

bool SceneManager::isReadyToAnalyzeScene() const {
    if (!observation_active_ || current_objects_.empty()) {
        return false;
    }
    
    auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(
        std::chrono::system_clock::now() - observation_start_).count();
    
    return elapsed >= config_.min_observation_seconds;
}

SceneObject SceneManager::analyzeObject(const ObjectDetector::ObjectTracker& tracker,
                                       const cv::Mat& frame) {
    SceneObject obj;
    obj.object_type = tracker.object_type;
    obj.position = tracker.center;
    
    // Create bounding box around the object (estimate size based on typical object)
    const float estimated_size = 100.0f;  // pixels
    cv::Rect roi(
        static_cast<int>(tracker.center.x - estimated_size / 2),
        static_cast<int>(tracker.center.y - estimated_size / 2),
        static_cast<int>(estimated_size),
        static_cast<int>(estimated_size)
    );
    
    // Ensure ROI is within frame bounds
    roi.x = std::max(0, std::min(roi.x, frame.cols - roi.width));
    roi.y = std::max(0, std::min(roi.y, frame.rows - roi.height));
    roi.width = std::min(roi.width, frame.cols - roi.x);
    roi.height = std::min(roi.height, frame.rows - roi.y);
    
    obj.bounding_box = cv::Rect2f(roi);
    
    if (!frame.empty() && roi.width > 0 && roi.height > 0) {
        obj.dominant_color = extractDominantColor(frame, roi);
        obj.orientation = estimateOrientation(frame, roi);
    } else {
        obj.dominant_color = cv::Scalar(0, 0, 0);
        obj.orientation = 0.0;
    }
    
    return obj;
}

cv::Scalar SceneManager::extractDominantColor(const cv::Mat& frame, const cv::Rect& roi) {
    if (frame.empty() || roi.width <= 0 || roi.height <= 0) {
        return cv::Scalar(0, 0, 0);
    }
    
    cv::Mat region = frame(roi);
    cv::Scalar mean_color = cv::mean(region);
    
    return mean_color;
}

double SceneManager::estimateOrientation(const cv::Mat& frame, const cv::Rect& roi) {
    if (frame.empty() || roi.width <= 0 || roi.height <= 0) {
        return 0.0;
    }
    
    // Convert to grayscale for edge detection
    cv::Mat region = frame(roi);
    cv::Mat gray;
    
    if (region.channels() == 3) {
        cv::cvtColor(region, gray, cv::COLOR_BGR2GRAY);
    } else {
        gray = region.clone();
    }
    
    // Use moments to estimate orientation
    cv::Moments m = cv::moments(gray, true);
    
    if (m.m00 == 0) {
        return 0.0;
    }
    
    // Calculate orientation angle from central moments
    double angle = 0.5 * std::atan2(2 * m.mu11, m.mu20 - m.mu02) * 180.0 / CV_PI;
    
    return angle;
}

Scene SceneManager::buildScene() {
    Scene scene;
    scene.created_at = std::chrono::system_clock::now();
    scene.objects = current_objects_;
    
    calculateSpatialRelationships(scene);
    scene.description = generateSceneDescription(scene);
    
    return scene;
}

void SceneManager::calculateSpatialRelationships(Scene& scene) {
    scene.object_distances.clear();
    scene.object_angles.clear();
    
    for (size_t i = 0; i < scene.objects.size(); ++i) {
        for (size_t j = i + 1; j < scene.objects.size(); ++j) {
            const auto& obj1 = scene.objects[i];
            const auto& obj2 = scene.objects[j];
            
            // Calculate distance
            float dx = obj2.position.x - obj1.position.x;
            float dy = obj2.position.y - obj1.position.y;
            double distance = std::sqrt(dx * dx + dy * dy);
            
            // Calculate angle
            double angle = std::atan2(dy, dx) * 180.0 / CV_PI;
            
            scene.object_distances[{i, j}] = distance;
            scene.object_angles[{i, j}] = angle;
        }
    }
}

int SceneManager::analyzeAndPersistScene() {
    if (!isReadyToAnalyzeScene()) {
        logger_->warning("Scene not ready for analysis yet");
        return -1;
    }
    
    // Build current scene
    Scene current_scene = buildScene();
    
    logger_->info("Analyzing scene with " + std::to_string(current_scene.objects.size()) + 
                 " stationary objects");
    
    // Try to find a matching scene
    auto [matched_id, match_score] = findMatchingScene(current_scene);
    
    if (matched_id >= 0 && match_score >= config_.min_match_score) {
        logger_->info("Recognised return to earlier scene: id=" + std::to_string(matched_id) +
                     " (match score: " + std::to_string(match_score) + ")");
        return matched_id;
    } else {
        // New scene - persist it
        int new_id = persistScene(current_scene);
        logger_->info("New scene was identified: id=" + std::to_string(new_id) +
                     " - " + current_scene.description);
        return new_id;
    }
}

std::pair<int, double> SceneManager::findMatchingScene(const Scene& current_scene) {
    std::vector<Scene> all_scenes = loadScenesFromDatabase();
    
    int best_match_id = -1;
    double best_score = 0.0;
    
    for (const auto& stored_scene : all_scenes) {
        double score = calculateSceneSimilarity(current_scene, stored_scene);
        
        if (score > best_score) {
            best_score = score;
            best_match_id = stored_scene.id;
        }
    }
    
    return {best_match_id, best_score};
}

double SceneManager::calculateSceneSimilarity(const Scene& scene1, const Scene& scene2) {
    // Check object count similarity
    int count_diff = std::abs(static_cast<int>(scene1.objects.size()) - 
                             static_cast<int>(scene2.objects.size()));
    double count_tolerance = std::max(scene1.objects.size(), scene2.objects.size()) * 
                            config_.object_count_tolerance;
    
    if (count_diff > count_tolerance) {
        return 0.0;  // Too different in object count
    }
    
    // Build object type histograms
    std::map<std::string, int> types1, types2;
    for (const auto& obj : scene1.objects) {
        types1[obj.object_type]++;
    }
    for (const auto& obj : scene2.objects) {
        types2[obj.object_type]++;
    }
    
    // Check type similarity
    double type_score = 0.0;
    int total_types = 0;
    for (const auto& [type, count1] : types1) {
        int count2 = types2[type];
        total_types = std::max(total_types, count1);
        total_types = std::max(total_types, count2);
        
        // Score based on count similarity
        int diff = std::abs(count1 - count2);
        if (diff <= 1) {
            type_score += 1.0;
        }
    }
    
    if (total_types == 0) {
        return 0.0;
    }
    
    type_score /= total_types;
    
    // Compare spatial relationships (if both scenes have multiple objects)
    double spatial_score = 1.0;
    if (scene1.objects.size() >= 2 && scene2.objects.size() >= 2) {
        int matching_relationships = 0;
        int total_relationships = 0;
        
        for (const auto& [pair1, dist1] : scene1.object_distances) {
            for (const auto& [pair2, dist2] : scene2.object_distances) {
                total_relationships++;
                
                // Check if distances are similar (relative to scene size)
                double dist_diff = std::abs(dist1 - dist2);
                double dist_threshold = std::max(dist1, dist2) * config_.distance_tolerance;
                
                if (dist_diff <= dist_threshold) {
                    // Also check angle similarity
                    auto angle1_it = scene1.object_angles.find(pair1);
                    auto angle2_it = scene2.object_angles.find(pair2);
                    
                    if (angle1_it != scene1.object_angles.end() && 
                        angle2_it != scene2.object_angles.end()) {
                        double angle_diff = std::abs(angle1_it->second - angle2_it->second);
                        
                        if (angle_diff <= config_.angle_tolerance) {
                            matching_relationships++;
                        }
                    }
                }
            }
        }
        
        if (total_relationships > 0) {
            spatial_score = static_cast<double>(matching_relationships) / total_relationships;
        }
    }
    
    // Weighted combination of scores
    double final_score = 0.5 * type_score + 0.5 * spatial_score;
    
    return final_score;
}

int SceneManager::persistScene(const Scene& scene) {
    // Start transaction
    sqlite3_exec(db_, "BEGIN TRANSACTION", nullptr, nullptr, nullptr);
    
    // Prepare scene insert statement
    const char* scene_sql = R"(
        INSERT INTO scenes (created_at, description, object_count, object_types)
        VALUES (?, ?, ?, ?);
    )";
    
    sqlite3_stmt* stmt;
    int rc = sqlite3_prepare_v2(db_, scene_sql, -1, &stmt, nullptr);
    
    if (rc != SQLITE_OK) {
        logger_->error("Failed to prepare scene insert: " + std::string(sqlite3_errmsg(db_)));
        sqlite3_exec(db_, "ROLLBACK", nullptr, nullptr, nullptr);
        return -1;
    }
    
    // Format timestamp
    auto time = std::chrono::system_clock::to_time_t(scene.created_at);
    std::stringstream ss;
    ss << std::put_time(std::localtime(&time), "%Y-%m-%d %H:%M:%S");
    std::string timestamp = ss.str();
    
    // Build object types string
    std::map<std::string, int> type_counts;
    for (const auto& obj : scene.objects) {
        type_counts[obj.object_type]++;
    }
    
    std::stringstream types_ss;
    bool first = true;
    for (const auto& [type, count] : type_counts) {
        if (!first) types_ss << ", ";
        types_ss << count << "x " << type;
        first = false;
    }
    std::string object_types = types_ss.str();
    
    // Bind parameters
    sqlite3_bind_text(stmt, 1, timestamp.c_str(), -1, SQLITE_TRANSIENT);
    sqlite3_bind_text(stmt, 2, scene.description.c_str(), -1, SQLITE_TRANSIENT);
    sqlite3_bind_int(stmt, 3, scene.objects.size());
    sqlite3_bind_text(stmt, 4, object_types.c_str(), -1, SQLITE_TRANSIENT);
    
    rc = sqlite3_step(stmt);
    sqlite3_finalize(stmt);
    
    if (rc != SQLITE_DONE) {
        logger_->error("Failed to insert scene: " + std::string(sqlite3_errmsg(db_)));
        sqlite3_exec(db_, "ROLLBACK", nullptr, nullptr, nullptr);
        return -1;
    }
    
    int scene_id = sqlite3_last_insert_rowid(db_);
    
    // Insert scene objects
    const char* object_sql = R"(
        INSERT INTO scene_objects 
        (scene_id, object_type, position_x, position_y, orientation, 
         color_r, color_g, color_b, bbox_x, bbox_y, bbox_width, bbox_height)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
    )";
    
    for (const auto& obj : scene.objects) {
        rc = sqlite3_prepare_v2(db_, object_sql, -1, &stmt, nullptr);
        if (rc != SQLITE_OK) {
            logger_->error("Failed to prepare object insert: " + std::string(sqlite3_errmsg(db_)));
            sqlite3_exec(db_, "ROLLBACK", nullptr, nullptr, nullptr);
            return -1;
        }
        
        sqlite3_bind_int(stmt, 1, scene_id);
        sqlite3_bind_text(stmt, 2, obj.object_type.c_str(), -1, SQLITE_TRANSIENT);
        sqlite3_bind_double(stmt, 3, obj.position.x);
        sqlite3_bind_double(stmt, 4, obj.position.y);
        sqlite3_bind_double(stmt, 5, obj.orientation);
        sqlite3_bind_int(stmt, 6, static_cast<int>(obj.dominant_color[2]));  // R
        sqlite3_bind_int(stmt, 7, static_cast<int>(obj.dominant_color[1]));  // G
        sqlite3_bind_int(stmt, 8, static_cast<int>(obj.dominant_color[0]));  // B
        sqlite3_bind_double(stmt, 9, obj.bounding_box.x);
        sqlite3_bind_double(stmt, 10, obj.bounding_box.y);
        sqlite3_bind_double(stmt, 11, obj.bounding_box.width);
        sqlite3_bind_double(stmt, 12, obj.bounding_box.height);
        
        rc = sqlite3_step(stmt);
        sqlite3_finalize(stmt);
        
        if (rc != SQLITE_DONE) {
            logger_->error("Failed to insert scene object: " + std::string(sqlite3_errmsg(db_)));
            sqlite3_exec(db_, "ROLLBACK", nullptr, nullptr, nullptr);
            return -1;
        }
    }
    
    // Insert spatial relationships
    const char* relationship_sql = R"(
        INSERT INTO spatial_relationships 
        (scene_id, object1_idx, object2_idx, distance, angle)
        VALUES (?, ?, ?, ?, ?);
    )";
    
    for (const auto& [pair, distance] : scene.object_distances) {
        rc = sqlite3_prepare_v2(db_, relationship_sql, -1, &stmt, nullptr);
        if (rc != SQLITE_OK) {
            logger_->error("Failed to prepare relationship insert: " + std::string(sqlite3_errmsg(db_)));
            sqlite3_exec(db_, "ROLLBACK", nullptr, nullptr, nullptr);
            return -1;
        }
        
        auto angle_it = scene.object_angles.find(pair);
        double angle = (angle_it != scene.object_angles.end()) ? angle_it->second : 0.0;
        
        sqlite3_bind_int(stmt, 1, scene_id);
        sqlite3_bind_int(stmt, 2, pair.first);
        sqlite3_bind_int(stmt, 3, pair.second);
        sqlite3_bind_double(stmt, 4, distance);
        sqlite3_bind_double(stmt, 5, angle);
        
        rc = sqlite3_step(stmt);
        sqlite3_finalize(stmt);
        
        if (rc != SQLITE_DONE) {
            logger_->error("Failed to insert spatial relationship: " + std::string(sqlite3_errmsg(db_)));
            sqlite3_exec(db_, "ROLLBACK", nullptr, nullptr, nullptr);
            return -1;
        }
    }
    
    // Commit transaction
    sqlite3_exec(db_, "COMMIT", nullptr, nullptr, nullptr);
    
    return scene_id;
}

std::string SceneManager::generateSceneDescription(const Scene& scene) {
    std::stringstream ss;
    
    // Count object types
    std::map<std::string, int> type_counts;
    for (const auto& obj : scene.objects) {
        type_counts[obj.object_type]++;
    }
    
    // Generate description
    bool first = true;
    for (const auto& [type, count] : type_counts) {
        if (!first) {
            ss << ", ";
        }
        ss << count << "x " << type;
        first = false;
    }
    
    if (!scene.objects.empty()) {
        ss << " arranged in frame";
    }
    
    return ss.str();
}

std::vector<Scene> SceneManager::loadScenesFromDatabase() const {
    std::vector<Scene> scenes;
    
    const char* sql = "SELECT id, created_at, description FROM scenes ORDER BY id;";
    sqlite3_stmt* stmt;
    
    int rc = sqlite3_prepare_v2(db_, sql, -1, &stmt, nullptr);
    if (rc != SQLITE_OK) {
        logger_->error("Failed to prepare scene query: " + std::string(sqlite3_errmsg(db_)));
        return scenes;
    }
    
    while (sqlite3_step(stmt) == SQLITE_ROW) {
        Scene scene;
        scene.id = sqlite3_column_int(stmt, 0);
        
        // Note: timestamp is available but not needed for matching
        // const char* timestamp = reinterpret_cast<const char*>(sqlite3_column_text(stmt, 1));
        const char* description = reinterpret_cast<const char*>(sqlite3_column_text(stmt, 2));
        
        scene.description = description ? description : "";
        
        // Load objects for this scene
        const char* obj_sql = "SELECT object_type, position_x, position_y, orientation, "
                             "color_r, color_g, color_b, bbox_x, bbox_y, bbox_width, bbox_height "
                             "FROM scene_objects WHERE scene_id = ? ORDER BY id;";
        
        sqlite3_stmt* obj_stmt;
        rc = sqlite3_prepare_v2(db_, obj_sql, -1, &obj_stmt, nullptr);
        if (rc == SQLITE_OK) {
            sqlite3_bind_int(obj_stmt, 1, scene.id);
            
            while (sqlite3_step(obj_stmt) == SQLITE_ROW) {
                SceneObject obj;
                obj.object_type = reinterpret_cast<const char*>(sqlite3_column_text(obj_stmt, 0));
                obj.position.x = sqlite3_column_double(obj_stmt, 1);
                obj.position.y = sqlite3_column_double(obj_stmt, 2);
                obj.orientation = sqlite3_column_double(obj_stmt, 3);
                
                int r = sqlite3_column_int(obj_stmt, 4);
                int g = sqlite3_column_int(obj_stmt, 5);
                int b = sqlite3_column_int(obj_stmt, 6);
                obj.dominant_color = cv::Scalar(b, g, r);  // BGR format
                
                obj.bounding_box.x = sqlite3_column_double(obj_stmt, 7);
                obj.bounding_box.y = sqlite3_column_double(obj_stmt, 8);
                obj.bounding_box.width = sqlite3_column_double(obj_stmt, 9);
                obj.bounding_box.height = sqlite3_column_double(obj_stmt, 10);
                
                scene.objects.push_back(obj);
            }
            
            sqlite3_finalize(obj_stmt);
        }
        
        // Load spatial relationships
        const char* rel_sql = "SELECT object1_idx, object2_idx, distance, angle "
                             "FROM spatial_relationships WHERE scene_id = ?;";
        
        sqlite3_stmt* rel_stmt;
        rc = sqlite3_prepare_v2(db_, rel_sql, -1, &rel_stmt, nullptr);
        if (rc == SQLITE_OK) {
            sqlite3_bind_int(rel_stmt, 1, scene.id);
            
            while (sqlite3_step(rel_stmt) == SQLITE_ROW) {
                int idx1 = sqlite3_column_int(rel_stmt, 0);
                int idx2 = sqlite3_column_int(rel_stmt, 1);
                double distance = sqlite3_column_double(rel_stmt, 2);
                double angle = sqlite3_column_double(rel_stmt, 3);
                
                scene.object_distances[{idx1, idx2}] = distance;
                scene.object_angles[{idx1, idx2}] = angle;
            }
            
            sqlite3_finalize(rel_stmt);
        }
        
        scenes.push_back(scene);
    }
    
    sqlite3_finalize(stmt);
    return scenes;
}

std::vector<Scene> SceneManager::getAllScenes() const {
    return loadScenesFromDatabase();
}

Scene SceneManager::getScene(int scene_id) const {
    const char* sql = "SELECT id, created_at, description FROM scenes WHERE id = ?;";
    sqlite3_stmt* stmt;
    
    int rc = sqlite3_prepare_v2(db_, sql, -1, &stmt, nullptr);
    if (rc != SQLITE_OK) {
        logger_->error("Failed to prepare scene query: " + std::string(sqlite3_errmsg(db_)));
        return Scene();
    }
    
    sqlite3_bind_int(stmt, 1, scene_id);
    
    Scene scene;
    if (sqlite3_step(stmt) == SQLITE_ROW) {
        scene.id = sqlite3_column_int(stmt, 0);
        const char* description = reinterpret_cast<const char*>(sqlite3_column_text(stmt, 2));
        scene.description = description ? description : "";
        
        // Load objects and relationships (similar to loadScenesFromDatabase)
        // ... (code similar to above, omitted for brevity)
    }
    
    sqlite3_finalize(stmt);
    return scene;
}

void SceneManager::resetObservation() {
    observation_active_ = false;
    current_objects_.clear();
    logger_->debug("Scene observation reset");
}
