# Configuration System

This document describes the configuration system for the Go Raspberry Pi Timelapse application.

## Configuration Sources

The application uses three sources for configuration:

1. **Default values** - Hard-coded defaults in the application
2. **Persisted settings** - Settings saved to disk (in `timelapse-settings.db`)
3. **CLI flags** - Command-line arguments provided when starting the application

## Priority Order

Configuration values are determined using the following priority order (highest to lowest):

1. **CLI flags** - Always take precedence when provided
2. **Persisted settings** - Used if no CLI flag is provided
3. **Default values** - Used as fallback if neither CLI flag nor persisted setting exists

This means:
- CLI flags override both persisted settings and defaults
- Persisted settings can be changed via the REST API during runtime
- If you restart the application with the same CLI flags, they will continue to override persisted settings

## Configuration Parameters

### CLI Flags

The following flags can be provided when starting the application:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-port` | string | `:8080` | HTTP port to listen on |
| `-logToFile` | bool | `false` | Enable logging to file instead of stdout |
| `-storageFolder` | string | `timelapse-pictures` | Folder for storing timelapse pictures |
| `-secondsBetweenCaptures` | int | `1800` (30 min) | Number of seconds between captures |

### Persisted Settings

The following settings can be changed via the REST API and are persisted to disk:

| Setting | Type | Validation | Description |
|---------|------|------------|-------------|
| `Quality` | int | 1-100 | Image quality (1=lowest, 100=highest) |
| `SecondsBetweenCaptures` | int | ≥ 10 | Seconds between captures (minimum 10 for sufficient exposure) |
| `OffsetWithinHour` | int | -1 or 0-3599 | Offset in seconds for scheduled captures (-1 to disable) |
| `PhotoResolutionWidth` | int | - | Width of captured photos in pixels |
| `PhotoResolutionHeight` | int | - | Height of captured photos in pixels |
| `PreviewResolutionWidth` | int | - | Width of preview images in pixels |
| `PreviewResolutionHeight` | int | - | Height of preview images in pixels |
| `RotateBy` | int | - | Rotation angle (0 or 180) |
| `ResolutionSetting` | int | 0-2 | Resolution preset (0=full, 1=66%, 2=50%) |
| `DebugEnabled` | bool | - | Enable debug logging |

## Validation and Sanitization

The application enforces validation rules to prevent broken configurations:

### Quality
- **Minimum**: 1 (quality=0 is not supported by the camera)
- **Maximum**: 100
- Invalid values are automatically sanitized to the nearest valid value

### Seconds Between Captures
- **Minimum**: 10 seconds (to allow sufficient exposure time)
- Values below the minimum are automatically raised to the minimum

### Offset Within Hour
- **Valid range**: -1 (disabled) or 0-3599 seconds
- -1 disables the offset feature
- Values outside the range are sanitized to the nearest valid value

### When Validation Occurs

1. **CLI flags**: Validated when the application starts
   - Invalid CLI flags cause the application to exit with an error message
   
2. **Persisted settings on write**: Sanitized before saving to disk
   - Ensures only valid configurations are persisted
   
3. **Persisted settings on load**: Sanitized after loading from disk
   - Provides backward compatibility if validation rules change

4. **REST API updates**: Validated before accepting the update
   - Invalid updates are rejected with an HTTP 400 error

## Examples

### Starting with CLI flag override

```bash
# Start with 60 second capture interval (overrides persisted setting)
./timelapsepi -secondsBetweenCaptures=60

# Even if you previously saved a different interval via the REST API,
# this CLI flag will take precedence
```

### Invalid CLI flag

```bash
# This will fail with a clear error message
./timelapsepi -secondsBetweenCaptures=5
# Error: secondsBetweenCaptures must be at least 10 seconds to allow sufficient exposure time (got 5)
```

### Updating configuration via REST API

```bash
# Update configuration via REST API
curl -X POST http://localhost:8080/configuration \
  -H "Content-Type: application/json" \
  -d '{
    "Quality": 85,
    "SecondsBetweenCaptures": 300,
    "OffsetWithinHour": 900
  }'

# Invalid configuration will be rejected
curl -X POST http://localhost:8080/configuration \
  -H "Content-Type: application/json" \
  -d '{
    "Quality": 0,
    "SecondsBetweenCaptures": 5
  }'
# Returns HTTP 400 with error message
```

### Automatic Sanitization

When loading or saving configuration, invalid values are automatically corrected:

```go
// If persisted settings contain Quality: 0, it will be sanitized to 1
// If persisted settings contain SecondsBetweenCaptures: 5, it will be sanitized to 10
// If persisted settings contain OffsetWithinHour: 5000, it will be sanitized to 3599
```

## Implementation Notes

The configuration system uses several components:

- **`conf/conf.go`**: Defines constants and global configuration variables
- **`conf/model.go`**: Defines the `Settings` struct with validation and sanitization methods
- **`conf/settings.go`**: Handles loading and saving configuration to/from disk
- **`conf/valid/`**: Contains validation logic used by the REST API
- **`main.go`**: Applies configuration priority (CLI → persisted → defaults)

The system ensures that:
1. Invalid configurations cannot be persisted
2. CLI flags always take precedence over persisted settings
3. Configuration is validated at multiple points to catch errors early
4. Backward compatibility is maintained through automatic sanitization
