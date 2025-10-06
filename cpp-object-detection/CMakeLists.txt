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
        # Try x86_64 path first
        find_package(OpenCV QUIET PATHS /usr/lib/x86_64-linux-gnu/cmake/opencv4 NO_DEFAULT_PATH)
    endif()
    if(NOT OpenCV_FOUND)
        # Try i386 path for 32-bit builds
        find_package(OpenCV QUIET PATHS /usr/lib/i386-linux-gnu/cmake/opencv4 NO_DEFAULT_PATH)
    endif()
    if(NOT OpenCV_FOUND)
        # Try i686 path for 32-bit builds
        find_package(OpenCV QUIET PATHS /usr/lib/i686-linux-gnu/cmake/opencv4 NO_DEFAULT_PATH)
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
                "  Ubuntu/Debian (64-bit): sudo apt-get install libopencv-dev\n"
                "  Ubuntu/Debian (32-bit): sudo apt-get install libopencv-dev:i386\n"
                "  CentOS/RHEL: sudo yum install opencv-devel")
    endif()
endif()
find_package(Threads REQUIRED)

# Find CURL library
# Try using pkg-config first, then fall back to find_package
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
    include_directories(${CURL_INCLUDE_DIRS})
endif()

# Find SQLite3 library
find_package(PkgConfig)
if(PKG_CONFIG_FOUND)
    pkg_check_modules(SQLITE3 REQUIRED sqlite3)
    if(SQLITE3_FOUND)
        include_directories(${SQLITE3_INCLUDE_DIRS})
        link_directories(${SQLITE3_LIBRARY_DIRS})
    endif()
else()
    # Fall back to find_library
    find_library(SQLITE3_LIBRARY NAMES sqlite3)
    if(NOT SQLITE3_LIBRARY)
        message(FATAL_ERROR "SQLite3 not found. Please install sqlite3 development libraries:\n"
                "  Ubuntu/Debian: sudo apt-get install libsqlite3-dev\n"
                "  CentOS/RHEL: sudo yum install sqlite-devel\n"
                "  macOS: brew install sqlite3")
    endif()
endif()

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
    src/system_monitor.cpp
    src/google_sheets_client.cpp
    src/scene_manager.cpp
)

# Create executable
add_executable(object_detection ${SOURCES})

# Link libraries
if(PKG_CONFIG_FOUND AND CURL_FOUND)
    # Link using pkg-config results
    target_link_libraries(object_detection 
        ${OpenCV_LIBS}
        Threads::Threads
        ${CURL_LIBRARIES}
        ${SQLITE3_LIBRARIES}
    )
else()
    # Link using find_package results
    target_link_libraries(object_detection 
        ${OpenCV_LIBS}
        Threads::Threads
        CURL::libcurl
        ${SQLITE3_LIBRARY}
    )
endif()

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
    )
    
    # Architecture-specific optimizations for Linux
    if(CMAKE_SYSTEM_PROCESSOR MATCHES "i386" OR CMAKE_SYSTEM_PROCESSOR MATCHES "i686")
        # 32-bit x86 - use conservative optimizations for compatibility
        message(STATUS "Building for 32-bit x86 (${CMAKE_SYSTEM_PROCESSOR})")
        target_compile_options(object_detection PRIVATE -mtune=generic)
    else()
        # 64-bit or other architectures - use native optimizations
        message(STATUS "Building for ${CMAKE_SYSTEM_PROCESSOR}")
        target_compile_options(object_detection PRIVATE -march=native)
    endif()
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