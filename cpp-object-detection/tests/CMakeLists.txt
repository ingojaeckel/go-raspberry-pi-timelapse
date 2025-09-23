# Test configuration
find_package(GTest REQUIRED)

# Test sources
set(TEST_SOURCES
    test_config_manager.cpp
    test_logger.cpp
    test_performance_monitor.cpp
)

# Create test executable
add_executable(object_detection_tests ${TEST_SOURCES})

# Include directories
target_include_directories(object_detection_tests PRIVATE 
    ${CMAKE_CURRENT_SOURCE_DIR}/../include
    ${GTEST_INCLUDE_DIRS}
)

# Link libraries
target_link_libraries(object_detection_tests 
    ${GTEST_LIBRARIES}
    ${GTEST_MAIN_LIBRARIES}
    pthread
)

# Add individual source files we want to test
target_sources(object_detection_tests PRIVATE
    ../src/config_manager.cpp
    ../src/logger.cpp
    ../src/performance_monitor.cpp
)

# Add tests to CTest
add_test(NAME ObjectDetectionTests COMMAND object_detection_tests)