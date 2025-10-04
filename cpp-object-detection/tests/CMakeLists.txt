# Test configuration
find_package(GTest REQUIRED)
find_package(OpenCV REQUIRED)

# Test sources
set(TEST_SOURCES
    test_config_manager.cpp
    test_logger.cpp
    test_performance_monitor.cpp
    test_webcam_interface.cpp
    test_object_detector.cpp
    test_parallel_frame_processor.cpp
    test_detection_model_interface.cpp
    test_viewfinder_window.cpp
    test_photo_storage_logic.cpp
    test_network_streamer.cpp
)

# Create test executable
add_executable(object_detection_tests ${TEST_SOURCES})

# Include directories
target_include_directories(object_detection_tests PRIVATE 
    ${CMAKE_CURRENT_SOURCE_DIR}/../include
    ${OpenCV_INCLUDE_DIRS}
)

# Link libraries
target_link_libraries(object_detection_tests 
    GTest::gtest
    GTest::gtest_main
    ${OpenCV_LIBS}
    pthread
)

# Add individual source files we want to test
target_sources(object_detection_tests PRIVATE
    ../src/config_manager.cpp
    ../src/logger.cpp
    ../src/performance_monitor.cpp
    ../src/webcam_interface.cpp
    ../src/object_detector.cpp
    ../src/parallel_frame_processor.cpp
    ../src/detection_model_factory.cpp
    ../src/yolo_v5_model.cpp
    ../src/viewfinder_window.cpp
    ../src/network_streamer.cpp
)

# Code coverage support for tests
if(ENABLE_COVERAGE)
    target_compile_options(object_detection_tests PRIVATE --coverage -O0 -g)
    target_link_libraries(object_detection_tests --coverage)
    
    # Add coverage target
    add_custom_target(coverage
        COMMAND lcov --directory . --capture --output-file coverage.info
        COMMAND lcov --remove coverage.info '/usr/*' --output-file coverage.info
        COMMAND lcov --list coverage.info
        COMMAND genhtml -o coverage coverage.info
        DEPENDS object_detection_tests
        WORKING_DIRECTORY ${CMAKE_BINARY_DIR}
        COMMENT "Generating code coverage report"
    )
endif()

# Add tests to CTest
add_test(NAME ObjectDetectionTests COMMAND object_detection_tests)