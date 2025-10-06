# Viewfinder Debug Information - Visual Mockup

## Example 1: Debug Info Enabled (Default)

```
┌──────────────────────────────────────────────────────────────────────┐
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░                                          │
│░ FPS: 15              ░░░░░                                          │
│░ Avg proc: 45 ms      ░░░░░                                          │
│░ Objects: 127         ░░░░░                                          │
│░ Images: 12           ░░░░░                                          │
│░ Uptime: 00:15:32     ░░░░░                                          │
│░ Camera 0: 1280x720   ░░░░░                                          │
│░ Detection: 640x360   ░░░░░                                          │
│░ GPU: ON              ░░░░░                                          │
│░ Burst: OFF           ░░░░░                                          │
│░ Disk: 45.2%          ░░░░░                                          │
│░ CPU: 58.3°C          ░░░░░                                          │
│░ --- Top Objects ---  ░░░░░                                          │
│░ person: 85           ░░░░░                                          │
│░ cat: 24              ░░░░░                                          │
│░ dog: 12              ░░░░░                                          │
│░ car: 4               ░░░░░                                          │
│░ bicycle: 2           ░░░░░                                          │
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░                                          │
│                                                                      │
│                                                                      │
│                    ┌────────────────────┐                            │
│                    │  person 92%        │                            │
│                    │                    │                            │
│                    │                    │                            │
│                    │                    │                            │
│                    │                    │                            │
│                    └────────────────────┘                            │
│                                                                      │
│         ┌────────────┐                                               │
│         │ cat 87%    │                                               │
│         │            │                                               │
│         │            │                                               │
│         └────────────┘                                               │
│                                                                      │
│                                                                      │
│                                                                      │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

**Notes:**
- ░ = Semi-transparent black background (60% opacity)
- Bounding boxes shown in color (person=green, cat=red, dog=blue, etc.)
- Debug overlay uses small font (0.4 scale) to minimize coverage
- Stats overlay ~200x330 pixels in top-left corner
- **NEW:** GPU, Burst mode, Disk usage, and CPU temperature metrics now displayed

## Example 2: Debug Info Disabled (After Pressing SPACE)

```
┌──────────────────────────────────────────────────────────────────────┐
│                                                                      │
│                                                                      │
│                                                                      │
│                                                                      │
│                                                                      │
│                                                                      │
│                                                                      │
│                                                                      │
│                                                                      │
│                                                                      │
│                    ┌────────────────────┐                            │
│                    │  person 92%        │                            │
│                    │                    │                            │
│                    │                    │                            │
│                    │                    │                            │
│                    │                    │                            │
│                    └────────────────────┘                            │
│                                                                      │
│         ┌────────────┐                                               │
│         │ cat 87%    │                                               │
│         │            │                                               │
│         │            │                                               │
│         └────────────┘                                               │
│                                                                      │
│                                                                      │
│                                                                      │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

**Notes:**
- Debug overlay completely hidden
- Full unobstructed view of camera feed
- Bounding boxes and labels still shown
- Press SPACE again to show debug overlay

## Example 3: High Activity Scenario

```
┌──────────────────────────────────────────────────────────────────────┐
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░                                          │
│░ FPS: 8               ░░░░░                                          │
│░ Avg proc: 95 ms      ░░░░░                                          │
│░ Objects: 1547        ░░░░░                                          │
│░ Images: 154          ░░░░░                                          │
│░ Uptime: 02:34:18     ░░░░░                                          │
│░ Camera 0: 1280x720   ░░░░░                                          │
│░ Detection: 640x360   ░░░░░                                          │
│░ GPU: ON              ░░░░░                                          │
│░ Burst: ON            ░░░░░                                          │
│░ Disk: 87.5%          ░░░░░                                          │
│░ CPU: 72.8°C          ░░░░░                                          │
│░ --- Top Objects ---  ░░░░░                                          │
│░ person: 842          ░░░░░                                          │
│░ car: 357             ░░░░░                                          │
│░ bicycle: 124         ░░░░░                                          │
│░ dog: 98              ░░░░░                                          │
│░ cat: 67              ░░░░░                                          │
│░ truck: 34            ░░░░░                                          │
│░ bus: 18              ░░░░░                                          │
│░ motorcycle: 7        ░░░░░                                          │
│░░░░░░░░░░░░░░░░░░░░░░░░░░░░                                          │
│                                                                      │
│  ┌──────┐   ┌──────┐                        ┌────────┐              │
│  │person│   │person│                        │ car 78%│              │
│  │ 95%  │   │ 89%  │                        │        │              │
│  └──────┘   └──────┘                        │        │              │
│                                             └────────┘              │
│                                                                      │
│               ┌──────────┐                                           │
│               │ bicycle  │                                           │
│               │   82%    │                                           │
│               └──────────┘                                           │
│                                                                      │
│                                        ┌─────┐                       │
│                                        │dog  │                       │
│                                        │92%  │                       │
│                                        └─────┘                       │
└──────────────────────────────────────────────────────────────────────┘
```

