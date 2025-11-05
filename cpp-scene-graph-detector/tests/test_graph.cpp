#include <gtest/gtest.h>
#include "scene_graph/graph.h"

using namespace scene_graph;

TEST(GraphTest, CreateEmptyGraph) {
    SceneGraph graph;
    EXPECT_EQ(graph.getNodeCount(), 0);
    EXPECT_EQ(graph.getEdgeCount(), 0);
}

TEST(GraphTest, AddNode) {
    SceneGraph graph;
    
    Node node;
    node.id = 0;
    node.class_id = 1;
    node.label = "person";
    node.bbox = BBox(0.5f, 0.5f, 0.2f, 0.3f);
    node.score = 0.95f;
    
    graph.addNode(node);
    
    EXPECT_EQ(graph.getNodeCount(), 1);
    EXPECT_EQ(graph.getNodes()[0].label, "person");
    EXPECT_FLOAT_EQ(graph.getNodes()[0].score, 0.95f);
}

TEST(GraphTest, AddEdge) {
    SceneGraph graph;
    
    Node node1, node2;
    node1.id = 0;
    node1.label = "person";
    node2.id = 1;
    node2.label = "car";
    
    graph.addNode(node1);
    graph.addNode(node2);
    
    Edge edge;
    edge.src_id = 0;
    edge.dst_id = 1;
    edge.label = "next_to";
    edge.score = 1.0f;
    
    graph.addEdge(edge);
    
    EXPECT_EQ(graph.getEdgeCount(), 1);
    EXPECT_EQ(graph.getEdges()[0].label, "next_to");
}

TEST(GraphTest, FindNode) {
    SceneGraph graph;
    
    Node node;
    node.id = 42;
    node.label = "test";
    graph.addNode(node);
    
    const Node* found = graph.findNode(42);
    ASSERT_NE(found, nullptr);
    EXPECT_EQ(found->label, "test");
    
    const Node* not_found = graph.findNode(999);
    EXPECT_EQ(not_found, nullptr);
}

TEST(GraphTest, JSONExport) {
    SceneGraph graph;
    
    Node node;
    node.id = 0;
    node.class_id = 1;
    node.label = "person";
    node.bbox = BBox(0.5f, 0.5f, 0.2f, 0.3f);
    node.score = 0.95f;
    graph.addNode(node);
    
    Edge edge;
    edge.src_id = 0;
    edge.dst_id = 0;
    edge.label = "test_relation";
    edge.score = 1.0f;
    graph.addEdge(edge);
    
    graph.setMetadata("timestamp", "2024-01-01");
    
    std::string json = graph.toJSON();
    
    // Check that JSON contains expected fields
    EXPECT_NE(json.find("\"meta\""), std::string::npos);
    EXPECT_NE(json.find("\"objects\""), std::string::npos);
    EXPECT_NE(json.find("\"relations\""), std::string::npos);
    EXPECT_NE(json.find("\"person\""), std::string::npos);
    EXPECT_NE(json.find("\"test_relation\""), std::string::npos);
}

TEST(GraphTest, DOTExport) {
    SceneGraph graph;
    
    Node node1, node2;
    node1.id = 0;
    node1.label = "house";
    node2.id = 1;
    node2.label = "tree";
    
    graph.addNode(node1);
    graph.addNode(node2);
    
    Edge edge;
    edge.src_id = 0;
    edge.dst_id = 1;
    edge.label = "next_to";
    graph.addEdge(edge);
    
    std::string dot = graph.toDOT();
    
    // Check that DOT contains expected elements
    EXPECT_NE(dot.find("digraph SceneGraph"), std::string::npos);
    EXPECT_NE(dot.find("house"), std::string::npos);
    EXPECT_NE(dot.find("tree"), std::string::npos);
    EXPECT_NE(dot.find("next_to"), std::string::npos);
    EXPECT_NE(dot.find("->"), std::string::npos);
}

TEST(GraphTest, ClearGraph) {
    SceneGraph graph;
    
    Node node;
    node.id = 0;
    graph.addNode(node);
    
    Edge edge;
    edge.src_id = 0;
    edge.dst_id = 0;
    graph.addEdge(edge);
    
    graph.setMetadata("key", "value");
    
    EXPECT_EQ(graph.getNodeCount(), 1);
    EXPECT_EQ(graph.getEdgeCount(), 1);
    EXPECT_EQ(graph.getMetadata().size(), 1);
    
    graph.clear();
    
    EXPECT_EQ(graph.getNodeCount(), 0);
    EXPECT_EQ(graph.getEdgeCount(), 0);
    EXPECT_EQ(graph.getMetadata().size(), 0);
}

int main(int argc, char** argv) {
    ::testing::InitGoogleTest(&argc, argv);
    return RUN_ALL_TESTS();
}
