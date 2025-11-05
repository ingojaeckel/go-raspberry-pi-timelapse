# Scene Graph JSON Format Specification

This document defines the JSON schema for scene graphs produced by the cpp-scene-graph-detector.

## Overview

A scene graph is a structured representation of objects and their spatial relationships in an image or video frame. The JSON format consists of three main sections:

1. **meta** - Metadata about the image/frame
2. **objects** - List of detected objects (nodes in the graph)
3. **relations** - List of spatial relationships (edges in the graph)

## JSON Schema

```json
{
  "meta": {
    "timestamp": "<ISO 8601 timestamp or formatted string>",
    "image_width": "<pixel width as string>",
    "image_height": "<pixel height as string>",
    "<custom_key>": "<custom_value>"
  },
  "objects": [
    {
      "id": <integer>,
      "class_id": <integer>,
      "label": "<string>",
      "score": <float 0-1>,
      "bbox": {
        "x": <float 0-1>,
        "y": <float 0-1>,
        "width": <float 0-1>,
        "height": <float 0-1>
      }
    }
  ],
  "relations": [
    {
      "subject_id": <integer>,
      "predicate": "<string>",
      "object_id": <integer>,
      "score": <float 0-1>
    }
  ]
}
```

## Field Descriptions

### Meta Section

Optional metadata about the scene/frame:

| Field | Type | Description |
|-------|------|-------------|
| `timestamp` | string | When the frame was processed |
| `image_width` | string | Image width in pixels |
| `image_height` | string | Image height in pixels |
| (custom) | string | Any additional metadata |

### Objects Section

Array of detected objects, each containing:

| Field | Type | Range | Description |
|-------|------|-------|-------------|
| `id` | integer | ≥0 | Unique identifier within this frame |
| `class_id` | integer | ≥0 | Object class ID from model |
| `label` | string | - | Human-readable class name (e.g., "person", "car") |
| `score` | float | 0-1 | Detection confidence score |
| `bbox.x` | float | 0-1 | Center X coordinate (normalized) |
| `bbox.y` | float | 0-1 | Center Y coordinate (normalized) |
| `bbox.width` | float | 0-1 | Bounding box width (normalized) |
| `bbox.height` | float | 0-1 | Bounding box height (normalized) |

**Bounding Box Coordinates:**
- Coordinates are normalized to [0, 1] range
- `(x, y)` represents the **center** of the bounding box
- To convert to pixel coordinates: multiply by image dimensions
- To get corners: `left = x - width/2`, `right = x + width/2`, etc.

### Relations Section

Array of spatial relationships between objects:

| Field | Type | Range | Description |
|-------|------|-------|-------------|
| `subject_id` | integer | ≥0 | ID of subject object (from objects array) |
| `predicate` | string | - | Spatial relationship type |
| `object_id` | integer | ≥0 | ID of object (from objects array) |
| `score` | float | 0-1 | Relation confidence (1.0 for geometric) |

**Supported Predicates:**
- `left_of` - Subject is to the left of object
- `right_of` - Subject is to the right of object
- `on` - Subject is on top of object
- `under` - Subject is below object
- `overlaps` - Subject and object overlap
- `contains` - Subject contains object
- `next_to` - Subject is adjacent to object
- `in_front_of` - Subject is in front (future)
- `behind` - Subject is behind (future)
- `between` - Subject is between two objects (future)

## Example: Simple Scene

```json
{
  "meta": {
    "timestamp": "2024-01-15 10:30:45",
    "image_width": "1280",
    "image_height": "720"
  },
  "objects": [
    {
      "id": 0,
      "class_id": 0,
      "label": "person",
      "score": 0.92,
      "bbox": {
        "x": 0.45,
        "y": 0.55,
        "width": 0.15,
        "height": 0.35
      }
    },
    {
      "id": 1,
      "class_id": 2,
      "label": "car",
      "score": 0.88,
      "bbox": {
        "x": 0.75,
        "y": 0.60,
        "width": 0.30,
        "height": 0.25
      }
    }
  ],
  "relations": [
    {
      "subject_id": 0,
      "predicate": "left_of",
      "object_id": 1,
      "score": 1.0
    }
  ]
}
```

**Interpretation:**
- A person (92% confidence) is detected in the left-center of the frame
- A car (88% confidence) is detected in the right-center of the frame  
- The person is to the left of the car

## Example: Complex Scene