**Notes:**
- Higher object counts showing busy scene
- Lower FPS (8) due to processing load
- Many different object types detected
- Extended uptime (2+ hours)
- Multiple detections in single frame
- **NEW:** High disk usage (87.5%) and elevated CPU temp (72.8°C) showing system stress

## Color Coding

The bounding boxes use the following colors:

- **Green** (0, 255, 0): Person
- **Red** (0, 0, 255): Cat
- **Blue** (255, 0, 0): Dog
- **Yellow** (0, 255, 255): Vehicles (car, truck, bus)
- **Magenta** (255, 0, 255): Motorcycles and bicycles
- **White** (255, 255, 255): Other detected objects

Note: Colors shown in OpenCV BGR format

## Keyboard Controls Visual Guide

```
┌────────────────────────────────────────────┐
│  Keyboard Controls:                        │
│                                            │
│  ┌───────┐                                 │
│  │ SPACE │  Toggle debug info on/off       │
│  └───────┘                                 │
│                                            │
│  ┌───┐                                     │
│  │ q │  Close viewfinder and stop app      │
│  └───┘                                     │
│                                            │
│  ┌─────┐                                   │
│  │ ESC │  Close viewfinder and stop app    │
│  └─────┘                                   │
│                                            │
└────────────────────────────────────────────┘
```

## Performance Indicators

The debug overlay helps identify performance issues:

### Good Performance
```
FPS: 25               ← High FPS (target: 20-30)
Avg proc: 35 ms       ← Fast processing
```

### Marginal Performance
```
FPS: 12               ← Moderate FPS (acceptable: 10-20)
Avg proc: 75 ms       ← Slower processing
```

### Poor Performance (Warning)
```
FPS: 3                ← Low FPS (< 5 indicates problems)
Avg proc: 250 ms      ← Very slow processing
```

## System Health Indicators

The debug overlay includes system monitoring metrics to track resource usage:

### Normal System Health
```
Disk: 45.2%           ← Moderate disk usage
CPU: 58.3°C           ← Normal operating temperature
```

### Elevated Usage (Monitor)
```
Disk: 87.5%           ← High disk usage - consider cleanup
CPU: 72.8°C           ← Elevated temperature - check cooling
```

### Critical (Action Required)
```
Disk: 95.0%           ← Critical! Free up space immediately
CPU: 85.0°C           ← Warning! Thermal throttling likely
```

**Note:** System monitor metrics only appear when available. On systems without thermal sensors, CPU temperature won't be displayed.

## Usage Scenarios

### Development & Debugging
- Monitor FPS to optimize performance
- Track processing time to identify bottlenecks
- Verify objects are being detected correctly
- Check detection resolution matches expectations
- Monitor system resources (disk space, CPU temperature)

### Field Deployment
- Verify camera is working (uptime increases)
- Monitor detection accuracy (which objects are seen)
- Validate image saving (images counter increases)
- Quick performance check without log files
- Track disk usage to prevent storage issues
- Monitor CPU temperature for thermal problems

### Demonstrations
- Show real-time statistics to stakeholders
- Prove detection is working
- Display system capabilities
- Toggle off for clean presentation view
