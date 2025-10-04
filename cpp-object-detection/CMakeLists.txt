cmake_minimum_required(VERSION 3.16)
project(ObjectDetection VERSION 1.0.0 LANGUAGES CXX)

# Set C++ standard
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Build configuration
set(CMAKE_BUILD_TYPE Release)

# Enable static linking for standalone executable
if(APPLE)
    # macOS doesn't support static-libgcc/libstdc++ in the same way
    # Use default dynamic linking with runtime resolution
    set(CMAKE_EXE_LINKER_FLAGS "")
else()
    # Linux static linking
    set(CMAKE_EXE_LINKER_FLAGS "-static-libgcc -static-libstdc++")
endif()

# Find required packages  
# OpenCV 4.x uses opencv4 as package name
if(APPLE)
    # macOS with Homebrew - find OpenCV without specific paths
    find_package(OpenCV REQUIRED)
else()
    # Linux - try multiple approaches to find OpenCV
    find_package(OpenCV QUIET)
    if(NOT OpenCV_FOUND)
        find_package(OpenCV REQUIRED PATHS /usr/lib/x86_64-linux-gnu/cmake/opencv4 NO_DEFAULT_PATH)
    endif()
    if(NOT OpenCV_FOUND)
        find_package(opencv4 REQUIRED)
    endif()
endif()
find_package(Threads REQUIRED)

# Include directories
include_directories(${CMAKE_CURRENT_SOURCE_DIR}/include)
include_directories(${OpenCV_INCLUDE_DIRS})

# Source files
set(SOURCES
    src/main.cpp
    src/application.cpp
    src/webcam_interface.cpp
    src/object_detector.cpp
    src/config_manager.cpp
    src/logger.cpp
    src/performance_monitor.cpp
    src/parallel_frame_processor.cpp
    src/detection_model_factory.cpp
    src/yolo_v5_model.cpp
    src/viewfinder_window.cpp
    src/network_streamer.cpp
)

# Create executable
add_executable(object_detection ${SOURCES})

# Link libraries
target_link_libraries(object_detection 
    ${OpenCV_LIBS}
    Threads::Threads
)

# Compiler flags for optimization and warnings
if(APPLE)
    # macOS-specific optimizations
    target_compile_options(object_detection PRIVATE 
        -O3 
        -Wall 
        -Wextra 
        -pedantic
        -mtune=native  # Use mtune instead of march on macOS for better compatibility
    )
    # Add camera usage description for macOS permissions
    add_compile_definitions(
        "CAMERA_USAGE_DESCRIPTION=\"This app needs camera access for object detection\""
    )
    set_target_properties(object_detection PROPERTIES
        MACOSX_BUNDLE TRUE
        MACOSX_BUNDLE_INFO_PLIST ${CMAKE_CURRENT_SOURCE_DIR}/Info.plist
    )
else()
    # Linux-specific optimizations
    target_compile_options(object_detection PRIVATE 
        -O3 
        -Wall 
        -Wextra 
        -pedantic
        -march=native
    )
endif()

# Code coverage support
option(ENABLE_COVERAGE "Enable code coverage" OFF)
if(ENABLE_COVERAGE)
    message(STATUS "Code coverage enabled")
    target_compile_options(object_detection PRIVATE --coverage -O0 -g)
    target_link_libraries(object_detection --coverage)
    
    # Find gcov
    find_program(GCOV_PATH gcov)
    if(NOT GCOV_PATH)
        message(FATAL_ERROR "gcov not found! Aborting...")
    endif()
    
    # Find lcov
    find_program(LCOV_PATH lcov)
    if(NOT LCOV_PATH)
        message(WARNING "lcov not found! Coverage reports will be limited...")
    endif()
endif()

# Enable testing (optional)
enable_testing()

# Add test subdirectory if it exists and GTest is available
find_package(GTest QUIET)
if(EXISTS ${CMAKE_CURRENT_SOURCE_DIR}/tests AND GTest_FOUND)
    message(STATUS "Google Test found, building tests")
    add_subdirectory(tests)
else()
    message(STATUS "Google Test not found or tests directory missing, skipping tests")
endif()

# Installation
install(TARGETS object_detection DESTINATION bin)