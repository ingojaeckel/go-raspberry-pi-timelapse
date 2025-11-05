#ifndef SCENE_GRAPH_REL_PREDICTOR_HPP
#define SCENE_GRAPH_REL_PREDICTOR_HPP

#include <string>
#include <vector>
#include <memory>
#include <opencv2/opencv.hpp>
#include "Graph.h"
#include "Detector.h"

namespace scene_graph {

// Spatial predicate types
enum class SpatialPredicate {
    LEFT_OF = 0,
    RIGHT_OF = 1,
    IN_FRONT_OF = 2,
    BEHIND = 3,
    OVERLAPS = 4,
    CONTAINS = 5,
    INTERSECTS = 6,
    BETWEEN = 7,
    ON = 8,
    UNDER = 9,
    NEXT_TO = 10,
    UNKNOWN = 99
};

// Relation prediction result
struct RelationPrediction {
    int subject_id;
    int object_id;
    SpatialPredicate predicate;
    std::string predicate_label;
    float confidence;

    RelationPrediction() 
        : subject_id(0), object_id(0), predicate(SpatialPredicate::UNKNOWN), 
          predicate_label("unknown"), confidence(0.0f) {}
};

// RelPredictor class for predicting spatial relations
class RelPredictor {
public:
    RelPredictor();
    ~RelPredictor();

    // Load relation predictor model (optional - can use geometric rules)
    bool loadRelPredictor(const std::string& model_path);

    // Predict relations between detections
    std::vector<RelationPrediction> infer(const std::vector<Detection>& detections,
                                         float confidence_threshold = 0.5f);

    // Check if model is loaded
    bool isModelLoaded() const { return model_loaded_; }

    // Get predicate label from enum
    static std::string getPredicateLabel(SpatialPredicate predicate);
    
    // Get predicate ID from enum
    static int getPredicateId(SpatialPredicate predicate);

private:
    cv::dnn::Net net_;
    bool model_loaded_;

    // Geometric relation inference (fallback when no model)
    std::vector<RelationPrediction> inferGeometric(const std::vector<Detection>& detections,
                                                   float confidence_threshold);

    // Geometric predicates
    bool isLeftOf(const cv::Rect& a, const cv::Rect& b) const;
    bool isRightOf(const cv::Rect& a, const cv::Rect& b) const;
    bool overlaps(const cv::Rect& a, const cv::Rect& b) const;
    bool contains(const cv::Rect& a, const cv::Rect& b) const;
    float computeIoU(const cv::Rect& a, const cv::Rect& b) const;
};

} // namespace scene_graph

#endif // SCENE_GRAPH_REL_PREDICTOR_HPP
