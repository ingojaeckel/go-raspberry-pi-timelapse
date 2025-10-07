# Configuration System Improvement - Implementation Summary

## Overview

This implementation addresses the issue of improving the configuration system by:
1. Better aligning various config sources (CLI flags, configuration package, settings persisted to disk)
2. Sanitizing config to prevent broken configurations from being persisted
3. Implementing proper priority and consistent setting of config values

## Problem Statement

The original issue identified several problems:
- Configuration sources (CLI flags, config package, persisted settings) were not properly aligned
- No sanitization prevented broken config from being persisted (e.g., quality=0, invalid secondsBetweenCaptures)
- No clear priority system for configuration values

## Solution Implemented

### 1. Validation Constants

Added centralized validation constants in `conf/conf.go`:

```go
const (
    MinSecondsBetweenCaptures = 10   // Allow sufficient exposure time (reduced from 60)
    MinQuality                = 1    // Quality must be at least 1 (prevents quality=0)
    MaxQuality                = 100  // Quality cannot exceed 100
    MinOffsetWithinHour       = 0    // Offset within hour minimum
    MaxOffsetWithinHour       = 3599 // Offset within hour maximum
)
```

### 2. Sanitization

Implemented `Settings.Sanitize()` method that enforces bounds:
- Quality: 1-100 (prevents quality=0 which is not supported)
- SecondsBetweenCaptures: ≥10 seconds (allows sufficient exposure time)
- OffsetWithinHour: -1 (disabled) or 0-3599 seconds

Applied sanitization at multiple points:
- **On write**: `WriteConfiguration()` sanitizes before persisting to disk
- **On load**: `LoadConfiguration()` sanitizes after loading from disk (backward compatibility)
- **On update**: REST API validates before accepting updates

### 3. Configuration Priority System

Implemented proper priority order: **CLI flags → Persisted settings → Defaults**

```go
// In main.go:
// 1. Load persisted settings (or defaults if none exist)
initialSettings, err := conf.LoadConfiguration()

// 2. Apply CLI overrides with proper priority
*initialSettings = initialSettings.ApplyCLIOverrides(secondsBetweenCaptures)
```

The `ApplyCLIOverrides()` method ensures CLI flags always take precedence over persisted settings.

### 4. CLI Flag Validation

Added validation for CLI flags before they are applied:

```go
func validateCLIFlags(secondsBetweenCaptures *int) error {
    if secondsBetweenCaptures != nil && *secondsBetweenCaptures < conf.MinSecondsBetweenCaptures {
        return fmt.Errorf("secondsBetweenCaptures must be at least %d seconds...", conf.MinSecondsBetweenCaptures)
    }
    return nil
}
```

Invalid CLI flags now cause the application to exit with a clear error message.

### 5. REST API Validation

The REST API already had validation via the `valid.Validator` interface. Updated it to use centralized constants:

```go
func (s strictValidator) Validate(settings conf.Settings) error {
    if settings.Quality < conf.MinQuality || settings.Quality > conf.MaxQuality {
        return errQualityOutOfBounds
    }
    if settings.SecondsBetweenCaptures < conf.MinSecondsBetweenCaptures {
        return errSecondsBetweenCapturesOutOfBounds
    }
    // ... more validation
}
```

## Files Modified

1. **conf/conf.go**
   - Added validation constants
   - Updated comments to document priority system

2. **conf/model.go**
   - Added `Sanitize()` method for automatic value correction
   - Added `ApplyCLIOverrides()` method for proper priority handling

3. **conf/settings.go**
   - Updated to use centralized constants
   - Applied sanitization in `LoadConfiguration()` and `WriteConfiguration()`

4. **conf/valid/strict_validator.go**
   - Updated to use centralized validation constants

5. **main.go**
   - Added CLI flag validation
   - Applied CLI overrides with proper priority

6. **rest/rest.go**
   - Removed TODO comment (validation is now complete)

## Files Created

1. **CONFIGURATION.md**
   - Comprehensive documentation of the configuration system
   - Lists all configuration sources and priority order
   - Documents validation rules and provides usage examples

## Tests Added

1. **conf/settings_test.go**
   - `TestSanitizeQuality` - Tests quality bounds enforcement
   - `TestSanitizeSecondsBetweenCaptures` - Tests minimum interval enforcement
   - `TestSanitizeOffsetWithinHour` - Tests offset validation
   - `TestApplyCLIOverrides` - Tests CLI priority system

2. **conf/valid/validator_test.go**
   - Updated to test new minimum values
   - Added tests for edge cases

## Verification

All tests pass successfully:
- Unit tests for sanitization
- Unit tests for validation
- Unit tests for CLI override priority
- End-to-end tests confirming CLI validation
- CodeQL security scan: 0 vulnerabilities

### Example Usage

**Valid CLI flag:**
```bash
./timelapsepi -secondsBetweenCaptures=30
# Application starts successfully with 30 second interval
```

**Invalid CLI flag:**
```bash
./timelapsepi -secondsBetweenCaptures=5
# Error: secondsBetweenCaptures must be at least 10 seconds to allow sufficient exposure time (got 5)
```

**REST API update:**
```bash
# Valid update
curl -X POST http://localhost:8080/configuration -d '{"Quality": 85, "SecondsBetweenCaptures": 300}'
# Returns: HTTP 200 with updated settings

# Invalid update
curl -X POST http://localhost:8080/configuration -d '{"Quality": 0, "SecondsBetweenCaptures": 5}'
# Returns: HTTP 400 with validation error
```

## Benefits

1. **Prevents broken configurations**: Sanitization ensures invalid values cannot be persisted
2. **Clear priority system**: CLI flags always override persisted settings, making behavior predictable
3. **Backward compatibility**: Automatic sanitization on load handles old configurations
4. **Better error messages**: Clear validation errors help users understand what went wrong
5. **Centralized validation**: Constants are defined once and used everywhere
6. **Well documented**: CONFIGURATION.md provides comprehensive documentation

## Conclusion

The configuration system has been significantly improved with:
- ✅ Proper alignment of config sources
- ✅ Sanitization preventing broken configs
- ✅ Clear priority system (CLI → persisted → defaults)
- ✅ Comprehensive validation at all entry points
- ✅ Thorough testing and documentation
- ✅ Zero security vulnerabilities
