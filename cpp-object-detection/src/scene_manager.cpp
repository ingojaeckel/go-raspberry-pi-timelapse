#include "scene_manager.hpp"
#include <sqlite3.h>
#include <sstream>
#include <iomanip>
#include <cmath>
#include <algorithm>

SceneManager::SceneManager(std::shared_ptr<Logger> logger, const std::string& db_path)
    : logger_(logger), db_path_(db_path), db_(nullptr), initialized_(false),
      last_analysis_time_(std::chrono::steady_clock::now()) {
}

SceneManager::~SceneManager() {
    if (db_) {
        sqlite3_close(db_);
        db_ = nullptr;
    }
}

bool SceneManager::initialize() {
    if (initialized_) {
        return true;
    }

    logger_->info("Initializing Scene Manager with database: " + db_path_);

    // Open SQLite database
    int rc = sqlite3_open(db_path_.c_str(), &db_);
    if (rc != SQLITE_OK) {
        logger_->error("Failed to open scene database: " + std::string(sqlite3_errmsg(db_)));
        return false;
    }

    // Create schema
    if (!createSchema()) {
        logger_->error("Failed to create database schema");
        sqlite3_close(db_);
        db_ = nullptr;
        return false;
    }

    initialized_ = true;
    logger_->info("Scene Manager initialized successfully");
    return true;
}

bool SceneManager::createSchema() {
    const char* schema = R"(
        CREATE TABLE IF NOT EXISTS scenes (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            created_at INTEGER NOT NULL,
            description TEXT NOT NULL,
            object_counts TEXT NOT NULL,
            spatial_histogram BLOB
        );

        CREATE TABLE IF NOT EXISTS scene_objects (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            scene_id INTEGER NOT NULL,
            object_type TEXT NOT NULL,
            position_x REAL NOT NULL,
            position_y REAL NOT NULL,
            color_b INTEGER,
            color_g INTEGER,
            color_r INTEGER,
            size_width INTEGER,
            size_height INTEGER,
            orientation REAL,
            FOREIGN KEY (scene_id) REFERENCES scenes(id)
        );

        CREATE TABLE IF NOT EXISTS object_relationships (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            scene_id INTEGER NOT NULL,
            object1_idx INTEGER NOT NULL,
            object2_idx INTEGER NOT NULL,
            distance REAL NOT NULL,
            angle REAL NOT NULL,
            FOREIGN KEY (scene_id) REFERENCES scenes(id)
        );

        CREATE INDEX IF NOT EXISTS idx_scenes_created_at ON scenes(created_at);
    )";

    char* err_msg = nullptr;
    int rc = sqlite3_exec(db_, schema, nullptr, nullptr, &err_msg);
    
    if (rc != SQLITE_OK) {
        logger_->error("SQL error: " + std::string(err_msg));
        sqlite3_free(err_msg);
        return false;
    }

    return true;
}

bool SceneManager::shouldAnalyzeScene(
    const std::vector<ObjectDetector::ObjectTracker>& tracked_objects) const {
    
    if (!initialized_) {
        return false;
    }

    // Check if enough time has passed since last analysis
    auto now = std::chrono::steady_clock::now();
    auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(now - last_analysis_time_).count();
    
    if (elapsed < SCENE_ANALYSIS_INTERVAL_SECONDS) {
        return false;
    }

    // Check if we have stationary objects
    bool has_stationary = false;
    for (const auto& tracker : tracked_objects) {
        if (tracker.is_stationary) {
            // Check if object has been stationary for at least the interval
            auto stationary_duration = std::chrono::duration_cast<std::chrono::seconds>(
                now - tracker.stationary_since).count();
            if (stationary_duration >= SCENE_ANALYSIS_INTERVAL_SECONDS) {
                has_stationary = true;
                break;
            }
        }
    }

    return has_stationary;
}

