#ifndef SCENE_GRAPH_GRAPH_H
#define SCENE_GRAPH_GRAPH_H

#include <string>
#include <vector>
#include <map>

namespace scene_graph {

// Bounding box structure
struct BBox {
    float x;      // Center X coordinate (normalized 0-1)
    float y;      // Center Y coordinate (normalized 0-1)
    float width;  // Width (normalized 0-1)
    float height; // Height (normalized 0-1)
    
    BBox() : x(0.0f), y(0.0f), width(0.0f), height(0.0f) {}
    BBox(float x_, float y_, float w_, float h_) : x(x_), y(y_), width(w_), height(h_) {}
};

// Node representing a detected object in the scene
struct Node {
    int id;               // Unique node ID
    int class_id;         // Object class ID from model
    std::string label;    // Human-readable object label
    BBox bbox;            // Bounding box
    float score;          // Detection confidence score (0-1)
    
    Node() : id(-1), class_id(-1), score(0.0f) {}
};

// Edge representing a relationship between two objects
struct Edge {
    int src_id;           // Source node ID (subject)
    int dst_id;           // Destination node ID (object)
    int predicate_id;     // Predicate/relation class ID
    std::string label;    // Human-readable predicate label
    float score;          // Relation confidence score (0-1)
    
    Edge() : src_id(-1), dst_id(-1), predicate_id(-1), score(0.0f) {}
};

// Complete scene graph representation
class SceneGraph {
public:
    SceneGraph();
    ~SceneGraph();
    
    // Add nodes and edges
    void addNode(const Node& node);
    void addEdge(const Edge& edge);
    
    // Clear the graph
    void clear();
    
    // Accessors
    const std::vector<Node>& getNodes() const { return nodes_; }
    const std::vector<Edge>& getEdges() const { return edges_; }
    size_t getNodeCount() const { return nodes_.size(); }
    size_t getEdgeCount() const { return edges_.size(); }
    
    // Find node by ID
    const Node* findNode(int id) const;
    
    // Export functions
    std::string toJSON() const;
    std::string toDOT() const;
    void toJSON(const std::string& filepath) const;
    void toDOT(const std::string& filepath) const;
    
    // Set metadata
    void setMetadata(const std::string& key, const std::string& value);
    const std::map<std::string, std::string>& getMetadata() const { return metadata_; }
    
private:
    std::vector<Node> nodes_;
    std::vector<Edge> edges_;
    std::map<std::string, std::string> metadata_;
    
    std::string escapeJSON(const std::string& str) const;
    std::string escapeDOT(const std::string& str) const;
};

} // namespace scene_graph

#endif // SCENE_GRAPH_GRAPH_H
