#include "scene_graph/RelPredictor.h"
#include "scene_graph/logger.h"

namespace scene_graph {

RelPredictor::RelPredictor() : model_loaded_(false) {
}

RelPredictor::~RelPredictor() {
}

bool RelPredictor::loadRelPredictor(const std::string& model_path) {
    auto& logger = Logger::getInstance();
    
    try {
        net_ = cv::dnn::readNetFromONNX(model_path);
        if (net_.empty()) {
            logger.warning("Failed to load relation predictor model: " + model_path);
            logger.info("Will use geometric relation inference as fallback");
            return false;
        }
        
        model_loaded_ = true;
        logger.info("Relation predictor model loaded successfully");
        return true;
        
    } catch (const cv::Exception& e) {
        logger.warning("OpenCV exception while loading relation predictor: " + std::string(e.what()));
        logger.info("Will use geometric relation inference as fallback");
        return false;
    } catch (const std::exception& e) {
        logger.warning("Exception while loading relation predictor: " + std::string(e.what()));
        logger.info("Will use geometric relation inference as fallback");
        return false;
    }
}

std::vector<RelationPrediction> RelPredictor::infer(const std::vector<Detection>& detections,
                                                    float confidence_threshold) {
    if (!model_loaded_) {
        // Use geometric inference as fallback
        return inferGeometric(detections, confidence_threshold);
    }
    
    // TODO: Implement model-based inference when model is available
    // For now, fall back to geometric inference
    return inferGeometric(detections, confidence_threshold);
}

std::vector<RelationPrediction> RelPredictor::inferGeometric(const std::vector<Detection>& detections,
                                                             float confidence_threshold) {
    std::vector<RelationPrediction> relations;
    
    // Generate pairwise relations based on geometric properties
    for (size_t i = 0; i < detections.size(); ++i) {
        for (size_t j = 0; j < detections.size(); ++j) {
            if (i == j) continue;
            
            const auto& det_i = detections[i];
            const auto& det_j = detections[j];
            
            RelationPrediction rel;
            rel.subject_id = static_cast<int>(i);
            rel.object_id = static_cast<int>(j);
            rel.confidence = 1.0f;  // Geometric relations have high confidence
            
            // Check various spatial relations
            if (isLeftOf(det_i.box, det_j.box)) {
                rel.predicate = SpatialPredicate::LEFT_OF;
                rel.predicate_label = getPredicateLabel(rel.predicate);
                relations.push_back(rel);
            } else if (isRightOf(det_i.box, det_j.box)) {
                rel.predicate = SpatialPredicate::RIGHT_OF;
                rel.predicate_label = getPredicateLabel(rel.predicate);
                relations.push_back(rel);
            }
            
            if (overlaps(det_i.box, det_j.box)) {
                rel.predicate = SpatialPredicate::OVERLAPS;
                rel.predicate_label = getPredicateLabel(rel.predicate);
                rel.confidence = computeIoU(det_i.box, det_j.box);
                relations.push_back(rel);
            }
            
            if (contains(det_i.box, det_j.box)) {
                rel.predicate = SpatialPredicate::CONTAINS;
                rel.predicate_label = getPredicateLabel(rel.predicate);
                relations.push_back(rel);
            }
        }
    }
    
    // Filter by confidence threshold
    std::vector<RelationPrediction> filtered;
    for (const auto& rel : relations) {
        if (rel.confidence >= confidence_threshold) {
            filtered.push_back(rel);
        }
    }
    
    return filtered;
}

bool RelPredictor::isLeftOf(const cv::Rect& a, const cv::Rect& b) const {
    // A is left of B if A's right edge is before B's center
    int a_right = a.x + a.width;
    int b_center = b.x + b.width / 2;
    return a_right < b_center && (a_right < b.x + 10);  // Small overlap tolerance
}

bool RelPredictor::isRightOf(const cv::Rect& a, const cv::Rect& b) const {
    // A is right of B if A's left edge is after B's center
    int a_left = a.x;
    int b_center = b.x + b.width / 2;
    return a_left > b_center && (a_left > b.x + b.width - 10);  // Small overlap tolerance
}

bool RelPredictor::overlaps(const cv::Rect& a, const cv::Rect& b) const {
    return computeIoU(a, b) > 0.0f;
}

bool RelPredictor::contains(const cv::Rect& a, const cv::Rect& b) const {
    // A contains B if B is entirely within A
    return (b.x >= a.x && 
            b.y >= a.y && 
            b.x + b.width <= a.x + a.width && 
            b.y + b.height <= a.y + a.height);
}

float RelPredictor::computeIoU(const cv::Rect& a, const cv::Rect& b) const {
    int x1 = std::max(a.x, b.x);
    int y1 = std::max(a.y, b.y);
    int x2 = std::min(a.x + a.width, b.x + b.width);
    int y2 = std::min(a.y + a.height, b.y + b.height);
    
    int intersection_area = std::max(0, x2 - x1) * std::max(0, y2 - y1);
    int union_area = a.area() + b.area() - intersection_area;
    
    return (union_area > 0) ? static_cast<float>(intersection_area) / union_area : 0.0f;
}

std::string RelPredictor::getPredicateLabel(SpatialPredicate predicate) {
    switch (predicate) {
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

int RelPredictor::getPredicateId(SpatialPredicate predicate) {
    return static_cast<int>(predicate);
}

} // namespace scene_graph