```json
{
  "meta": {
    "timestamp": "2024-01-15 14:22:10",
    "image_width": "1920",
    "image_height": "1080"
  },
  "objects": [
    {
      "id": 0,
      "class_id": 14,
      "label": "house",
      "score": 0.95,
      "bbox": {
        "x": 0.30,
        "y": 0.40,
        "width": 0.25,
        "height": 0.40
      }
    },
    {
      "id": 1,
      "class_id": 14,
      "label": "house",
      "score": 0.91,
      "bbox": {
        "x": 0.70,
        "y": 0.42,
        "width": 0.22,
        "height": 0.38
      }
    },
    {
      "id": 2,
      "class_id": 15,
      "label": "tree",
      "score": 0.87,
      "bbox": {
        "x": 0.50,
        "y": 0.35,
        "width": 0.12,
        "height": 0.45
      }
    },
    {
      "id": 3,
      "class_id": 16,
      "label": "bush",
      "score": 0.76,
      "bbox": {
        "x": 0.15,
        "y": 0.70,
        "width": 0.10,
        "height": 0.15
      }
    }
  ],
  "relations": [
    {
      "subject_id": 2,
      "predicate": "between",
      "object_id": 0,
      "score": 1.0
    },
    {
      "subject_id": 2,
      "predicate": "between",
      "object_id": 1,
      "score": 1.0
    },
    {
      "subject_id": 0,
      "predicate": "left_of",
      "object_id": 1,
      "score": 1.0
    },
    {
      "subject_id": 3,
      "predicate": "left_of",
      "object_id": 0,
      "score": 1.0
    }
  ]
}
```

**Interpretation:**
- Two houses (left and right)
- One tree in the middle
- One bush in the lower left
- The tree is between the two houses
- The first house is left of the second house
- The bush is left of the first house

## Coordinate System

The bounding box uses a **normalized center-based** coordinate system:

```
(0,0) ────────────────────> X (1,0)
  │
  │         ┌─────────┐
  │         │  width  │
  │    ─────┼────●────┼─────  ← center (x, y)
  │         │ height  │
  │         └─────────┘
  │
  v Y
(0,1)
```

**Example Conversion to Pixel Coordinates:**

```python
# Given normalized bbox and image size
image_width = 1280
image_height = 720
bbox = {"x": 0.5, "y": 0.5, "width": 0.2, "height": 0.3}

# Calculate pixel coordinates
center_x_px = bbox["x"] * image_width          # 640
center_y_px = bbox["y"] * image_height         # 360
width_px = bbox["width"] * image_width         # 256
height_px = bbox["height"] * image_height      # 216

# Calculate corners
left = center_x_px - width_px / 2              # 512
right = center_x_px + width_px / 2             # 768
top = center_y_px - height_px / 2              # 252
bottom = center_y_px + height_px / 2           # 468
```

## Validation

A valid scene graph JSON must satisfy:

1. **Object IDs are unique** within the objects array
2. **Relation references are valid**: `subject_id` and `object_id` must exist in objects array
3. **Scores are in range**: All score values must be between 0.0 and 1.0
4. **Bounding boxes are normalized**: All bbox coordinates must be between 0.0 and 1.0
5. **Labels are non-empty**: Object and predicate labels must be valid strings

## Use Cases

### 1. Visual Question Answering
Query: "Is there a person next to a car?"
```python
for relation in scene_graph["relations"]:
    if relation["predicate"] == "next_to":
        subject = find_object(relation["subject_id"])
        object = find_object(relation["object_id"])
        if subject["label"] == "person" and object["label"] == "car":
            return True
```

### 2. Scene Understanding
Generate natural language description:
```python
description = []
for relation in scene_graph["relations"]:
    subject = find_object(relation["subject_id"])
    predicate = relation["predicate"]
    object = find_object(relation["object_id"])
    description.append(f"{subject['label']} {predicate} {object['label']}")
# Output: ["person left_of car", "tree between house"]
```

### 3. Spatial Reasoning
Find all objects to the left of a specific object:
```python
target_id = 1  # car
left_objects = []
for relation in scene_graph["relations"]:
    if relation["predicate"] == "left_of" and relation["object_id"] == target_id:
        left_objects.append(find_object(relation["subject_id"]))
```

## Extensions

The format can be extended with additional fields:

### Object Attributes
```json
{
  "id": 0,
  "label": "car",
  "attributes": {
    "color": "red",
    "moving": true,
    "brand": "toyota"
  },
  ...
}
```

### Temporal Information
```json
{
  "meta": {
    "frame_id": 42,
    "video_timestamp": "00:01:23.456"
  },
  ...
}
```

### Confidence Breakdown
```json
{
  "score": 0.92,
  "score_breakdown": {
    "detection": 0.95,
    "classification": 0.97
  },
  ...
}
```

## See Also

- [Visual Genome](https://visualgenome.org/) - Scene graph dataset
- [Scene Graph Benchmark](https://github.com/KaihuaTang/Scene-Graph-Benchmark.pytorch)
- [COCO Dataset](https://cocodataset.org/)
