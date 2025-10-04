# C++ Object Detection - Architecture & Component Interaction

## Table of Contents
1. [Overview](#overview)
2. [System Architecture](#system-architecture)
3. [Component Descriptions](#component-descriptions)
4. [Sequence Diagrams](#sequence-diagrams)
5. [State Machine & Transitions](#state-machine--transitions)
6. [Data Flow](#data-flow)
7. [Object Tracking Mechanism](#object-tracking-mechanism)
8. [Photo Storage Flow](#photo-storage-flow)

---

## Overview

The C++ Object Detection application is a real-time computer vision system designed for monitoring and detecting objects (people, vehicles, animals) from webcam feeds. The system features a pluggable model architecture, parallel processing capabilities, and automated photo storage with bounding box annotations.

**Key Capabilities:**
- Real-time object detection using YOLO models
- Pluggable model architecture (YOLOv5s, YOLOv5l, YOLOv8n, YOLOv8m)
- Parallel frame processing with configurable thread pool
- Object tracking for enter/exit event detection
- Automatic photo storage with timestamped filenames and bounding boxes
- Performance monitoring and adaptive frame rate control
- Headless operation for deployment on edge devices

---

## System Architecture

### High-Level Architecture Diagram

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                          C++ Object Detection System                          │
└──────────────────────────────────────────────────────────────────────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    │                 │                 │
                    ▼                 ▼                 ▼
        ┌───────────────────┐ ┌──────────────┐ ┌────────────────┐
        │  Input Subsystem  │ │  Processing  │ │ Output         │
        │                   │ │  Subsystem   │ │ Subsystem      │
        └───────────────────┘ └──────────────┘ └────────────────┘
                │                     │                 │
        ┌───────▼─────────┐   ┌──────▼─────────┐   ┌──▼──────────┐
        │ WebcamInterface │   │ ParallelFrame  │   │ Logger      │
        │                 │   │ Processor      │   │             │
        │ - Camera init   │   │                │   │ - Events    │
        │ - Frame capture │   │ - Thread pool  │   │ - Heartbeat │
        │ - Buffering     │   │ - Queue mgmt   │   │ - Metrics   │
        └─────────────────┘   └────────────────┘   └─────────────┘
                                      │
                        ┌─────────────┼─────────────┐
                        │             │             │
                        ▼             ▼             ▼
              ┌──────────────┐ ┌────────────┐ ┌─────────────┐
              │ Object       │ │ Detection  │ │ Performance │
              │ Detector     │ │ Model      │ │ Monitor     │
              │              │ │ Interface  │ │             │
              │ - Tracking   │ │            │ │ - FPS calc  │
              │ - Filtering  │ │ - YOLOv5s  │ │ - Warnings  │
              │ - Events     │ │ - YOLOv5l  │ │ - Reports   │
              └──────────────┘ │ - YOLOv8n  │ └─────────────┘
                               │ - YOLOv8m  │
                               └────────────┘
```

### Component Interaction Map

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Application Lifecycle                           │
└─────────────────────────────────────────────────────────────────────────┘

  main.cpp
     │
     ├─► setupSignalHandlers()
     │
     ├─► parseAndValidateConfig()
     │       │
     │       └─► ConfigManager
     │              ├─► Parse command line arguments
     │              └─► Validate configuration
     │
     ├─► initializeComponents()
     │       │
     │       ├─► Logger (logging system)
     │       ├─► PerformanceMonitor (metrics tracking)
     │       ├─► WebcamInterface (camera init)
     │       ├─► ObjectDetector (model loading)
     │       │      │
     │       │      └─► DetectionModelFactory
     │       │             └─► IDetectionModel (YOLOv5s/l/8n/m)
     │       │
     │       └─► ParallelFrameProcessor (worker threads)
     │
     ├─► runMainProcessingLoop()
     │       │
     │       └─► [Continuous loop while running = true]
     │              │
     │              ├─► WebcamInterface::captureFrame()
     │              ├─► ParallelFrameProcessor::submitFrame()
     │              ├─► Process completed frames
     │              └─► Periodic heartbeat logging
     │
     └─► performGracefulShutdown()
            │
            ├─► ParallelFrameProcessor::shutdown()
            ├─► Process remaining frames
            └─► WebcamInterface::release()
```

### Thread Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Thread Organization                              │
└─────────────────────────────────────────────────────────────────────────┘

┌──────────────────────┐        ┌────────────────────────────────────────┐
│   Main Thread        │        │      Worker Thread Pool                │
│                      │        │  (if parallel processing enabled)      │
│  - Signal handling   │        │                                        │
│  - Frame capture     │        │  Thread 1    Thread 2    Thread N      │
│  - Submit frames     │        │     │            │           │         │
│  - Collect results   │        │     │            │           │         │
│  - Heartbeat logging │        │     ▼            ▼           ▼         │
│                      │        │  ┌─────────────────────────────────┐   │
└──────────┬───────────┘        │  │  processFrameInternal()         │   │
           │                    │  │                                 │   │
           │  submit frame      │  │  1. ObjectDetector::detect()    │   │
           ├───────────────────►│  │  2. Filter target classes       │   │
           │                    │  │  3. Log detections              │   │
           │  get result        │  │  4. saveDetectionPhoto()        │   │
           ◄───────────────────┤  │     - Draw bounding boxes        │   │
           │                    │  │     - Add labels                │   │
           │                    │  │     - Save to disk              │   │
           │                    │  └─────────────────────────────────┘   │
           │                    └────────────────────────────────────────┘
           │
           ▼
    ┌──────────────────┐
    │  Mutex-Protected │
    │  Resources       │
    │                  │
    │  - Frame Queue   │
    │  - Photo Saving  │
    │  - Logging       │
    └──────────────────┘
```

---

## Component Descriptions

### 1. Main Application (`main.cpp`)

**Responsibilities:**
- Entry point for the application
- Signal handling (SIGINT, SIGTERM) for graceful shutdown
- High-level application lifecycle management

**Key Functions:**
- `setupSignalHandlers()`: Registers signal handlers for clean shutdown
- `parseAndValidateConfig()`: Processes command-line arguments
- `initializeComponents()`: Sets up all subsystems
- `runMainProcessingLoop()`: Main event loop
- `performGracefulShutdown()`: Cleanup and resource release

### 2. ApplicationContext (`application_context.hpp`)

**Responsibilities:**
- Centralized state container for all components
- Shared resource management

**Key Members:**
- `ConfigManager config_manager`: Configuration holder
- `Logger logger`: Logging subsystem
- `WebcamInterface webcam`: Camera interface
- `ObjectDetector detector`: Detection orchestrator
- `ParallelFrameProcessor frame_processor`: Processing pipeline
- `PerformanceMonitor perf_monitor`: Performance tracking
- `ViewfinderWindow viewfinder`: Real-time local preview (optional)
- `NetworkStreamer network_streamer`: Network video streaming (optional)

### 3. ConfigManager (`config_manager.hpp/cpp`)

**Responsibilities:**
- Command-line argument parsing
- Configuration validation
- Default value management

**Key Configuration Parameters:**
- Camera settings (ID, resolution, FPS)
- Model settings (path, type, confidence threshold)
- Processing settings (threads, queue size)
- Output settings (log file, output directory)

### 4. WebcamInterface (`webcam_interface.hpp/cpp`)

**Responsibilities:**
- USB camera initialization and management
- Frame capture from video device
- Camera capability detection

**Key Methods:**
- `initialize()`: Opens video device and configures camera
- `captureFrame()`: Retrieves next frame from camera
- `release()`: Closes camera and releases resources
- `getCameraInfo()`: Returns camera metadata

**States:**
```
[Not Initialized] ──initialize()──► [Ready] ──captureFrame()──► [Capturing]
                                       │                           │
                                       │◄──────────────────────────┘
                                       │
                                       └──release()──► [Released]
```

### 5. ObjectDetector (`object_detector.hpp/cpp`)

**Responsibilities:**
- Object detection orchestration
- Model management (loading, switching)
- Target class filtering
- Object tracking for enter/exit events

**Key Methods:**
- `initialize()`: Loads detection model via factory
- `detectObjects()`: Performs detection on frame
- `processFrame()`: Detects, filters, tracks, and logs
- `switchModel()`: Hot-swaps detection model
- `isTargetClass()`: Filters for target objects

**Target Classes:**
- People: `person`
- Vehicles: `car`, `truck`, `bus`, `motorcycle`, `bicycle`
- Animals: `cat`, `dog`

### 6. Detection Model Interface (`detection_model_interface.hpp`)

**Responsibilities:**
- Abstract interface for pluggable detection models
- Standardized API for all model implementations

**Key Methods:**
- `initialize()`: Loads and prepares model
- `detect()`: Runs inference on frame
- `getMetrics()`: Returns performance characteristics
- `warmUp()`: Pre-loads model for faster first inference

**Model Factory Pattern:**
```
DetectionModelFactory::createModel(ModelType)
                │
                ├─► YOLOv5s (fast, 75% accuracy, ~65ms)
                ├─► YOLOv5l (accurate, 85% accuracy, ~120ms)
                ├─► YOLOv8n (fastest, 70% accuracy, ~35ms)
                └─► YOLOv8m (best, 88% accuracy, ~150ms)
```

### 7. ParallelFrameProcessor (`parallel_frame_processor.hpp/cpp`)

**Responsibilities:**
- Manages worker thread pool for parallel processing
- Frame queue management
- Photo storage with bounding boxes
- Rate limiting for photo saves

**Key Methods:**
- `initialize()`: Creates worker threads and output directory
- `submitFrame()`: Queues frame for processing
- `processFrameInternal()`: Core processing logic
- `saveDetectionPhoto()`: Saves annotated frames to disk
- `generateFilename()`: Creates timestamped filenames

**Processing Modes:**
- **Sequential Mode** (num_threads = 1): Synchronous processing
- **Parallel Mode** (num_threads > 1): Multi-threaded queue processing

### 8. Logger (`logger.hpp/cpp`)

**Responsibilities:**
- Structured logging with timestamps
- Object detection event logging
- Performance reporting
- Heartbeat messages

**Log Levels:**
- `DEBUG`: Detailed diagnostic information
- `INFO`: General informational messages
- `WARNING`: Warning conditions
- `ERROR`: Error conditions

### 9. PerformanceMonitor (`performance_monitor.hpp/cpp`)

**Responsibilities:**
- Frame rate calculation
- Processing time tracking
- Performance warning system
- Periodic performance reports

**Metrics Tracked:**
- Frames per second (FPS)
- Average processing time per frame
- Total frames processed
- Performance warnings (when FPS drops below threshold)

### 10. ViewfinderWindow (`viewfinder_window.hpp/cpp`)

**Responsibilities:**
- Real-time display of detection results
- Drawing bounding boxes and labels
- User interaction (keyboard input for closing)

**Key Methods:**
- `initialize()`: Creates OpenCV window
- `showFrame()`: Displays frame with bounding boxes
- `shouldClose()`: Checks for user quit request
- `drawBoundingBoxes()`: Renders detection annotations

**States:**
```
[Not Initialized] ──initialize()──► [Ready] ──showFrame()──► [Displaying]
                                       │                        │
                                       │◄───────────────────────┘
                                       │
                                       └──close()──► [Closed]
```

### 11. NetworkStreamer (`network_streamer.hpp/cpp`)

**Responsibilities:**
- MJPEG HTTP streaming over network
- Multi-client connection handling
- Frame encoding to JPEG
- Network socket management

**Key Methods:**
- `initialize()`: Creates and binds TCP socket
- `start()`: Starts server thread for accepting connections
- `updateFrame()`: Updates current frame to stream
- `stop()`: Closes connections and stops server
- `getStreamingUrl()`: Returns HTTP URL for stream access

**Key Features:**
- Protocol: MJPEG over HTTP (multipart/x-mixed-replace)
- Port: Configurable (default 8080)
- Quality: 80% JPEG compression
- Frame Rate: ~10 fps (configurable)
- Accessibility: Compatible with browsers, VLC, ffplay

**States:**
```
[Not Initialized] ──initialize()──► [Ready] ──start()──► [Listening]
                                       │                     │
                                       │                     │ accept()
                                       │                     ▼
                                       │              [Client Connected]
                                       │                     │
                                       │                     │ stream frames
                                       │                     │
                                       │                     ▼
                                       │              [Client Disconnected]
                                       │                     │
                                       │◄────────────────────┘
                                       │
                                       └──stop()──► [Stopped]
```

**Network Flow:**
```
Client Device          NetworkStreamer           Main Loop
(Browser/VLC)               │                        │
    │                       │                        │
    │  HTTP GET /stream     │                        │
    ├──────────────────────►│                        │
    │                       │                        │
    │  HTTP 200 OK          │                        │
    │  (MJPEG headers)      │                        │
    │◄──────────────────────┤                        │
    │                       │                        │
    │                       │  updateFrame()         │
    │                       │◄───────────────────────┤
    │                       │  (frame + detections)  │
    │                       │                        │
    │  JPEG frame 1         │                        │
    │  (with bboxes)        │                        │
    │◄──────────────────────┤                        │
    │                       │                        │
    │  JPEG frame 2         │                        │
    │◄──────────────────────┤                        │
    │                       │                        │
    │        ...            │         ...            │
```

**Security Note:**
- No authentication/encryption (designed for local network only)
- Binds to all interfaces (0.0.0.0)
- Should not be exposed to public internet

---

## Sequence Diagrams

### Application Startup Sequence

```
User          main.cpp       ConfigMgr    Logger    Webcam    Detector    FrameProc
 │               │              │           │         │          │            │
 │   execute     │              │           │         │          │            │
 ├──────────────►│              │           │         │          │            │
 │               │              │           │         │          │            │
 │               │  parseArgs() │           │         │          │            │
 │               ├─────────────►│           │         │          │            │
 │               │◄─────────────┤           │         │          │            │
 │               │   Config     │           │         │          │            │
 │               │              │           │         │          │            │
 │               │  create      │           │         │          │            │
 │               ├──────────────┼──────────►│         │          │            │
 │               │              │           │  info() │          │            │
 │               │              │           │◄────────┤          │            │
 │               │              │           │         │          │            │
 │               │  initialize()│           │         │          │            │
 │               ├──────────────┼───────────┼────────►│          │            │
 │               │              │           │         │  open()  │            │
 │               │              │           │         │  camera  │            │
 │               │              │           │         │◄─────────┤            │
 │               │              │           │         │          │            │
 │               │  create Detector        │         │          │            │
 │               ├──────────────┼───────────┼─────────┼─────────►│            │
 │               │              │           │         │          │ loadModel()│
 │               │              │           │         │          │◄───────────┤
 │               │              │           │         │          │            │
 │               │  create FrameProcessor  │         │          │            │
 │               ├──────────────┼───────────┼─────────┼──────────┼───────────►│
 │               │              │           │         │          │            │
 │               │              │           │         │          │      start │
 │               │              │           │         │          │      workers
 │               │              │           │         │          │            │
 │               │              │           │  ready  │          │            │
 │               │◄─────────────┴───────────┴─────────┴──────────┴────────────┤
 │               │              │           │         │          │            │
 │   running     │              │           │         │          │            │
 │◄──────────────┤              │           │         │          │            │
 │               │              │           │         │          │            │
```

### Frame Processing Sequence (Parallel Mode)

```
MainLoop     Webcam    FrameProc   WorkerThread  Detector   PhotoStorage  Logger
   │           │           │            │           │            │          │
   │ capture   │           │            │           │            │          │
   ├──────────►│           │            │           │            │          │
   │           │  frame    │            │           │            │          │
   │◄──────────┤           │            │           │            │          │
   │           │           │            │           │            │          │
   │ submit    │           │            │           │            │          │
   ├───────────┼──────────►│            │           │            │          │
   │           │           │  queue     │           │            │          │
   │           │           │  frame     │           │            │          │
   │           │           ├───────────►│           │            │          │
   │  future   │           │            │           │            │          │
   │◄──────────┼───────────┤            │           │            │          │
   │           │           │            │           │            │          │
   │           │           │            │  dequeue  │            │          │
   │           │           │            │◄──────────┤            │          │
   │           │           │            │           │            │          │
   │           │           │            │  detect() │            │          │
   │           │           │            ├──────────►│            │          │
   │           │           │            │           │  inference │          │
   │           │           │            │           │  (YOLO)    │          │
   │           │           │            │◄──────────┤            │          │
   │           │           │            │ detections│            │          │
   │           │           │            │           │            │          │
   │           │           │            │  filter target classes │          │
   │           │           │            │           │            │          │
   │           │           │            │  log detections        │          │
   │           │           │            ├───────────┼────────────┼─────────►│
   │           │           │            │           │            │          │
   │           │           │            │ save photo?            │          │
   │           │           │            │ (check rate limit)     │          │
   │           │           │            ├───────────────────────►│          │
   │           │           │            │           │            │          │
   │           │           │            │           │     [if allowed]      │
   │           │           │            │           │            │          │
   │           │           │            │           │    draw bounding boxes│
   │           │           │            │           │    add labels         │
   │           │           │            │           │    generate filename  │
   │           │           │            │           │    save to disk       │
   │           │           │            │           │            │          │
   │           │           │            │           │        log success    │
   │           │           │            │           │            ├─────────►│
   │           │           │            │           │            │          │
   │           │           │  result    │           │            │          │
   │           │           │◄───────────┤           │            │          │
   │           │           │            │           │            │          │
   │  get      │           │            │           │            │          │
   │  result   │           │            │           │            │          │
   ├───────────┼──────────►│            │           │            │          │
   │◄──────────┼───────────┤            │           │            │          │
   │ detections│           │            │           │            │          │
   │           │           │            │           │            │          │
```

### Object Tracking & Event Logging Sequence

```
Detector    TrackedObjs   CurrentFrame   Logger
   │             │              │           │
   │  detect     │              │           │
   ├─────────────┼─────────────►│           │
   │             │              │           │
   │◄────────────┼──────────────┤           │
   │  detections │              │           │
   │             │              │           │
   │  update tracking           │           │
   ├────────────►│              │           │
   │             │              │           │
   │             │ for each detection       │
   │             │ find matching tracked obj│
   │             │              │           │
   │             │ [if match found]         │
   │             │   update center          │
   │             │   reset frame counter    │
   │             │              │           │
   │             │ [if new object]          │
   │             │   create tracker         │
   │             │   mark as present        │
   │             │              │           │
   │◄────────────┤              │           │
   │             │              │           │
   │  log events │              │           │
   ├─────────────┼──────────────┼──────────►│
   │             │              │           │
   │             │  [for each tracked obj]  │
   │             │              │           │
   │             │  was absent >5 frames?   │
   │             │  now present?            │
   │             │              │           │
   │             │   [YES = ENTER EVENT]    │
   │             │              │  log      │
   │             │              │  "entered"│
   │             │              │           │
   │             │              │           │
   │             │  remove stale objects    │
   │             │  (>30 frames absent)     │
   │             │              │           │
```

### Photo Storage with Rate Limiting Sequence

```
FrameProc   PhotoMutex   RateLimit    Annotator   Filesystem   Logger
   │            │            │            │             │          │
   │  save photo?           │            │             │          │
   ├───────────►│            │            │             │          │
   │            │  lock      │            │             │          │
   │            │◄───────────┤            │             │          │
   │            │            │            │             │          │
   │            │  check elapsed time     │             │          │
   │            ├───────────►│            │             │          │
   │            │            │            │             │          │
   │            │  [< 10 seconds elapsed] │             │          │
   │            │  SKIP      │            │             │          │
   │            │◄───────────┤            │             │          │
   │            │  unlock    │            │             │          │
   │◄───────────┤            │            │             │          │
   │  aborted   │            │            │             │          │
   │            │            │            │             │          │
   │  OR        │            │            │             │          │
   │            │            │            │             │          │
   │            │  [>= 10 seconds elapsed]│             │          │
   │            │  PROCEED   │            │             │          │
   │            │◄───────────┤            │             │          │
   │            │            │            │             │          │
   │            │  update last_photo_time │             │          │
   │            │            │            │             │          │
   │            │  annotate frame         │             │          │
   │            ├────────────┼───────────►│             │          │
   │            │            │            │  clone frame│          │
   │            │            │            │             │          │
   │            │            │      for each detection  │          │
   │            │            │        getColorForClass()│          │
   │            │            │        draw rectangle    │          │
   │            │            │        draw label bg     │          │
   │            │            │        draw label text   │          │
   │            │            │            │             │          │
   │            │            │◄───────────┤             │          │
   │            │            │ annotated  │             │          │
   │            │            │            │             │          │
   │            │  generate filename      │             │          │
   │            │  timestamp + objects    │             │          │
   │            │            │            │             │          │
   │            │  write to disk          │             │          │
   │            ├────────────┼────────────┼────────────►│          │
   │            │            │            │             │  imwrite()│
   │            │            │            │             │          │
   │            │            │            │            success     │
   │            │            │            │             ├─────────►│
   │            │            │            │             │  log     │
   │            │            │            │             │          │
   │            │  unlock    │            │             │          │
   │◄───────────┤            │            │             │          │
   │  saved     │            │            │             │          │
   │            │            │            │             │          │
```

---

## State Machine & Transitions

### Application State Machine

```
┌──────────────┐
│  STARTING    │  Initial state
└──────┬───────┘
       │
       │ parse config
       │ validate params
       │
       ▼
┌──────────────┐
│ INITIALIZING │  Loading components
└──────┬───────┘
       │
       │ create logger
       │ create webcam
       │ load detection model
       │ start worker threads
       │
       ▼
┌──────────────┐
│   RUNNING    │◄──────────┐  Main processing loop
└──────┬───────┘           │
       │                   │
       │ capture frame     │
       │ submit for detect │
       │ process results   │
       │ log events        │
       │                   │
       └───────────────────┘
       │
       │ signal received (SIGINT/SIGTERM)
       │
       ▼
┌──────────────┐
│  SHUTTING    │  Graceful cleanup
│   DOWN       │
└──────┬───────┘
       │
       │ stop workers
       │ process remaining frames
       │ release camera
       │ flush logs
       │
       ▼
┌──────────────┐
│   STOPPED    │  Final state
└──────────────┘
```

### Frame Processing State Machine

```
┌─────────────┐
│   IDLE      │  Waiting for frame
└──────┬──────┘
       │
       │ frame captured
       │
       ▼
┌─────────────┐
│  SUBMITTED  │  Frame in queue
└──────┬──────┘
       │
       │ worker available
       │
       ▼
┌─────────────┐
│ DETECTING   │  Running YOLO inference
└──────┬──────┘
       │
       │ detections ready
       │
       ▼
┌─────────────┐
│ FILTERING   │  Filter target classes
└──────┬──────┘
       │
       │ targets identified
       │
       ▼
┌─────────────┐
│  LOGGING    │  Log detection events
└──────┬──────┘
       │
       │ check photo save conditions
       │
       ├─────────►[NO targets]───────┐
       │                             │
       │ [targets present]           │
       │                             │
       ▼                             │
┌─────────────┐                      │
│ PHOTO_CHECK │                      │
└──────┬──────┘                      │
       │                             │
       ├►[<10s elapsed]──────────────┤
       │                             │
       │ [>=10s elapsed]             │
       │                             │
       ▼                             │
┌─────────────┐                      │
│  ANNOTATING │  Draw bounding boxes │
└──────┬──────┘                      │
       │                             │
       │ annotations complete        │
       │                             │
       ▼                             │
┌─────────────┐                      │
│   SAVING    │  Write to disk       │
└──────┬──────┘                      │
       │                             │
       │ save complete               │
       │                             │
       ▼                             │
┌─────────────┐                      │
│  COMPLETED  │◄─────────────────────┘
└─────────────┘
```

### Object Tracking State Machine (per object)

```
┌─────────────┐
│   NEW       │  Object just detected
└──────┬──────┘
       │
       │ create tracker
       │ record center position
       │ frames_since_detection = 0
       │
       ▼
┌─────────────┐
│   PRESENT   │  Object currently visible
└──────┬──────┘
       │
       │                          ┌─────────────────┐
       ├─►[detected this frame]──►│ UPDATE          │
       │                          │ - update center │
       │                          │ - reset counter │
       │                          └────────┬────────┘
       │                                   │
       │                                   │
       │                                   ▼
       │                          ┌─────────────────┐
       │                          │  CONTINUE       │
       │                          │  PRESENT        │
       │                          └────────┬────────┘
       │                                   │
       │                                   │
       │◄──────────────────────────────────┘
       │
       │
       │ [not detected this frame]
       │ increment frames_since_detection
       │
       ▼
┌─────────────┐
│   ABSENT    │  Temporarily not visible
└──────┬──────┘
       │
       ├─►[frames_since_detection <= 5]─────┐
       │                                     │
       │ [frames_since_detection > 5]       │
       │                                     │
       ▼                                     │
┌─────────────┐                              │
│   EXIT      │  Log exit event              │
└──────┬──────┘                              │
       │                                     │
       ├─►[frames_since_detection <= 30]────┤
       │                                     │
       │ [frames_since_detection > 30]      │
       │                                     │
       ▼                                     │
┌─────────────┐                              │
│   REMOVED   │  Delete tracker              │
└─────────────┘                              │
                                             │
       ┌─────────────────────────────────────┘
       │
       │ [detected again]
       │
       ▼
┌─────────────┐
│   ENTER     │  Log enter event
└──────┬──────┘
       │
       │ move to PRESENT
       │
       ▼
  [back to PRESENT state]
```

### Photo Storage Decision Tree

```
                    ┌───────────────────┐
                    │  Frame Processed  │
                    │  with Detections  │
                    └─────────┬─────────┘
                              │
                              ▼
                    ┌───────────────────┐
                    │ Any target        │
                    │ objects detected? │
                    └─────────┬─────────┘
                              │
                    ┌─────────┴─────────┐
                    │                   │
                   NO                  YES
                    │                   │
                    ▼                   ▼
            ┌──────────────┐   ┌──────────────────┐
            │   SKIP       │   │  Acquire photo   │
            │   Photo      │   │  mutex (lock)    │
            └──────────────┘   └────────┬─────────┘
                                        │
                                        ▼
                               ┌──────────────────┐
                               │ Calculate time   │
                               │ since last photo │
                               └────────┬─────────┘
                                        │
                          ┌─────────────┴─────────────┐
                          │                           │
                    < 10 seconds                >= 10 seconds
                          │                           │
                          ▼                           ▼
                 ┌──────────────────┐       ┌──────────────────┐
                 │  RATE LIMITED    │       │  Update          │
                 │  Skip this photo │       │  last_photo_time │
                 │  Unlock mutex    │       └────────┬─────────┘
                 └──────────────────┘                │
                                                     ▼
                                            ┌──────────────────┐
                                            │  Clone frame     │
                                            └────────┬─────────┘
                                                     │
                                                     ▼
                                            ┌──────────────────┐
                                            │  For each        │
                                            │  detection:      │
                                            │  - Get color     │
                                            │  - Draw bbox     │
                                            │  - Draw label    │
                                            └────────┬─────────┘
                                                     │
                                                     ▼
                                            ┌──────────────────┐
                                            │  Generate        │
                                            │  filename with   │
                                            │  timestamp +     │
                                            │  object types    │
                                            └────────┬─────────┘
                                                     │
                                                     ▼
                                            ┌──────────────────┐
                                            │  cv::imwrite()   │
                                            │  Save to disk    │
                                            └────────┬─────────┘
                                                     │
                                                     ▼
                                            ┌──────────────────┐
                                            │  Log success/    │
                                            │  failure         │
                                            │  Unlock mutex    │
                                            └──────────────────┘
```

---

## Data Flow

### End-to-End Data Flow

```
┌──────────────┐
│  USB Webcam  │
│   Hardware   │
└──────┬───────┘
       │
       │ video frames (raw pixels)
       │
       ▼
┌──────────────────────────────────────────────────────────┐
│  WebcamInterface                                         │
│  - OpenCV VideoCapture                                   │
│  - Resolution: 1280x720 (configurable)                   │
│  - Format: BGR (OpenCV default)                          │
└──────┬───────────────────────────────────────────────────┘
       │
       │ cv::Mat frame (BGR image)
       │
       ▼
┌──────────────────────────────────────────────────────────┐
│  ParallelFrameProcessor                                  │
│  - Queue management                                      │
│  - Thread dispatching                                    │
└──────┬───────────────────────────────────────────────────┘
       │
       │ cv::Mat frame (cloned)
       │
       ▼
┌──────────────────────────────────────────────────────────┐
│  Worker Thread                                           │
│  - processFrameInternal()                                │
└──────┬───────────────────────────────────────────────────┘
       │
       │ cv::Mat frame
       │
       ▼
┌──────────────────────────────────────────────────────────┐
│  ObjectDetector                                          │
│  - detectObjects()                                       │
└──────┬───────────────────────────────────────────────────┘
       │
       │ cv::Mat frame
       │
       ▼
┌──────────────────────────────────────────────────────────┐
│  IDetectionModel (YOLO implementation)                   │
│  - Preprocessing: resize, normalize, blob                │
│  - Inference: neural network forward pass                │
│  - Postprocessing: NMS, confidence filtering             │
└──────┬───────────────────────────────────────────────────┘
       │
       │ std::vector<Detection>
       │   struct Detection {
       │     std::string class_name;
       │     float confidence;
       │     cv::Rect bbox;
       │   }
       │
       ▼
┌──────────────────────────────────────────────────────────┐
│  ObjectDetector (filter)                                 │
│  - isTargetClass() for each detection                    │
└──────┬───────────────────────────────────────────────────┘
       │
       │ std::vector<Detection> (filtered)
       │
       ├──────────────────┬────────────────┐
       │                  │                │
       ▼                  ▼                ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────────┐
│  Logger      │  │  Object      │  │  Photo Storage   │
│  - Log coords│  │  Tracker     │  │  (if allowed)    │
│  - Log events│  │  - Update    │  │                  │
└──────────────┘  │    tracking  │  └──────┬───────────┘
                  │  - Log enter │         │
                  │    /exit     │         │
                  └──────────────┘         │
                                           ▼
                                  ┌──────────────────┐
                                  │  Annotate Frame  │
                                  │  - Draw bboxes   │
                                  │  - Add labels    │
                                  └──────┬───────────┘
                                         │
                                         │ cv::Mat (annotated)
                                         │
                                         ▼
                                  ┌──────────────────┐
                                  │  Filesystem      │
                                  │  - JPEG file     │
                                  │  - Filename:     │
                                  │    YYYY-MM-DD    │
                                  │    HHMMSS        │
                                  │    [objs]        │
                                  │    detected.jpg  │
                                  └──────────────────┘
```

### Detection Data Structure

```cpp
struct Detection {
    std::string class_name;      // e.g., "person", "car", "cat"
    float confidence;            // e.g., 0.92 (92%)
    cv::Rect bbox;              // Bounding box (x, y, width, height)
};

// Example after detection:
Detection {
    class_name = "person",
    confidence = 0.92,
    bbox = cv::Rect(320, 180, 200, 400)  // x=320, y=180, w=200, h=400
}

// Used to calculate center:
center_x = bbox.x + bbox.width / 2   // 320 + 100 = 420
center_y = bbox.y + bbox.height / 2  // 180 + 200 = 380
```

### Tracking Data Structure

```cpp
struct ObjectTracker {
    std::string object_type;              // "person", "car", etc.
    cv::Point2f center;                   // (x, y) center position
    bool was_present_last_frame;          // tracking flag
    int frames_since_detection;           // counter for staleness
};

// Example tracker state:
ObjectTracker {
    object_type = "person",
    center = cv::Point2f(420.0, 380.0),
    was_present_last_frame = true,
    frames_since_detection = 0
}
```

---

## Object Tracking Mechanism

### Tracking Algorithm Overview

The object tracking system uses a **simple centroid-based tracking** approach with temporal filtering:

1. **Centroid Calculation**: For each detection, calculate the center point of the bounding box
2. **Matching**: Compare current detections with previously tracked objects by distance
3. **Update or Create**: If a match is found (distance < 100 pixels), update the tracker; otherwise, create a new tracker
4. **Temporal Filtering**: Track how long an object has been absent to determine enter/exit events
5. **Cleanup**: Remove trackers for objects absent for >30 frames

### Tracking State Transitions

```
Frame N: person detected at (640, 360)
  ├─► Create new tracker
  │   - object_type = "person"
  │   - center = (640, 360)
  │   - frames_since_detection = 0
  │   - was_present_last_frame = true
  │
Frame N+1: person detected at (642, 362)
  ├─► Find match (distance = 2.8 pixels < 100)
  │   - Update center = (642, 362)
  │   - frames_since_detection = 0
  │   - was_present_last_frame = true
  │
Frame N+2: person NOT detected
  ├─► No match found
  │   - frames_since_detection = 1
  │   - was_present_last_frame = false
  │
Frame N+3 to N+5: person NOT detected
  ├─► Incrementing counter
  │   - frames_since_detection = 2, 3, 4
  │
Frame N+6: person detected at (645, 365)
  ├─► Find match (distance still acceptable)
  │   - frames_since_detection was > 5
  │   - Log "person entered frame"  ◄── ENTER EVENT
  │   - Update center = (645, 365)
  │   - frames_since_detection = 0
  │
Frame N+7 to N+36: person NOT detected
  ├─► Incrementing counter to 30
  │
Frame N+37: person still NOT detected
  ├─► frames_since_detection = 31 > 30
  │   - Remove tracker from list  ◄── CLEANUP
```

### Distance Calculation

```cpp
// Simple Euclidean distance between centers
cv::Point2f detection_center(
    detection.bbox.x + detection.bbox.width / 2.0f,
    detection.bbox.y + detection.bbox.height / 2.0f
);

float distance = cv::norm(tracked.center - detection_center);

// Example:
// Tracked center: (640, 360)
// Detection center: (642, 362)
// distance = sqrt((642-640)^2 + (362-360)^2)
//          = sqrt(4 + 4)
//          = 2.83 pixels
// Since 2.83 < 100, this is a match!
```

### Tracking Limitations & Design Decisions

**Current Implementation:**
- **Simple centroid matching**: Fast but may struggle with crossing objects
- **Fixed distance threshold**: 100 pixels works for most cases
- **No occlusion handling**: Objects temporarily hidden are treated as absent
- **No trajectory prediction**: Only uses current and past positions

**Design Rationale:**
- Optimized for **performance** over perfect accuracy
- Suitable for **security monitoring** where brief tracking gaps are acceptable
- **Lightweight**: No heavy ML tracking algorithms (e.g., SORT, DeepSORT)
- **Deterministic**: Predictable behavior for debugging

**When Tracking Works Well:**
- Single object in frame
- Objects moving at consistent speeds
- Clear separation between objects
- Good lighting conditions

**When Tracking May Struggle:**
- Multiple objects crossing paths
- Very fast moving objects
- Occlusions (object behind obstacle)
- Extreme lighting changes

---

## Photo Storage Flow

### When Photos Are Stored

Photos are saved **automatically** when:
1. ✅ At least one **target object** is detected (person, vehicle, or animal)
2. ✅ **Rate limit** has passed (>= 10 seconds since last photo)
3. ✅ **Thread-safe** access to photo saving (mutex acquired)

Photos are **NOT saved** when:
- ❌ No target objects detected
- ❌ Less than 10 seconds elapsed since last photo
- ❌ Frame processing failed

### Bounding Box Drawing Process

```
Step 1: Clone Frame
  ├─► Create copy of original frame to preserve original
  └─► cv::Mat annotated_frame = frame.clone();

Step 2: For Each Detection
  ├─► Get color based on object class
  │   └─► cv::Scalar color = getColorForClass(detection.class_name);
  │
  ├─► Draw bounding box rectangle
  │   └─► cv::rectangle(annotated_frame, detection.bbox, color, 2);
  │
  ├─► Create label text
  │   └─► label = "person 92%"
  │
  ├─► Calculate text size
  │   └─► cv::getTextSize(label, font, 0.5, 1, &baseline);
  │
  ├─► Draw label background (filled rectangle)
  │   └─► cv::rectangle(annotated_frame, bg_rect, color, cv::FILLED);
  │
  └─► Draw label text
      └─► cv::putText(annotated_frame, label, origin, font, 0.5, BLACK, 1);

Step 3: Save Annotated Frame
  └─► cv::imwrite(filepath, annotated_frame);
```

### Color Mapping

```cpp
Object Class        Color       BGR Values      Visual
─────────────────────────────────────────────────────────
person              Green       (0, 255, 0)     ████ 
cat                 Red         (0, 0, 255)     ████
dog                 Blue        (255, 0, 0)     ████
car/truck/bus       Yellow      (0, 255, 255)   ████
motorcycle/bicycle  Magenta     (255, 0, 255)   ████
other               White       (255, 255, 255) ████
```

### Filename Generation

```
Format: YYYY-MM-DD HHMMSS [object1] [object2] ... detected.jpg

Examples:
  2025-10-04 140020 person detected.jpg
  2025-10-04 140035 person cat detected.jpg
  2025-10-04 140050 car truck detected.jpg

Generation Process:
  1. Get current timestamp
     └─► std::chrono::system_clock::now()
  
  2. Format timestamp
     └─► std::put_time(&tm, "%Y-%m-%d %H%M%S")
  
  3. Collect unique object types (sorted)
     └─► std::set<std::string> object_types
  
  4. Build object string
     └─► "person cat"
  
  5. Combine: timestamp + " " + objects + " detected.jpg"
```

### Rate Limiting Implementation

```cpp
// Prevent saving photos too frequently (10 second minimum interval)
const int PHOTO_INTERVAL_SECONDS = 10;

void saveDetectionPhoto(...) {
    std::lock_guard<std::mutex> lock(photo_mutex_);  // Thread-safe
    
    auto now = std::chrono::steady_clock::now();
    auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(
        now - last_photo_time_
    );
    
    if (elapsed.count() < PHOTO_INTERVAL_SECONDS) {
        return;  // Skip - too soon since last photo
    }
    
    last_photo_time_ = now;  // Update for next check
    
    // Proceed with photo saving...
}
```

**Rate Limiting Benefits:**
- ✅ Prevents disk space exhaustion
- ✅ Reduces I/O load on system
- ✅ Makes photo review more manageable
- ✅ Thread-safe with mutex protection

---

## Summary

This architecture document provides a comprehensive view of the C++ Object Detection system, including:

- **Component organization**: Modular design with clear separation of concerns
- **Data flow**: From camera capture to photo storage
- **State machines**: Application lifecycle, frame processing, object tracking
- **Sequence diagrams**: Key interaction patterns
- **Implementation details**: Tracking algorithm, photo storage, bounding boxes

The system is designed for **real-time performance**, **extensibility** (pluggable models), **reliability** (thread-safe, error handling), and **ease of deployment** (headless operation, configurable).

For feature-specific details, see:
- [PHOTO_STORAGE_FEATURE.md](PHOTO_STORAGE_FEATURE.md) - Photo storage implementation
- [README.md](README.md) - Usage and deployment guide
- [DEPLOYMENT.md](DEPLOYMENT.md) - Deployment instructions
