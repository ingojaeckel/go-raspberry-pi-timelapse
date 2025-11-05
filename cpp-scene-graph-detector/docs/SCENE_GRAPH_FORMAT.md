# Scene Graph JSON Format Specification

This document specifies the JSON schema for scene graphs produced by `cpp-scene-graph-detector`.

## Schema Overview

A scene graph JSON file contains three main sections:
1. **meta** - Metadata about the scene graph
2. **objects** - List of detected objects (nodes)
3. **relations** - List of spatial/semantic relations (edges)

## JSON Schema

### Top-Level Structure

```json
{
  "meta": { ... },
  "objects": [ ... ],
  "relations": [ ... ]
}
```

### Meta Section

Contains summary statistics about the scene graph.

**Fields:**
- `num_objects` (integer): Total number of objects detected
- `num_relations` (integer): Total number of relations found

**Example:**
```json
"meta": {
  "num_objects": 4,
  "num_relations": 6
}
```

### Objects Section

Array of detected objects, each representing a node in the scene graph.

**Object Schema:**
```json
{
  "id": integer,
  "label": string,
  "class_id": integer,
  "score": float,
  "bbox": {
    "x": float,
    "y": float,
    "width": float,
    "height": float
  }
}
```

**Fields:**
- `id` (integer): Unique identifier for this object within the scene graph
- `label` (string): Human-readable class name (e.g., "person", "car", "dog")
- `class_id` (integer): Numeric class ID from the model
- `score` (float): Detection confidence score [0.0, 1.0]
- `bbox` (object): Bounding box in center format
  - `x` (float): Center X coordinate
  - `y` (float): Center Y coordinate
  - `width` (float): Box width
  - `height` (float): Box height

**Example:**
```json
{
  "id": 0,
  "label": "person",
  "class_id": 0,
  "score": 0.87,
  "bbox": {
    "x": 320.5,
    "y": 240.8,
    "width": 150.2,
    "height": 280.5
  }
}
```

### Relations Section

Array of spatial or semantic relations between objects, each representing an edge in the scene graph.

**Relation Schema:**
```json
{
  "subject_id": integer,
  "predicate": string,
  "object_id": integer,
  "score": float
}
```

**Fields:**
- `subject_id` (integer): ID of the subject object (references `objects[].id`)
- `predicate` (string): Relation type (e.g., "left_of", "contains", "overlaps")
- `object_id` (integer): ID of the object (references `objects[].id`)
- `score` (float): Confidence score for this relation [0.0, 1.0]

**Example:**
```json
{
  "subject_id": 0,
  "predicate": "left_of",
  "object_id": 1,
  "score": 1.0
}
```

## Supported Predicates

### Spatial Predicates

| Predicate | Description | Inference Method |
|-----------|-------------|------------------|
| `left_of` | Subject is to the left of object | Geometric |
| `right_of` | Subject is to the right of object | Geometric |
| `in_front_of` | Subject is in front of object (closer) | Model-based |
| `behind` | Subject is behind object (farther) | Model-based |
| `overlaps` | Bounding boxes overlap | Geometric (IoU > 0) |
| `contains` | Subject's bbox contains object's bbox | Geometric |
| `intersects` | Bounding boxes intersect | Geometric |
| `between` | Subject is between two objects | Geometric/Model |
| `on` | Subject is on top of object | Model-based |
| `under` | Subject is under object | Model-based |
| `next_to` | Subject is adjacent to object | Geometric |

**Note:** Geometric predicates are always available. Model-based predicates require a relation prediction model.

## Complete Example

### Simple Scene: Two Houses and a Tree

```json
{
  "meta": {
    "num_objects": 3,
    "num_relations": 2
  },
  "objects": [
    {
      "id": 0,
      "label": "house",
      "class_id": 75,
      "score": 0.92,
      "bbox": {
        "x": 200.0,
        "y": 300.0,
        "width": 180.0,
        "height": 200.0
      }
    },
    {
      "id": 1,
      "label": "house",
      "class_id": 75,
      "score": 0.88,
      "bbox": {
        "x": 600.0,
        "y": 300.0,
        "width": 180.0,
        "height": 200.0
      }
    },
    {
      "id": 2,
      "label": "tree",
      "class_id": 68,
      "score": 0.95,
      "bbox": {
        "x": 400.0,
        "y": 250.0,
        "width": 100.0,
        "height": 300.0
      }
    }
  ],
  "relations": [
    {
      "subject_id": 0,
      "predicate": "left_of",
      "object_id": 1,
      "score": 1.0
    },
    {
      "subject_id": 2,
      "predicate": "between",
      "object_id": 0,
      "score": 0.95
    }
  ]
}
```

### Complex Scene: Outdoor Setting with Multiple Objects