std::vector<SceneObject> SceneManager::analyzeStationaryObjects(
    const cv::Mat& frame,
    const std::vector<ObjectDetector::ObjectTracker>& tracked_objects) {
    
    std::vector<SceneObject> scene_objects;

    for (const auto& tracker : tracked_objects) {
        if (!tracker.is_stationary) {
            continue;
        }

        SceneObject obj;
        obj.object_type = tracker.object_type;
        obj.position = tracker.center;
        
        // Estimate bounding box size (use a default size based on object type)
        // In a real implementation, this would come from the detection
        int default_size = 100;
        if (tracker.object_type == "person") {
            default_size = 150;
        } else if (tracker.object_type == "car" || tracker.object_type == "truck") {
            default_size = 200;
        }
        
        cv::Rect roi(
            std::max(0, static_cast<int>(tracker.center.x - default_size / 2)),
            std::max(0, static_cast<int>(tracker.center.y - default_size / 2)),
            std::min(default_size, frame.cols - static_cast<int>(tracker.center.x - default_size / 2)),
            std::min(default_size, frame.rows - static_cast<int>(tracker.center.y - default_size / 2))
        );

        // Ensure ROI is valid
        if (roi.x >= 0 && roi.y >= 0 && 
            roi.x + roi.width <= frame.cols && 
            roi.y + roi.height <= frame.rows &&
            roi.width > 0 && roi.height > 0) {
            obj.color = extractColor(frame, roi);
            obj.size = cv::Size(roi.width, roi.height);
        } else {
            obj.color = cv::Scalar(0, 0, 0);
            obj.size = cv::Size(default_size, default_size);
        }
        
        obj.orientation = 0.0f;  // Could be enhanced with edge detection
        
        scene_objects.push_back(obj);
    }

    return scene_objects;
}

cv::Scalar SceneManager::extractColor(const cv::Mat& frame, const cv::Rect& roi) {
    if (frame.empty() || roi.area() == 0) {
        return cv::Scalar(0, 0, 0);
    }

    try {
        cv::Mat roi_mat = frame(roi);
        return cv::mean(roi_mat);
    } catch (const cv::Exception& e) {
        logger_->warning("Failed to extract color: " + std::string(e.what()));
        return cv::Scalar(0, 0, 0);
    }
}

std::vector<ObjectRelationship> SceneManager::calculateRelationships(
    const std::vector<SceneObject>& objects) {
    
    std::vector<ObjectRelationship> relationships;

    for (size_t i = 0; i < objects.size(); ++i) {
        for (size_t j = i + 1; j < objects.size(); ++j) {
            ObjectRelationship rel;
            rel.object1_idx = static_cast<int>(i);
            rel.object2_idx = static_cast<int>(j);

            // Calculate distance
            float dx = objects[j].position.x - objects[i].position.x;
            float dy = objects[j].position.y - objects[i].position.y;
            rel.distance = std::sqrt(dx * dx + dy * dy);

            // Calculate angle
            rel.angle = std::atan2(dy, dx);

            relationships.push_back(rel);
        }
    }

    return relationships;
}

SceneFingerprint SceneManager::createFingerprint(const std::vector<SceneObject>& objects) {
    SceneFingerprint fp;

    // Count object types
    for (const auto& obj : objects) {
        fp.object_counts[obj.object_type]++;
    }

    // Calculate relationships
    fp.relationships = calculateRelationships(objects);

    // Create spatial histogram (simple grid-based approach)
    // Divide frame into 4x4 grid and count objects in each cell
    fp.spatial_histogram = cv::Mat::zeros(4, 4, CV_32F);
    
    // Normalize positions to 0-1 range (assuming typical frame dimensions)
    for (const auto& obj : objects) {
        int grid_x = std::min(3, static_cast<int>(obj.position.x / 320));  // Assuming ~1280 width
        int grid_y = std::min(3, static_cast<int>(obj.position.y / 180));  // Assuming ~720 height
        fp.spatial_histogram.at<float>(grid_y, grid_x) += 1.0f;
    }

    return fp;
}

