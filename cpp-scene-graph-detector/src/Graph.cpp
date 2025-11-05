#include "scene_graph/Graph.h"
#include <fstream>
#include <sstream>
#include <iomanip>
#include <algorithm>

namespace scene_graph {

SceneGraph::SceneGraph() {}

SceneGraph::~SceneGraph() {}

void SceneGraph::addNode(const Node& node) {
    nodes_.push_back(node);
}

void SceneGraph::addEdge(const Edge& edge) {
    edges_.push_back(edge);
}

void SceneGraph::clear() {
    nodes_.clear();
    edges_.clear();
    metadata_.clear();
}

const Node* SceneGraph::findNode(int id) const {
    for (const auto& node : nodes_) {
        if (node.id == id) {
            return &node;
        }
    }
    return nullptr;
}

void SceneGraph::setMetadata(const std::string& key, const std::string& value) {
    metadata_[key] = value;
}

std::string SceneGraph::escapeJSON(const std::string& str) const {
    std::string escaped;
    for (char c : str) {
        switch (c) {
            case '"': escaped += "\\\""; break;
            case '\\': escaped += "\\\\"; break;
            case '\n': escaped += "\\n"; break;
            case '\r': escaped += "\\r"; break;
            case '\t': escaped += "\\t"; break;
            default: escaped += c;
        }
    }
    return escaped;
}

std::string SceneGraph::escapeDOT(const std::string& str) const {
    std::string escaped;
    for (char c : str) {
        if (c == '"' || c == '\\') {
            escaped += '\\';
        }
        escaped += c;
    }
    return escaped;
}

std::string SceneGraph::toJSON() const {
    std::ostringstream oss;
    oss << std::fixed << std::setprecision(3);
    
    oss << "{\n";
    
    // Metadata
    if (!metadata_.empty()) {
        oss << "  \"meta\": {\n";
        bool first = true;
        for (const auto& pair : metadata_) {
            if (!first) oss << ",\n";
            oss << "    \"" << escapeJSON(pair.first) << "\": \"" 
                << escapeJSON(pair.second) << "\"";
            first = false;
        }
        oss << "\n  },\n";
    }
    
    // Objects (nodes)
    oss << "  \"objects\": [\n";
    for (size_t i = 0; i < nodes_.size(); ++i) {
        const Node& node = nodes_[i];
        oss << "    {\n";
        oss << "      \"id\": " << node.id << ",\n";
        oss << "      \"class_id\": " << node.class_id << ",\n";
        oss << "      \"label\": \"" << escapeJSON(node.label) << "\",\n";
        oss << "      \"score\": " << node.score << ",\n";
        oss << "      \"bbox\": {\n";
        oss << "        \"x\": " << node.bbox.x << ",\n";
        oss << "        \"y\": " << node.bbox.y << ",\n";
        oss << "        \"width\": " << node.bbox.width << ",\n";
        oss << "        \"height\": " << node.bbox.height << "\n";
        oss << "      }\n";
        oss << "    }";
        if (i < nodes_.size() - 1) oss << ",";
        oss << "\n";
    }
    oss << "  ],\n";
    
    // Relations (edges)
    oss << "  \"relations\": [\n";
    for (size_t i = 0; i < edges_.size(); ++i) {
        const Edge& edge = edges_[i];
        oss << "    {\n";
        oss << "      \"subject_id\": " << edge.src_id << ",\n";
        oss << "      \"predicate\": \"" << escapeJSON(edge.label) << "\",\n";
        oss << "      \"object_id\": " << edge.dst_id << ",\n";
        oss << "      \"score\": " << edge.score << "\n";
        oss << "    }";
        if (i < edges_.size() - 1) oss << ",";
        oss << "\n";
    }
    oss << "  ]\n";
    
    oss << "}\n";
    return oss.str();
}

std::string SceneGraph::toDOT() const {
    std::ostringstream oss;
    
    oss << "digraph SceneGraph {\n";
    oss << "  rankdir=LR;\n";
    oss << "  node [shape=box];\n\n";
    
    // Add nodes
    for (const auto& node : nodes_) {
        oss << "  n" << node.id << " [label=\"" << escapeDOT(node.label) 
            << " (" << std::fixed << std::setprecision(2) << node.score << ")\"];\n";
    }
    
    oss << "\n";
    
    // Add edges
    for (const auto& edge : edges_) {
        oss << "  n" << edge.src_id << " -> n" << edge.dst_id 
            << " [label=\"" << escapeDOT(edge.label) << "\"];\n";
    }
    
    oss << "}\n";
    return oss.str();
}

void SceneGraph::toJSON(const std::string& filepath) const {
    std::ofstream ofs(filepath);
    if (!ofs.is_open()) {
        throw std::runtime_error("Failed to open file: " + filepath);
    }
    ofs << toJSON();
    ofs.close();
}

void SceneGraph::toDOT(const std::string& filepath) const {
    std::ofstream ofs(filepath);
    if (!ofs.is_open()) {
        throw std::runtime_error("Failed to open file: " + filepath);
    }
    ofs << toDOT();
    ofs.close();
}

} // namespace scene_graph
