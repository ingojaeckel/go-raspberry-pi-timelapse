#ifndef SCENE_GRAPH_REL_PREDICTOR_H
#define SCENE_GRAPH_REL_PREDICTOR_H

#include <string>
#include <vector>
#include <opencv2/opencv.hpp>
#include "Graph.h"
#include "Detector.h"

namespace scene_graph {

// Spatial predicate types
enum class SpatialPredicate {
    LEFT_OF = 0,
    RIGHT_OF,
    IN_FRONT_OF,
    BEHIND,
    OVERLAPS,
    CONTAINS,
    INTERSECTS,
    BETWEEN,
    ON,
    UNDER,
    NEXT_TO,
    UNKNOWN
};

// Relation prediction result
struct RelationPrediction {
    int subject_id;  // Index of subject detection
    int object_id;   // Index of object detection
    SpatialPredicate predicate;
    std::string predicate_label;
    float score;
    
    RelationPrediction() : subject_id(-1), object_id(-1), 
                          predicate(SpatialPredicate::UNKNOWN), score(0.0f) {}
};

// Relation predictor interface
class RelPredictor {
public:
    RelPredictor();
    ~RelPredictor();
    
    // Initialize predictor
    bool initialize(const std::string& model_path = "",
                   const std::string& backend = "cpu");
    
    // Predict relations between detections
    std::vector<RelationPrediction> predict(const std::vector<Detection>& detections,
                                           const cv::Mat& image,
                                           float threshold = 0.5);
    
    // Check if initialized
    bool isInitialized() const { return initialized_; }
    
    // Get predicate label
    static std::string getPredicateLabel(SpatialPredicate pred);
    
private:
    cv::dnn::Net net_;
    bool initialized_;
    bool use_geometric_predictor_; // Use geometric rules if no model
    
    // Geometric-based relation prediction (fallback)
    std::vector<RelationPrediction> predictGeometric(const std::vector<Detection>& detections,
                                                     float threshold);
    
    // Compute spatial relations based on bounding boxes
    SpatialPredicate computeSpatialRelation(const BBox& bbox1, const BBox& bbox2);
    
    // Helper functions for geometric relations
    bool isLeftOf(const BBox& bbox1, const BBox& bbox2) const;
    bool isRightOf(const BBox& bbox1, const BBox& bbox2) const;
    bool overlaps(const BBox& bbox1, const BBox& bbox2) const;
    bool contains(const BBox& bbox1, const BBox& bbox2) const;
    float computeIoU(const BBox& bbox1, const BBox& bbox2) const;
};

// Factory function to load relation predictor
RelPredictor* loadRelPredictor(const std::string& model_path = "",
                               const std::string& backend = "cpu");

} // namespace scene_graph

#endif // SCENE_GRAPH_REL_PREDICTOR_H