float SceneManager::calculateSimilarity(const SceneFingerprint& fp1, const SceneFingerprint& fp2) {
    float similarity = 0.0f;
    int components = 0;

    // 1. Compare object counts (40% weight)
    float count_similarity = 0.0f;
    std::set<std::string> all_types;
    for (const auto& pair : fp1.object_counts) {
        all_types.insert(pair.first);
    }
    for (const auto& pair : fp2.object_counts) {
        all_types.insert(pair.first);
    }

    if (!all_types.empty()) {
        for (const auto& type : all_types) {
            int count1 = fp1.object_counts.count(type) ? fp1.object_counts.at(type) : 0;
            int count2 = fp2.object_counts.count(type) ? fp2.object_counts.at(type) : 0;
            
            // Use ratio of min to max for similarity
            if (count1 > 0 || count2 > 0) {
                float ratio = static_cast<float>(std::min(count1, count2)) / 
                             static_cast<float>(std::max(count1, count2));
                count_similarity += ratio;
            }
        }
        count_similarity /= all_types.size();
        similarity += count_similarity * 0.4f;
        components++;
    }

    // 2. Compare spatial distribution (40% weight)
    if (!fp1.spatial_histogram.empty() && !fp2.spatial_histogram.empty() &&
        fp1.spatial_histogram.size() == fp2.spatial_histogram.size()) {
        
        // Use correlation coefficient
        double corr = cv::compareHist(fp1.spatial_histogram, fp2.spatial_histogram, cv::HISTCMP_CORREL);
        similarity += static_cast<float>((corr + 1.0) / 2.0) * 0.4f;  // Normalize from [-1,1] to [0,1]
        components++;
    }

    // 3. Compare relationships (20% weight)
    if (!fp1.relationships.empty() && !fp2.relationships.empty()) {
        float rel_similarity = 0.0f;
        int matching_relationships = 0;

        // Compare distances between similar object pairs
        for (const auto& rel1 : fp1.relationships) {
            for (const auto& rel2 : fp2.relationships) {
                // Check if distances are similar (within 20% tolerance)
                float dist_ratio = std::min(rel1.distance, rel2.distance) / 
                                  std::max(rel1.distance, rel2.distance);
                if (dist_ratio > 0.8f) {
                    matching_relationships++;
                    break;
                }
            }
        }

        rel_similarity = static_cast<float>(matching_relationships) / 
                        static_cast<float>(std::max(fp1.relationships.size(), fp2.relationships.size()));
        similarity += rel_similarity * 0.2f;
        components++;
    }

    // Average the components
    return components > 0 ? similarity : 0.0f;
}

int SceneManager::findMatchingScene(const SceneFingerprint& fingerprint) {
    if (!initialized_ || !db_) {
        return -1;
    }

    // Query all scenes
    const char* query = "SELECT id FROM scenes ORDER BY created_at DESC LIMIT 100";
    sqlite3_stmt* stmt;
    
    int rc = sqlite3_prepare_v2(db_, query, -1, &stmt, nullptr);
    if (rc != SQLITE_OK) {
        logger_->error("Failed to prepare query: " + std::string(sqlite3_errmsg(db_)));
        return -1;
    }

    int best_match_id = -1;
    float best_similarity = 0.0f;

    while (sqlite3_step(stmt) == SQLITE_ROW) {
        int scene_id = sqlite3_column_int(stmt, 0);
        
        // Load fingerprint for this scene
        SceneFingerprint stored_fp = loadFingerprint(scene_id);
        
        // Calculate similarity
        float similarity = calculateSimilarity(fingerprint, stored_fp);
        
        if (similarity > best_similarity && similarity >= MATCH_THRESHOLD) {
            best_similarity = similarity;
            best_match_id = scene_id;
        }
    }

    sqlite3_finalize(stmt);
    
    if (best_match_id >= 0) {
        logger_->info("Found matching scene (ID: " + std::to_string(best_match_id) + 
                     ", similarity: " + std::to_string(static_cast<int>(best_similarity * 100)) + "%)");
    }

    return best_match_id;
}

