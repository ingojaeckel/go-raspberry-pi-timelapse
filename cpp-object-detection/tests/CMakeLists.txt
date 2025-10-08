# Test configuration
find_package(GTest REQUIRED)
find_package(OpenCV REQUIRED)

# CURL is already found in parent CMakeLists.txt
# Use the same approach as parent
find_package(PkgConfig)
if(PKG_CONFIG_FOUND)
    pkg_check_modules(CURL REQUIRED libcurl)
    if(CURL_FOUND)
        # Use pkg-config results
        include_directories(${CURL_INCLUDE_DIRS})
        link_directories(${CURL_LIBRARY_DIRS})
    endif()
else()
    # Fall back to find_package
    find_package(CURL REQUIRED)
endif()

# Test sources
set(TEST_SOURCES
    test_config_manager.cpp
    test_logger.cpp
    test_hourly_summary.cpp
    test_performance_monitor.cpp
    test_webcam_interface.cpp
    test_object_detector.cpp
    test_parallel_frame_processor.cpp
    test_detection_model_interface.cpp
    test_viewfinder_window.cpp
    test_photo_storage_logic.cpp
    test_network_streamer.cpp
    test_long_term_operation.cpp
    test_stationary_detection.cpp
    test_google_sheets_client.cpp
)

# Create test executable
add_executable(object_detection_tests ${TEST_SOURCES})

# Include directories
target_include_directories(object_detection_tests PRIVATE 
    ${CMAKE_CURRENT_SOURCE_DIR}/../include
    ${OpenCV_INCLUDE_DIRS}
)
if(PKG_CONFIG_FOUND AND CURL_FOUND)
    target_include_directories(object_detection_tests PRIVATE ${CURL_INCLUDE_DIRS})
else()
    target_include_directories(object_detection_tests PRIVATE ${CURL_INCLUDE_DIRS})
endif()

# Link libraries
if(PKG_CONFIG_FOUND AND CURL_FOUND)
    # Link using pkg-config results
    target_link_libraries(object_detection_tests 
        GTest::gtest
        GTest::gtest_main
        ${OpenCV_LIBS}
        pthread
        ${CURL_LIBRARIES}
    )
else()
    # Link using find_package results
    target_link_libraries(object_detection_tests 
        GTest::gtest
        GTest::gtest_main
        ${OpenCV_LIBS}
        pthread
        CURL::libcurl
    )
endif()

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
    ../src/efficientdet_d3_model.cpp
    ../src/viewfinder_window.cpp
    ../src/network_streamer.cpp
    ../src/system_monitor.cpp
    ../src/google_sheets_client.cpp
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