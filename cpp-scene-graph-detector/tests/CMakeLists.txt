# Test configuration
find_package(GTest REQUIRED)
find_package(OpenCV REQUIRED)

# Test sources
set(TEST_SOURCES
    test_graph.cpp
    test_detector.cpp
    test_rel_predictor.cpp
    test_config.cpp
)

# Create test executable
add_executable(scene_graph_tests ${TEST_SOURCES})

# Include directories
target_include_directories(scene_graph_tests PRIVATE 
    ${CMAKE_CURRENT_SOURCE_DIR}/../include
    ${OpenCV_INCLUDE_DIRS}
)

# Link libraries
target_link_libraries(scene_graph_tests 
    GTest::gtest
    GTest::gtest_main
    ${OpenCV_LIBS}
    pthread
)

# Add individual source files we want to test
target_sources(scene_graph_tests PRIVATE
    ../src/graph.cpp
    ../src/detector.cpp
    ../src/rel_predictor.cpp
    ../src/runner.cpp
    ../src/config_manager.cpp
    ../src/logger.cpp
)

# Add tests to CTest
add_test(NAME SceneGraphTests COMMAND scene_graph_tests)
