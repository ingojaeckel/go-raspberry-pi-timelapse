cmake_minimum_required(VERSION 3.16)

# Find GTest
find_package(GTest REQUIRED)
include_directories(${GTEST_INCLUDE_DIRS})
include_directories(${CMAKE_SOURCE_DIR}/include)

# Test source files
set(TEST_SOURCES
    test_graph.cpp
    test_detector.cpp
    test_rel_predictor.cpp
)

# Create test executable
add_executable(scene_graph_tests ${TEST_SOURCES})

# Link libraries
target_link_libraries(scene_graph_tests
    ${GTEST_BOTH_LIBRARIES}
    ${OpenCV_LIBS}
    Threads::Threads
    pthread
)

# Add library sources (not main.cpp)
target_sources(scene_graph_tests PRIVATE
    ${CMAKE_SOURCE_DIR}/src/Graph.cpp
    ${CMAKE_SOURCE_DIR}/src/Detector.cpp
    ${CMAKE_SOURCE_DIR}/src/RelPredictor.cpp
    ${CMAKE_SOURCE_DIR}/src/Runner.cpp
)

# Add tests to CTest
add_test(NAME SceneGraphTests COMMAND scene_graph_tests)
