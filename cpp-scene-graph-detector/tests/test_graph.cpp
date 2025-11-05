#include <gtest/gtest.h>
#include "scene_graph/Graph.h"

using namespace scene_graph;

class GraphTest : public ::testing::Test {
protected:
    void SetUp() override {
        graph = std::make_unique<SceneGraph>();
    }

    std::unique_ptr<SceneGraph> graph;
};

TEST_F(GraphTest, EmptyGraph) {
    EXPECT_EQ(graph->getNodes().size(), 0);
    EXPECT_EQ(graph->getEdges().size(), 0);
}

TEST_F(GraphTest, AddNode) {
    Node node(0, 1, "person", BBox(100, 100, 50, 100), 0.85f);
    graph->addNode(node);
    
    EXPECT_EQ(graph->getNodes().size(), 1);
    EXPECT_EQ(graph->getNodes()[0].id, 0);
    EXPECT_EQ(graph->getNodes()[0].label, "person");
    EXPECT_FLOAT_EQ(graph->getNodes()[0].score, 0.85f);
}

TEST_F(GraphTest, AddEdge) {
    Edge edge(0, 1, 0, "left_of", 1.0f);
    graph->addEdge(edge);
    
    EXPECT_EQ(graph->getEdges().size(), 1);
    EXPECT_EQ(graph->getEdges()[0].src_id, 0);
    EXPECT_EQ(graph->getEdges()[0].dst_id, 1);
    EXPECT_EQ(graph->getEdges()[0].predicate_label, "left_of");
}

TEST_F(GraphTest, ClearGraph) {
    Node node(0, 1, "person", BBox(100, 100, 50, 100), 0.85f);
    Edge edge(0, 1, 0, "left_of", 1.0f);
    
    graph->addNode(node);
    graph->addEdge(edge);
    
    EXPECT_EQ(graph->getNodes().size(), 1);
    EXPECT_EQ(graph->getEdges().size(), 1);
    
    graph->clear();
    
    EXPECT_EQ(graph->getNodes().size(), 0);
    EXPECT_EQ(graph->getEdges().size(), 0);
}

TEST_F(GraphTest, JSONExport) {
    Node node1(0, 0, "person", BBox(100, 100, 50, 100), 0.85f);
    Node node2(1, 2, "car", BBox(300, 200, 150, 100), 0.92f);
    Edge edge(0, 1, 0, "left_of", 1.0f);
    
    graph->addNode(node1);
    graph->addNode(node2);
    graph->addEdge(edge);
    
    std::string json = graph->toJSON();
    
    // Verify JSON contains expected content
    EXPECT_NE(json.find("\"num_objects\": 2"), std::string::npos);
    EXPECT_NE(json.find("\"num_relations\": 1"), std::string::npos);
    EXPECT_NE(json.find("\"person\""), std::string::npos);
    EXPECT_NE(json.find("\"car\""), std::string::npos);
    EXPECT_NE(json.find("\"left_of\""), std::string::npos);
}

TEST_F(GraphTest, DOTExport) {
    Node node1(0, 0, "person", BBox(100, 100, 50, 100), 0.85f);
    Node node2(1, 2, "car", BBox(300, 200, 150, 100), 0.92f);
    Edge edge(0, 1, 0, "left_of", 1.0f);
    
    graph->addNode(node1);
    graph->addNode(node2);
    graph->addEdge(edge);
    
    std::string dot = graph->toDOT();
    
    // Verify DOT contains expected content
    EXPECT_NE(dot.find("digraph SceneGraph"), std::string::npos);
    EXPECT_NE(dot.find("node0"), std::string::npos);
    EXPECT_NE(dot.find("node1"), std::string::npos);
    EXPECT_NE(dot.find("person"), std::string::npos);
    EXPECT_NE(dot.find("car"), std::string::npos);
    EXPECT_NE(dot.find("left_of"), std::string::npos);
}

TEST_F(GraphTest, GetSummary) {
    Node node1(0, 0, "person", BBox(100, 100, 50, 100), 0.85f);
    Node node2(1, 2, "car", BBox(300, 200, 150, 100), 0.92f);
    Edge edge(0, 1, 0, "left_of", 1.0f);
    
    graph->addNode(node1);
    graph->addNode(node2);
    graph->addEdge(edge);
    
    std::string summary = graph->getSummary();
    
    EXPECT_NE(summary.find("Objects: 2"), std::string::npos);
    EXPECT_NE(summary.find("Relations: 1"), std::string::npos);
    EXPECT_NE(summary.find("person"), std::string::npos);
    EXPECT_NE(summary.find("car"), std::string::npos);
}
