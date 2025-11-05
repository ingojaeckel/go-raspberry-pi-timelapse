#ifndef SCENE_GRAPH_GRAPH_HPP
#define SCENE_GRAPH_GRAPH_HPP

#include <string>
#include <vector>
#include <memory>
#include <opencv2/opencv.hpp>

namespace scene_graph {

// Bounding box structure
struct BBox {
    float x;      // Center X
    float y;      // Center Y
    float width;  // Width
    float height; // Height

    BBox() : x(0), y(0), width(0), height(0) {}
    BBox(float x_, float y_, float w_, float h_) : x(x_), y(y_), width(w_), height(h_) {}
};

// Node represents an object in the scene
struct Node {
    int id;
    int class_id;
    std::string label;
    BBox bbox;
    float score;

    Node() : id(0), class_id(0), label(""), bbox(), score(0.0f) {}
    Node(int id_, int class_id_, const std::string& label_, const BBox& bbox_, float score_)
        : id(id_), class_id(class_id_), label(label_), bbox(bbox_), score(score_) {}
};

// Edge represents a relation between two objects
struct Edge {
    int src_id;
    int dst_id;
    int predicate_id;
    std::string predicate_label;
    float score;

    Edge() : src_id(0), dst_id(0), predicate_id(0), predicate_label(""), score(0.0f) {}
    Edge(int src_, int dst_, int pred_id_, const std::string& pred_label_, float score_)
        : src_id(src_), dst_id(dst_), predicate_id(pred_id_), predicate_label(pred_label_), score(score_) {}
};

// SceneGraph represents a complete scene graph
class SceneGraph {
public:
    SceneGraph();
    ~SceneGraph();

    // Add nodes and edges
    void addNode(const Node& node);
    void addEdge(const Edge& edge);

    // Clear the graph
    void clear();

    // Get nodes and edges
    const std::vector<Node>& getNodes() const { return nodes_; }
    const std::vector<Edge>& getEdges() const { return edges_; }

    // Export to JSON format
    std::string toJSON() const;

    // Export to Graphviz DOT format
    std::string toDOT() const;

    // Save to files
    bool saveJSON(const std::string& filepath) const;
    bool saveDOT(const std::string& filepath) const;

    // Get summary
    std::string getSummary() const;

private:
    std::vector<Node> nodes_;
    std::vector<Edge> edges_;

    std::string escapeJSON(const std::string& str) const;
    std::string escapeDOT(const std::string& str) const;
};

} // namespace scene_graph

#endif // SCENE_GRAPH_GRAPH_HPP