```json
{
  "meta": {
    "num_objects": 5,
    "num_relations": 8
  },
  "objects": [
    {
      "id": 0,
      "label": "person",
      "class_id": 0,
      "score": 0.87,
      "bbox": {
        "x": 320.0,
        "y": 400.0,
        "width": 120.0,
        "height": 240.0
      }
    },
    {
      "id": 1,
      "label": "car",
      "class_id": 2,
      "score": 0.93,
      "bbox": {
        "x": 150.0,
        "y": 450.0,
        "width": 200.0,
        "height": 150.0
      }
    },
    {
      "id": 2,
      "label": "dog",
      "class_id": 16,
      "score": 0.78,
      "bbox": {
        "x": 500.0,
        "y": 480.0,
        "width": 80.0,
        "height": 100.0
      }
    },
    {
      "id": 3,
      "label": "tree",
      "class_id": 68,
      "score": 0.91,
      "bbox": {
        "x": 700.0,
        "y": 300.0,
        "width": 120.0,
        "height": 400.0
      }
    },
    {
      "id": 4,
      "label": "bench",
      "class_id": 13,
      "score": 0.82,
      "bbox": {
        "x": 450.0,
        "y": 500.0,
        "width": 180.0,
        "height": 80.0
      }
    }
  ],
  "relations": [
    {
      "subject_id": 1,
      "predicate": "left_of",
      "object_id": 0,
      "score": 1.0
    },
    {
      "subject_id": 0,
      "predicate": "left_of",
      "object_id": 2,
      "score": 1.0
    },
    {
      "subject_id": 2,
      "predicate": "left_of",
      "object_id": 3,
      "score": 1.0
    },
    {
      "subject_id": 0,
      "predicate": "next_to",
      "object_id": 4,
      "score": 0.85
    },
    {
      "subject_id": 2,
      "predicate": "next_to",
      "object_id": 4,
      "score": 0.78
    },
    {
      "subject_id": 1,
      "predicate": "in_front_of",
      "object_id": 3,
      "score": 0.72
    },
    {
      "subject_id": 0,
      "predicate": "between",
      "object_id": 1,
      "score": 0.88
    },
    {
      "subject_id": 4,
      "predicate": "on",
      "object_id": -1,
      "score": 0.90
    }
  ]
}
```

## Validation

### Required Fields

All objects must have:
- `id` (unique within scene graph)
- `label` (non-empty string)
- `class_id` (non-negative integer)
- `score` (0.0 to 1.0)
- `bbox` with all four coordinates

All relations must have:
- `subject_id` (valid object ID)
- `predicate` (valid predicate string)
- `object_id` (valid object ID)
- `score` (0.0 to 1.0)

### Consistency Rules

1. All `subject_id` and `object_id` must reference existing objects
2. Object IDs should be unique within the scene graph
3. Bounding box coordinates should be non-negative
4. Confidence scores should be in range [0.0, 1.0]
5. At least one object should be present for a valid scene graph

## Usage in Code

### Reading JSON (Python example)

```python
import json

with open('scene_graph.json', 'r') as f:
    graph = json.load(f)

print(f"Objects: {graph['meta']['num_objects']}")
print(f"Relations: {graph['meta']['num_relations']}")

for obj in graph['objects']:
    print(f"  {obj['label']} at ({obj['bbox']['x']}, {obj['bbox']['y']})")

for rel in graph['relations']:
    subj_label = graph['objects'][rel['subject_id']]['label']
    obj_label = graph['objects'][rel['object_id']]['label']
    print(f"  {subj_label} {rel['predicate']} {obj_label}")
```

### Validation Schema (JSON Schema format)

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["meta", "objects", "relations"],
  "properties": {
    "meta": {
      "type": "object",
      "required": ["num_objects", "num_relations"],
      "properties": {
        "num_objects": { "type": "integer", "minimum": 0 },
        "num_relations": { "type": "integer", "minimum": 0 }
      }
    },
    "objects": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["id", "label", "class_id", "score", "bbox"],
        "properties": {
          "id": { "type": "integer" },
          "label": { "type": "string" },
          "class_id": { "type": "integer" },
          "score": { "type": "number", "minimum": 0, "maximum": 1 },
          "bbox": {
            "type": "object",
            "required": ["x", "y", "width", "height"],
            "properties": {
              "x": { "type": "number" },
              "y": { "type": "number" },
              "width": { "type": "number" },
              "height": { "type": "number" }
            }
          }
        }
      }
    },
    "relations": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["subject_id", "predicate", "object_id", "score"],
        "properties": {
          "subject_id": { "type": "integer" },
          "predicate": { "type": "string" },
          "object_id": { "type": "integer" },
          "score": { "type": "number", "minimum": 0, "maximum": 1 }
        }
      }
    }
  }
}
```

## Extensions

### Future Additions

The schema may be extended in future versions to support:
- Temporal information (frame numbers, timestamps)
- Object attributes (color, texture, size categories)
- Confidence matrices for relation alternatives
- Hierarchical scene structure
- 3D bounding boxes and depth information
