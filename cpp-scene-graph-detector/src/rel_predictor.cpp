#include "scene_graph/rel_predictor.h"
#include <iostream>
#include <algorithm>
#include <cmath>

namespace scene_graph {

RelPredictor::RelPredictor() 
    : initialized_(false), use_geometric_predictor_(true) {}

RelPredictor::~RelPredictor() {}

bool RelPredictor::initialize(const std::string& model_path,
                              const std::string& backend) {
    // If no model path provided, use geometric predictor
    if (model_path.empty()) {
        use_geometric_predictor_ = true;
        initialized_ = true;
        return true;
    }
    
    try {
        // Try to load ONNX model for learned relation prediction
        net_ = cv::dnn::readNetFromONNX(model_path);
        if (net_.empty()) {
            std::cerr << "Failed to load relation model, falling back to geometric predictor" << std::endl;
            use_geometric_predictor_ = true;
        } else {
            // Set backend
            if (backend == "opencl") {
                net_.setPreferableBackend(cv::dnn::DNN_BACKEND_OPENCV);
                net_.setPreferableTarget(cv::dnn::DNN_TARGET_OPENCL);
            } else {
                net_.setPreferableBackend(cv::dnn::DNN_BACKEND_OPENCV);
                net_.setPreferableTarget(cv::dnn::DNN_TARGET_CPU);
            }
            use_geometric_predictor_ = false;
        }
        
        initialized_ = true;
        return true;
    } catch (const cv::Exception& e) {
        std::cerr << "OpenCV exception in RelPredictor::initialize: " << e.what() << std::endl;
        use_geometric_predictor_ = true;
        initialized_ = true;
        return true;
    }
}

std::vector<RelationPrediction> RelPredictor::predict(
    const std::vector<Detection>& detections,
    const cv::Mat& image,
    float threshold) {
    
    if (!initialized_) {
        return std::vector<RelationPrediction>();
    }
    
    if (use_geometric_predictor_) {
        return predictGeometric(detections, threshold);
    }
    
    // TODO: Implement learned model prediction
    // For now, fall back to geometric
    return predictGeometric(detections, threshold);
}

std::vector<RelationPrediction> RelPredictor::predictGeometric(
    const std::vector<Detection>& detections,
    float threshold) {
    
    std::vector<RelationPrediction> predictions;
    
    // Predict pairwise relations
    for (size_t i = 0; i < detections.size(); ++i) {
        for (size_t j = 0; j < detections.size(); ++j) {
            if (i == j) continue;
            
            const BBox& bbox1 = detections[i].bbox;
            const BBox& bbox2 = detections[j].bbox;
            
            SpatialPredicate pred = computeSpatialRelation(bbox1, bbox2);
            
            if (pred != SpatialPredicate::UNKNOWN) {
                RelationPrediction rel_pred;
                rel_pred.subject_id = static_cast<int>(i);
                rel_pred.object_id = static_cast<int>(j);
                rel_pred.predicate = pred;
                rel_pred.predicate_label = getPredicateLabel(pred);
                rel_pred.score = 1.0f; // Geometric relations are deterministic
                
                predictions.push_back(rel_pred);
            }
        }
    }
    
    return predictions;
}

SpatialPredicate RelPredictor::computeSpatialRelation(const BBox& bbox1, const BBox& bbox2) {
    // Check various spatial relationships in order of specificity
    
    // Contains: bbox1 fully contains bbox2
    if (contains(bbox1, bbox2)) {
        return SpatialPredicate::CONTAINS;
    }
    
    // Overlaps: significant overlap but not contains
    float iou = computeIoU(bbox1, bbox2);
    if (iou > 0.1f) {
        return SpatialPredicate::OVERLAPS;
    }
    
    // Left/Right
    if (isLeftOf(bbox1, bbox2)) {
        return SpatialPredicate::LEFT_OF;
    }
    if (isRightOf(bbox1, bbox2)) {
        return SpatialPredicate::RIGHT_OF;
    }
    
    // Vertical relations (on/under) - based on Y coordinate
    float y1_bottom = bbox1.y + bbox1.height / 2;
    float y2_top = bbox2.y - bbox2.height / 2;
    float y1_top = bbox1.y - bbox1.height / 2;
    float y2_bottom = bbox2.y + bbox2.height / 2;
    
    // Check if bbox1 is above bbox2 (y increases downward in image coords)
    if (y1_bottom < y2_top && std::abs(bbox1.x - bbox2.x) < 0.2f) {
        return SpatialPredicate::ON;
    }
    
    // Check if bbox1 is below bbox2
    if (y1_top > y2_bottom && std::abs(bbox1.x - bbox2.x) < 0.2f) {
        return SpatialPredicate::UNDER;
    }
    
    // Next to: close proximity horizontally
    float x_dist = std::abs(bbox1.x - bbox2.x);
    float y_dist = std::abs(bbox1.y - bbox2.y);
    if (x_dist < 0.3f && y_dist < 0.1f) {
        return SpatialPredicate::NEXT_TO;
    }
    
    return SpatialPredicate::UNKNOWN;
}

bool RelPredictor::isLeftOf(const BBox& bbox1, const BBox& bbox2) const {
    float bbox1_right = bbox1.x + bbox1.width / 2;
    float bbox2_left = bbox2.x - bbox2.width / 2;
    return (bbox1_right < bbox2_left - 0.05f); // Small margin
}

bool RelPredictor::isRightOf(const BBox& bbox1, const BBox& bbox2) const {
    float bbox1_left = bbox1.x - bbox1.width / 2;
    float bbox2_right = bbox2.x + bbox2.width / 2;
    return (bbox1_left > bbox2_right + 0.05f); // Small margin
}

bool RelPredictor::overlaps(const BBox& bbox1, const BBox& bbox2) const {
    float iou = computeIoU(bbox1, bbox2);
    return iou > 0.1f;
}

bool RelPredictor::contains(const BBox& bbox1, const BBox& bbox2) const {
    float bbox1_left = bbox1.x - bbox1.width / 2;
    float bbox1_right = bbox1.x + bbox1.width / 2;
    float bbox1_top = bbox1.y - bbox1.height / 2;
    float bbox1_bottom = bbox1.y + bbox1.height / 2;
    
    float bbox2_left = bbox2.x - bbox2.width / 2;
    float bbox2_right = bbox2.x + bbox2.width / 2;
    float bbox2_top = bbox2.y - bbox2.height / 2;
    float bbox2_bottom = bbox2.y + bbox2.height / 2;
    
    return (bbox1_left <= bbox2_left && bbox1_right >= bbox2_right &&
            bbox1_top <= bbox2_top && bbox1_bottom >= bbox2_bottom);
}

float RelPredictor::computeIoU(const BBox& bbox1, const BBox& bbox2) const {
    float x1 = std::max(bbox1.x - bbox1.width/2, bbox2.x - bbox2.width/2);
    float y1 = std::max(bbox1.y - bbox1.height/2, bbox2.y - bbox2.height/2);
    float x2 = std::min(bbox1.x + bbox1.width/2, bbox2.x + bbox2.width/2);
    float y2 = std::min(bbox1.y + bbox1.height/2, bbox2.y + bbox2.height/2);
    
    float inter_area = std::max(0.0f, x2 - x1) * std::max(0.0f, y2 - y1);
    float box1_area = bbox1.width * bbox1.height;
    float box2_area = bbox2.width * bbox2.height;
    float union_area = box1_area + box2_area - inter_area;
    
    return (union_area > 0) ? (inter_area / union_area) : 0.0f;
}

std::string RelPredictor::getPredicateLabel(SpatialPredicate pred) {
    switch (pred) {
        case SpatialPredicate::LEFT_OF: return "left_of";
        case SpatialPredicate::RIGHT_OF: return "right_of";
        case SpatialPredicate::IN_FRONT_OF: return "in_front_of";
        case SpatialPredicate::BEHIND: return "behind";
        case SpatialPredicate::OVERLAPS: return "overlaps";
        case SpatialPredicate::CONTAINS: return "contains";
        case SpatialPredicate::INTERSECTS: return "intersects";
        case SpatialPredicate::BETWEEN: return "between";
        case SpatialPredicate::ON: return "on";
        case SpatialPredicate::UNDER: return "under";
        case SpatialPredicate::NEXT_TO: return "next_to";
        default: return "unknown";
    }
}

RelPredictor* loadRelPredictor(const std::string& model_path,
                               const std::string& backend) {
    RelPredictor* predictor = new RelPredictor();
    if (!predictor->initialize(model_path, backend)) {
        delete predictor;
        return nullptr;
    }
    return predictor;
}

} // namespace scene_graph