SceneFingerprint SceneManager::loadFingerprint(int scene_id) {
    SceneFingerprint fp;

    if (!initialized_ || !db_) {
        return fp;
    }

    // Load object counts from scenes table
    const char* query1 = "SELECT object_counts FROM scenes WHERE id = ?";
    sqlite3_stmt* stmt1;
    
    if (sqlite3_prepare_v2(db_, query1, -1, &stmt1, nullptr) == SQLITE_OK) {
        sqlite3_bind_int(stmt1, 1, scene_id);
        
        if (sqlite3_step(stmt1) == SQLITE_ROW) {
            const char* counts_str = reinterpret_cast<const char*>(sqlite3_column_text(stmt1, 0));
            if (counts_str) {
                // Parse object counts (format: "type1:count1,type2:count2,...")
                std::stringstream ss(counts_str);
                std::string item;
                while (std::getline(ss, item, ',')) {
                    size_t colon_pos = item.find(':');
                    if (colon_pos != std::string::npos) {
                        std::string type = item.substr(0, colon_pos);
                        int count = std::stoi(item.substr(colon_pos + 1));
                        fp.object_counts[type] = count;
                    }
                }
            }
        }
        sqlite3_finalize(stmt1);
    }

    // Load relationships
    const char* query2 = "SELECT object1_idx, object2_idx, distance, angle FROM object_relationships WHERE scene_id = ?";
    sqlite3_stmt* stmt2;
    
    if (sqlite3_prepare_v2(db_, query2, -1, &stmt2, nullptr) == SQLITE_OK) {
        sqlite3_bind_int(stmt2, 1, scene_id);
        
        while (sqlite3_step(stmt2) == SQLITE_ROW) {
            ObjectRelationship rel;
            rel.object1_idx = sqlite3_column_int(stmt2, 0);
            rel.object2_idx = sqlite3_column_int(stmt2, 1);
            rel.distance = static_cast<float>(sqlite3_column_double(stmt2, 2));
            rel.angle = static_cast<float>(sqlite3_column_double(stmt2, 3));
            fp.relationships.push_back(rel);
        }
        sqlite3_finalize(stmt2);
    }

    // Create spatial histogram from scene objects
    const char* query3 = "SELECT position_x, position_y FROM scene_objects WHERE scene_id = ?";
    sqlite3_stmt* stmt3;
    
    fp.spatial_histogram = cv::Mat::zeros(4, 4, CV_32F);
    
    if (sqlite3_prepare_v2(db_, query3, -1, &stmt3, nullptr) == SQLITE_OK) {
        sqlite3_bind_int(stmt3, 1, scene_id);
        
        while (sqlite3_step(stmt3) == SQLITE_ROW) {
            float x = static_cast<float>(sqlite3_column_double(stmt3, 0));
            float y = static_cast<float>(sqlite3_column_double(stmt3, 1));
            
            int grid_x = std::min(3, static_cast<int>(x / 320));
            int grid_y = std::min(3, static_cast<int>(y / 180));
            fp.spatial_histogram.at<float>(grid_y, grid_x) += 1.0f;
        }
        sqlite3_finalize(stmt3);
    }

    return fp;
}

std::string SceneManager::generateDescription(const std::vector<SceneObject>& objects) {
    std::map<std::string, int> counts;
    for (const auto& obj : objects) {
        counts[obj.object_type]++;
    }

    std::stringstream desc;
    bool first = true;
    for (const auto& pair : counts) {
        if (!first) {
            desc << ", ";
        }
        desc << pair.second << "x " << pair.first;
        if (pair.second > 1) {
            desc << "s";
        }
        first = false;
    }

    return desc.str();
}

int SceneManager::saveScene(const Scene& scene) {
    if (!initialized_ || !db_) {
        return -1;
    }

    // Begin transaction
    sqlite3_exec(db_, "BEGIN TRANSACTION", nullptr, nullptr, nullptr);

    // Serialize object counts
    std::stringstream counts_ss;
    bool first = true;
    for (const auto& pair : scene.fingerprint.object_counts) {
        if (!first) {
            counts_ss << ",";
        }
        counts_ss << pair.first << ":" << pair.second;
        first = false;
    }

    // Insert scene
    const char* insert_scene = "INSERT INTO scenes (created_at, description, object_counts) VALUES (?, ?, ?)";
    sqlite3_stmt* stmt;
    
    int rc = sqlite3_prepare_v2(db_, insert_scene, -1, &stmt, nullptr);
    if (rc != SQLITE_OK) {
        logger_->error("Failed to prepare scene insert: " + std::string(sqlite3_errmsg(db_)));
        sqlite3_exec(db_, "ROLLBACK", nullptr, nullptr, nullptr);
        return -1;
    }

    auto timestamp = std::chrono::duration_cast<std::chrono::seconds>(
        scene.created_at.time_since_epoch()).count();

    sqlite3_bind_int64(stmt, 1, timestamp);
    sqlite3_bind_text(stmt, 2, scene.description.c_str(), -1, SQLITE_TRANSIENT);
    sqlite3_bind_text(stmt, 3, counts_ss.str().c_str(), -1, SQLITE_TRANSIENT);

    rc = sqlite3_step(stmt);
    sqlite3_finalize(stmt);

    if (rc != SQLITE_DONE) {
        logger_->error("Failed to insert scene: " + std::string(sqlite3_errmsg(db_)));
        sqlite3_exec(db_, "ROLLBACK", nullptr, nullptr, nullptr);
        return -1;
    }

    int scene_id = static_cast<int>(sqlite3_last_insert_rowid(db_));

    // Insert scene objects
    const char* insert_object = R"(
        INSERT INTO scene_objects 
        (scene_id, object_type, position_x, position_y, color_b, color_g, color_r, size_width, size_height, orientation)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    )";

    for (const auto& obj : scene.objects) {
        if (sqlite3_prepare_v2(db_, insert_object, -1, &stmt, nullptr) == SQLITE_OK) {
            sqlite3_bind_int(stmt, 1, scene_id);
            sqlite3_bind_text(stmt, 2, obj.object_type.c_str(), -1, SQLITE_TRANSIENT);
            sqlite3_bind_double(stmt, 3, obj.position.x);
            sqlite3_bind_double(stmt, 4, obj.position.y);
            sqlite3_bind_int(stmt, 5, static_cast<int>(obj.color[0]));
            sqlite3_bind_int(stmt, 6, static_cast<int>(obj.color[1]));
            sqlite3_bind_int(stmt, 7, static_cast<int>(obj.color[2]));
            sqlite3_bind_int(stmt, 8, obj.size.width);
            sqlite3_bind_int(stmt, 9, obj.size.height);
            sqlite3_bind_double(stmt, 10, obj.orientation);
            
            sqlite3_step(stmt);
            sqlite3_finalize(stmt);
        }
    }

    // Insert relationships
    const char* insert_rel = R"(
        INSERT INTO object_relationships (scene_id, object1_idx, object2_idx, distance, angle)
        VALUES (?, ?, ?, ?, ?)
    )";

    for (const auto& rel : scene.fingerprint.relationships) {
        if (sqlite3_prepare_v2(db_, insert_rel, -1, &stmt, nullptr) == SQLITE_OK) {
            sqlite3_bind_int(stmt, 1, scene_id);
            sqlite3_bind_int(stmt, 2, rel.object1_idx);
            sqlite3_bind_int(stmt, 3, rel.object2_idx);
            sqlite3_bind_double(stmt, 4, rel.distance);
            sqlite3_bind_double(stmt, 5, rel.angle);
            
            sqlite3_step(stmt);
            sqlite3_finalize(stmt);
        }
    }

    // Commit transaction
    sqlite3_exec(db_, "COMMIT", nullptr, nullptr, nullptr);

    logger_->info("Saved new scene (ID: " + std::to_string(scene_id) + "): " + scene.description);
    
    return scene_id;
}

