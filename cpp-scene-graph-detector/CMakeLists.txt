cmake_minimum_required(VERSION 3.16)
project(SceneGraphDetector VERSION 1.0.0 LANGUAGES CXX)

# Set C++ standard
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Build configuration
set(CMAKE_BUILD_TYPE Release)

# CMake options
option(SCENE_GRAPH_WITH_OPENCL "Enable OpenCL backend support" OFF)
option(SCENE_GRAPH_BUILD_TESTS "Build unit tests" ON)

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
        # Try x86_64 path first
        find_package(OpenCV QUIET PATHS /usr/lib/x86_64-linux-gnu/cmake/opencv4 NO_DEFAULT_PATH)
    endif()
    if(NOT OpenCV_FOUND)
        # Try aarch64 path for ARM64 builds (Raspberry Pi)
        find_package(OpenCV QUIET PATHS /usr/lib/aarch64-linux-gnu/cmake/opencv4 NO_DEFAULT_PATH)
    endif()
    if(NOT OpenCV_FOUND)
        # Try opencv4 package name
        find_package(opencv4 QUIET)
        if(opencv4_FOUND)
            set(OpenCV_FOUND TRUE)
        endif()
    endif()
    if(NOT OpenCV_FOUND)
        message(FATAL_ERROR "OpenCV not found. Please install OpenCV development libraries:\n"
                "  Ubuntu/Debian (64-bit x86): sudo apt-get install libopencv-dev\n"
                "  macOS: brew install opencv")
    endif()
endif()
find_package(Threads REQUIRED)

# Include directories
include_directories(${CMAKE_CURRENT_SOURCE_DIR}/include)
include_directories(${OpenCV_INCLUDE_DIRS})

# Source files
set(SOURCES
    src/main.cpp
    src/config_manager.cpp
    src/logger.cpp
    src/graph.cpp
    src/detector.cpp
    src/rel_predictor.cpp
    src/runner.cpp
)

# Create executable
add_executable(cpp-scene-graph-detector ${SOURCES})

# Link libraries
target_link_libraries(cpp-scene-graph-detector 
    ${OpenCV_LIBS}
    Threads::Threads
)

# Add OpenCL support if enabled
if(SCENE_GRAPH_WITH_OPENCL)
    find_package(OpenCL)
    if(OpenCL_FOUND)
        target_compile_definitions(cpp-scene-graph-detector PRIVATE WITH_OPENCL)
        target_include_directories(cpp-scene-graph-detector PRIVATE ${OpenCL_INCLUDE_DIRS})
        target_link_libraries(cpp-scene-graph-detector ${OpenCL_LIBRARIES})
        message(STATUS "OpenCL support enabled")
    else()
        message(WARNING "OpenCL requested but not found, building without OpenCL support")
    endif()
endif()

# Compiler flags for optimization and warnings
if(APPLE)
    # macOS-specific optimizations
    target_compile_options(cpp-scene-graph-detector PRIVATE 
        -O3 
        -Wall 
        -Wextra 
        -pedantic
        -mtune=native  # Use mtune instead of march on macOS for better compatibility
    )
else()
    # Linux-specific optimizations
    target_compile_options(cpp-scene-graph-detector PRIVATE 
        -O3 
        -Wall 
        -Wextra 
        -pedantic
    )
    
    # Architecture-specific optimizations for Linux
    if(CMAKE_SYSTEM_PROCESSOR MATCHES "x86_64" OR CMAKE_SYSTEM_PROCESSOR MATCHES "amd64")
        message(STATUS "Building for x86_64/amd64 (${CMAKE_SYSTEM_PROCESSOR})")
        target_compile_options(cpp-scene-graph-detector PRIVATE -march=native)
    elseif(CMAKE_SYSTEM_PROCESSOR MATCHES "aarch64" OR CMAKE_SYSTEM_PROCESSOR MATCHES "arm64")
        message(STATUS "Building for ARM64/aarch64 (${CMAKE_SYSTEM_PROCESSOR})")
        target_compile_options(cpp-scene-graph-detector PRIVATE -mtune=cortex-a72)
    else()
        message(STATUS "Building for ${CMAKE_SYSTEM_PROCESSOR}")
        target_compile_options(cpp-scene-graph-detector PRIVATE -march=native)
    endif()
endif()

# Enable testing
enable_testing()

# Add test subdirectory if it exists and GTest is available
if(SCENE_GRAPH_BUILD_TESTS)
    find_package(GTest QUIET)
    if(EXISTS ${CMAKE_CURRENT_SOURCE_DIR}/tests AND GTest_FOUND)
        message(STATUS "Google Test found, building tests")
        add_subdirectory(tests)
    else()
        message(STATUS "Google Test not found or tests directory missing, skipping tests")
    endif()
endif()

# Installation
install(TARGETS cpp-scene-graph-detector DESTINATION bin)
