cmake_minimum_required(VERSION 3.16)
project(ObjectDetection VERSION 1.0.0 LANGUAGES CXX)

# Set C++ standard
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Build configuration
set(CMAKE_BUILD_TYPE Release)

# Enable static linking for standalone executable
set(CMAKE_EXE_LINKER_FLAGS "-static-libgcc -static-libstdc++")

# Find required packages
find_package(OpenCV REQUIRED)
find_package(Threads REQUIRED)

# Include directories
include_directories(${CMAKE_CURRENT_SOURCE_DIR}/include)
include_directories(${OpenCV_INCLUDE_DIRS})

# Source files
set(SOURCES
    src/main.cpp
    src/webcam_interface.cpp
    src/object_detector.cpp
    src/config_manager.cpp
    src/logger.cpp
    src/performance_monitor.cpp
)

# Create executable
add_executable(object_detection ${SOURCES})

# Link libraries
target_link_libraries(object_detection 
    ${OpenCV_LIBS}
    Threads::Threads
)

# Compiler flags for optimization and warnings
target_compile_options(object_detection PRIVATE 
    -O3 
    -Wall 
    -Wextra 
    -pedantic
    -march=native
)

# Enable testing
enable_testing()

# Add test subdirectory if it exists
if(EXISTS ${CMAKE_CURRENT_SOURCE_DIR}/tests)
    add_subdirectory(tests)
endif()

# Installation
install(TARGETS object_detection DESTINATION bin)