std::string SceneManager::getSceneDescription(int scene_id) const {
    if (!initialized_ || !db_) {
        return "";
    }

    const char* query = "SELECT description FROM scenes WHERE id = ?";
    sqlite3_stmt* stmt;
    
    if (sqlite3_prepare_v2(db_, query, -1, &stmt, nullptr) != SQLITE_OK) {
        return "";
    }

    sqlite3_bind_int(stmt, 1, scene_id);
    
    std::string description;
    if (sqlite3_step(stmt) == SQLITE_ROW) {
        const char* desc = reinterpret_cast<const char*>(sqlite3_column_text(stmt, 0));
        if (desc) {
            description = desc;
        }
    }

    sqlite3_finalize(stmt);
    return description;
}

int SceneManager::analyzeAndMatchScene(
    const cv::Mat& frame,
    const std::vector<ObjectDetector::ObjectTracker>& tracked_objects) {
    
    if (!initialized_) {
        logger_->warning("Scene manager not initialized, skipping analysis");
        return -1;
    }

    // Update last analysis time
    last_analysis_time_ = std::chrono::steady_clock::now();

    // Analyze stationary objects
    auto scene_objects = analyzeStationaryObjects(frame, tracked_objects);
    
    if (scene_objects.empty()) {
        logger_->debug("No stationary objects to analyze");
        return -1;
    }

    // Create fingerprint
    auto fingerprint = createFingerprint(scene_objects);

    // Try to find matching scene
    int matching_scene_id = findMatchingScene(fingerprint);

    if (matching_scene_id >= 0) {
        // Found a matching scene
        std::string desc = getSceneDescription(matching_scene_id);
        logger_->info("Recognized return to earlier scene: id=" + std::to_string(matching_scene_id) + 
                     " (" + desc + ")");
        return matching_scene_id;
    }

    // Create and save new scene
    Scene new_scene;
    new_scene.id = -1;  // Will be set by database
    new_scene.created_at = std::chrono::system_clock::now();
    new_scene.objects = scene_objects;
    new_scene.fingerprint = fingerprint;
    new_scene.description = generateDescription(scene_objects);

    int new_scene_id = saveScene(new_scene);
    
    if (new_scene_id >= 0) {
        logger_->info("New scene identified: id=" + std::to_string(new_scene_id) + 
                     " (" + new_scene.description + ")");
    }

    return new_scene_id;
}
