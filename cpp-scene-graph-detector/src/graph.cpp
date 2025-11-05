#include "scene_graph/Graph.h"
#include <sstream>
#include <fstream>
#include <iomanip>

namespace scene_graph {

SceneGraph::SceneGraph() {
}

SceneGraph::~SceneGraph() {
}

void SceneGraph::addNode(const Node& node) {
    nodes_.push_back(node);
}

void SceneGraph::addEdge(const Edge& edge) {
    edges_.push_back(edge);
}

void SceneGraph::clear() {
    nodes_.clear();
    edges_.clear();
}

std::string SceneGraph::escapeJSON(const std::string& str) const {
    std::ostringstream oss;
    for (char c : str) {
        switch (c) {
            case '"': oss << "\\\""; break;
            case '\\': oss << "\\\\"; break;
            case '\b': oss << "\\b"; break;
            case '\f': oss << "\\f"; break;
            case '\n': oss << "\\n"; break;
            case '\r': oss << "\\r"; break;
            case '\t': oss << "\\t"; break;
            default: oss << c; break;
        }
    }
    return oss.str();
}

std::string SceneGraph::escapeDOT(const std::string& str) const {
    std::ostringstream oss;
    for (char c : str) {
        if (c == '"' || c == '\\') {
            oss << '\\';
        }
        oss << c;
    }
    return oss.str();
}

std::string SceneGraph::toJSON() const {
    std::ostringstream json;
    json << std::fixed << std::setprecision(2);
    
    json << "{\n";
    json << "  \"meta\": {\n";
    json << "    \"num_objects\": " << nodes_.size() << ",\n";
    json << "    \"num_relations\": " << edges_.size() << "\n";
    json << "  },\n";
    
    json << "  \"objects\": [\n";
    for (size_t i = 0; i < nodes_.size(); ++i) {
        const auto& node = nodes_[i];
        json << "    {\n";
        json << "      \"id\": " << node.id << ",\n";
        json << "      \"label\": \"" << escapeJSON(node.label) << "\",\n";
        json << "      \"class_id\": " << node.class_id << ",\n";
        json << "      \"score\": " << node.score << ",\n";
        json << "      \"bbox\": {\n";
        json << "        \"x\": " << node.bbox.x << ",\n";
        json << "        \"y\": " << node.bbox.y << ",\n";
        json << "        \"width\": " << node.bbox.width << ",\n";
        json << "        \"height\": " << node.bbox.height << "\n";
        json << "      }\n";
        json << "    }";
        if (i < nodes_.size() - 1) json << ",";
        json << "\n";
    }
    json << "  ],\n";
    
    json << "  \"relations\": [\n";
    for (size_t i = 0; i < edges_.size(); ++i) {
        const auto& edge = edges_[i];
        json << "    {\n";
        json << "      \"subject_id\": " << edge.src_id << ",\n";
        json << "      \"predicate\": \"" << escapeJSON(edge.predicate_label) << "\",\n";
        json << "      \"object_id\": " << edge.dst_id << ",\n";
        json << "      \"score\": " << edge.score << "\n";
        json << "    }";
        if (i < edges_.size() - 1) json << ",";
        json << "\n";
    }
    json << "  ]\n";
    json << "}\n";
    
    return json.str();
}

std::string SceneGraph::toDOT() const {
    std::ostringstream dot;
    
    dot << "digraph SceneGraph {\n";
    dot << "  rankdir=LR;\n";
    dot << "  node [shape=box];\n\n";
    
    // Add nodes
    for (const auto& node : nodes_) {
        dot << "  node" << node.id << " [label=\"" 
            << escapeDOT(node.label) << "\\nID:" << node.id 
            << "\\nScore:" << std::fixed << std::setprecision(2) << node.score
            << "\"];\n";
    }
    
    dot << "\n";
    
    // Add edges
    for (const auto& edge : edges_) {
        dot << "  node" << edge.src_id << " -> node" << edge.dst_id
            << " [label=\"" << escapeDOT(edge.predicate_label) 
            << " (" << std::fixed << std::setprecision(2) << edge.score << ")\"];\n";
    }
    
    dot << "}\n";
    
    return dot.str();
}

bool SceneGraph::saveJSON(const std::string& filepath) const {
    std::ofstream file(filepath);
    if (!file.is_open()) {
        return false;
    }
    file << toJSON();
    file.close();
    return true;
}

bool SceneGraph::saveDOT(const std::string& filepath) const {
    std::ofstream file(filepath);
    if (!file.is_open()) {
        return false;
    }
    file << toDOT();
    file.close();
    return true;
}

std::string SceneGraph::getSummary() const {
    std::ostringstream summary;
    summary << "Scene Graph Summary:\n";
    summary << "  Objects: " << nodes_.size() << "\n";
    summary << "  Relations: " << edges_.size() << "\n";
    
    if (!nodes_.empty()) {
        summary << "\n  Detected Objects:\n";
        for (const auto& node : nodes_) {
            summary << "    - " << node.label << " (ID: " << node.id 
                   << ", Score: " << std::fixed << std::setprecision(2) << node.score << ")\n";
        }
    }
    
    if (!edges_.empty()) {
        summary << "\n  Relations:\n";
        for (const auto& edge : edges_) {
            summary << "    - Object " << edge.src_id << " " 
                   << edge.predicate_label << " Object " << edge.dst_id
                   << " (Score: " << std::fixed << std::setprecision(2) << edge.score << ")\n";
        }
    }
    
    return summary.str();
}

} // namespace scene_graph
