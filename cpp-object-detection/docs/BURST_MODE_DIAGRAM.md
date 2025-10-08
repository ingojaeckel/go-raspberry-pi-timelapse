# Burst Mode - Visual Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         BURST MODE STATE MACHINE                        │
└─────────────────────────────────────────────────────────────────────────┘

                              ┌──────────────┐
                              │   STARTUP    │
                              │ (Burst OFF)  │
                              └──────┬───────┘
                                     │
                                     ▼
                    ┌────────────────────────────────┐
                    │      INACTIVE STATE            │
                    │                                │
                    │  • Normal rate limiting        │
                    │  • Sleep between frames        │
                    │  • Energy efficient (~1 FPS)   │
                    └────────┬──────────────▲────────┘
                             │              │
              New object     │              │    All objects stationary
              type enters    │              │    OR no objects
                             │              │    OR only known objects
                             ▼              │
                    ┌────────────────────────────────┐
                    │       ACTIVE STATE             │
                    │                                │
                    │  • Skip sleep intervals        │
                    │  • Max out FPS                 │
                    │  • Capture rapid motion        │
                    │  • High CPU usage (~5-30 FPS)  │
                    └────────────────────────────────┘


┌─────────────────────────────────────────────────────────────────────────┐
│                          EXAMPLE SCENARIO                               │
└─────────────────────────────────────────────────────────────────────────┘

Frame N:     ┌──────────────┐
             │ car          │    Burst: OFF    Sleep: 950ms (1 FPS baseline)
             │ (stationary) │
             └──────────────┘    All objects known and stationary
                    │
                    ▼
Frame N+1:   ┌──────────────┬──────────────┐
             │ car          │ person       │    Burst: ON → ACTIVATED
             │ (stationary) │ (NEW!)       │
             └──────────────┴──────────────┘    Log: "Burst mode ACTIVATED"
                    │                            Sleep: 1ms (max FPS)
                    │
                    ▼
Frame N+2:   ┌──────────────┬──────────────┐
             │ car          │ person       │    Burst: ON (active)
             │ (stationary) │ (moving)     │
             └──────────────┴──────────────┘    Log: "Burst mode active: skipping..."
                    │                            Sleep: 1ms (max FPS)
                    │
                    ▼
Frame N+3:   ┌──────────────┬──────────────┐
             │ car          │ person       │    Burst: ON → OFF (deactivating)
             │ (stationary) │ (stationary) │
             └──────────────┴──────────────┘    Log: "Burst mode DEACTIVATED"
                    │                            Sleep: 950ms (back to 1 FPS)
                    │
                    ▼
Frame N+4:   ┌──────────────┐
             │ car          │    Burst: OFF
             │ (stationary) │
             └──────────────┘    Sleep: 950ms (1 FPS baseline)
                                  Person left scene


┌─────────────────────────────────────────────────────────────────────────┐
│                      TRANSITION TRIGGERS                                │
└─────────────────────────────────────────────────────────────────────────┘

ACTIVATE (OFF → ON):
  ┌─────────────────────────────────────────────────────────────┐
  │ • New object TYPE appears (not in previous frame)           │
  │ • OR new object INSTANCE detected (is_new flag set)         │
  └─────────────────────────────────────────────────────────────┘

DEACTIVATE (ON → OFF):
  ┌─────────────────────────────────────────────────────────────┐
  │ • All tracked objects become stationary                     │
  │ • OR no objects present in scene                            │
  │ • OR only previously-known objects remain                   │
  └─────────────────────────────────────────────────────────────┘


┌─────────────────────────────────────────────────────────────────────────┐
│                        PERFORMANCE IMPACT                               │
└─────────────────────────────────────────────────────────────────────────┘

            │
  100% CPU  │          ┌─────┐         Burst Active
            │          │     │         (new object)
            │          │     │
   50% CPU  │          │     │
            │──────────┘     └──────── Burst Inactive
            │                          (baseline 1 FPS)
    0% CPU  │
            └────────────────────────────────────────────▶ Time
               Idle    Event    Idle

            Energy Saved: ~95% during idle periods
            Event Capture: Maximum temporal resolution


┌─────────────────────────────────────────────────────────────────────────┐
│                         INTEGRATION POINTS                              │
└─────────────────────────────────────────────────────────────────────────┘

  ┌──────────────────┐         ┌──────────────────┐
  │  Object Tracker  │────────▶│   Burst Mode     │
  │                  │         │   Logic          │
  │  • is_new flag   │         │                  │
  │  • is_stationary │         │  • State machine │
  │  • object_type   │         │  • Logging       │
  └──────────────────┘         └────────┬─────────┘
                                        │
                                        ▼
                              ┌──────────────────┐
                              │  Rate Limiter    │
                              │                  │
                              │  Skip sleep if   │
                              │  burst active    │
                              └──────────────────┘
```

## Key Benefits

1. **Energy Efficiency**: ~95% CPU idle time during quiet periods
2. **Event Capture**: Maximum FPS during important moments
3. **Automatic**: No manual intervention required
4. **Smart**: Adapts to scene activity in real-time
5. **Debuggable**: Comprehensive logging of state transitions

## Use Cases

- **Security Monitoring**: Low power idle, high capture when intruder appears
- **Wildlife Observation**: Efficient battery use, detailed tracking when animals move
- **Long-term Deployments**: Minimize power consumption while maintaining coverage
- **Embedded Systems**: Optimize limited resources (Raspberry Pi, etc.)
