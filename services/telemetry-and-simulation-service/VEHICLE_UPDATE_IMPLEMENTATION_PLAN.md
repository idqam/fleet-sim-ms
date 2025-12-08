# Vehicle Update and Simulation Loop Implementation Plan

## Overview

This document provides a high-level implementation plan for the vehicle update system and simulation loop. The goal is to create a working simulation where vehicles navigate along precomputed routes using waypoint-based navigation.

---

## Architecture Components

### 1. Vehicle Update System
- **File**: `vehicle-update.go`
- **Purpose**: Core logic for moving vehicles along their routes
- **Key Function**: `UpdateVehicle(vehicle, graph, deltaTime)`

### 2. Simulation Engine
- **File**: `simulation-engine.go`
- **Purpose**: Manages the simulation loop and vehicle lifecycle
- **Key Components**: Main update loop, vehicle registry, timing control

### 3. Telemetry System (Optional Phase 2)
- **File**: `telemetry.go`
- **Purpose**: Emit position updates and events
- **Integration**: Called from update loop

---

## Phase 1: Vehicle Update Function

### File Structure
```
vehicle-update.go
├── UpdateVehicle(vehicle, graph, deltaTime) error
├── interpolatePosition(from, to, progress) Vector2D
├── calculateVelocity(from, to, speed) Vector2D
└── Helper functions
```

### UpdateVehicle Algorithm

#### Input Validation
```
FUNCTION UpdateVehicle(vehicle, graph, deltaTime):
    IF vehicle.Route == nil:
        RETURN nil
    
    IF vehicle.Route.CompletedAt != nil:
        SET vehicle.State.Status = Arrived
        RETURN nil
```

#### Route Completion Check
```
    IF vehicle.Route.CurrentEdgeIndex >= LENGTH(vehicle.Route.Edges):
        now = CurrentTime()
        SET vehicle.Route.CompletedAt = &now
        SET vehicle.State.Status = Arrived
        SET vehicle.State.Velocity = ZeroVector
        SET vehicle.State.CurrentPosition = EndNodePosition
        RETURN nil
```

#### Get Current Edge
```
    currentEdgeID = vehicle.Route.Edges[CurrentEdgeIndex]
    edge = graph.Edges[currentEdgeID]
    
    IF edge == nil:
        RETURN Error("Edge not found")
```

#### Calculate Movement
```
    speed = edge.Conditions.EffectiveSpeedLimit
    IF speed <= 0:
        speed = edge.BaseSpeedLimit
    
    distanceTraveled = speed * deltaTime
    progressIncrement = distanceTraveled / edge.Length
    
    vehicle.State.ProgressOnEdge += progressIncrement
```

#### Waypoint Transition Logic
```
    IF vehicle.State.ProgressOnEdge >= 1.0:
        # Waypoint reached!
        vehicle.Route.CurrentNode = vehicle.Route.TargetNode
        vehicle.Route.CurrentEdgeIndex++
        
        IF CurrentEdgeIndex < LENGTH(Edges):
            # More edges to traverse
            vehicle.State.ProgressOnEdge = 0.0  # Simple version
            # OR: vehicle.State.ProgressOnEdge -= 1.0  # Carry excess
            
            vehicle.State.CurrentEdge = Edges[CurrentEdgeIndex]
            
            nextEdge = graph.Edges[CurrentEdge]
            IF nextEdge != nil:
                vehicle.Route.TargetNode = nextEdge.To
        ELSE:
            # No more edges
            vehicle.State.ProgressOnEdge = 1.0
            vehicle.State.CurrentEdge = ""
```

#### Position Interpolation
```
    fromNode = graph.Nodes[edge.From]
    toNode = graph.Nodes[edge.To]
    
    progress = CLAMP(vehicle.State.ProgressOnEdge, 0.0, 1.0)
    
    vehicle.State.CurrentPosition = LERP(
        fromNode.Position,
        toNode.Position,
        progress
    )
```

**LERP Formula:**
```
FUNCTION LERP(start, end, t):
    RETURN Vector2D{
        X: start.X + (end.X - start.X) * t,
        Y: start.Y + (end.Y - start.Y) * t
    }
```

#### Velocity Calculation
```
    direction = toNode.Position - fromNode.Position
    length = MAGNITUDE(direction)
    
    IF length > 0:
        normalized = direction / length
        vehicle.State.Velocity = normalized * speed
    ELSE:
        vehicle.State.Velocity = ZeroVector
```

#### Metadata Update
```
    vehicle.State.LastUpdateTime = CurrentTime()
    
    IF vehicle.State.Status != Arrived:
        vehicle.State.Status = Moving
    
    RETURN nil
```

---

## Phase 2: Simulation Engine

### File Structure
```
simulation-engine.go
├── SimulationEngine struct
├── NewSimulationEngine() *SimulationEngine
├── AddVehicle(vehicle)
├── RemoveVehicle(vehicleID)
├── Run() (single loop version)
├── RunVehicleGoroutine(vehicle) (per-vehicle version)
└── Stop()
```

### SimulationEngine Structure
```
STRUCT SimulationEngine:
    vehicles: MAP[string]*Vehicle
    graph: *MapGraph
    updateRate: Duration (e.g., 100ms)
    stopChan: Channel
    mutex: Mutex (for thread safety)
    isRunning: bool
```

### Initialization
```
FUNCTION NewSimulationEngine(graph, updateRate):
    RETURN SimulationEngine{
        vehicles: EmptyMap,
        graph: graph,
        updateRate: updateRate,
        stopChan: MakeChannel(),
        isRunning: false
    }
```

### Vehicle Management
```
FUNCTION AddVehicle(sim, vehicle):
    LOCK sim.mutex
    sim.vehicles[vehicle.ID] = vehicle
    UNLOCK sim.mutex
    
    IF sim.isRunning AND using goroutine-per-vehicle:
        sim.RunVehicleGoroutine(vehicle)

FUNCTION RemoveVehicle(sim, vehicleID):
    LOCK sim.mutex
    DELETE sim.vehicles[vehicleID]
    UNLOCK sim.mutex
```

---

## Phase 3: Update Loop Patterns

### Pattern A: Single Central Loop (Recommended First)

**Advantages:**
- Simpler synchronization
- Easier to debug
- Deterministic execution order
- Better for testing

**Implementation:**
```
FUNCTION Run(sim):
    sim.isRunning = true
    ticker = NewTicker(sim.updateRate)
    lastUpdate = CurrentTime()
    
    LOOP:
        SELECT:
            CASE <-ticker.C:
                now = CurrentTime()
                deltaTime = now - lastUpdate
                lastUpdate = now
                
                LOCK sim.mutex
                FOR EACH vehicle IN sim.vehicles:
                    error = UpdateVehicle(vehicle, sim.graph, deltaTime)
                    IF error != nil:
                        LOG error
                UNLOCK sim.mutex
                
            CASE <-sim.stopChan:
                ticker.Stop()
                RETURN
```

**Timing Considerations:**
- Use actual elapsed time for `deltaTime`
- Handle variable tick rates gracefully
- Consider max deltaTime cap (e.g., 0.5s) for lag spikes

---

### Pattern B: Goroutine Per Vehicle (Advanced)

**Advantages:**
- True parallelism
- Vehicles update independently
- Scales better with many vehicles

**Disadvantages:**
- More complex synchronization
- Harder to debug
- Non-deterministic ordering

**Implementation:**
```
FUNCTION RunVehicleGoroutine(sim, vehicle):
    GO ROUTINE:
        ticker = NewTicker(sim.updateRate)
        lastUpdate = CurrentTime()
        
        LOOP:
            SELECT:
                CASE <-ticker.C:
                    now = CurrentTime()
                    deltaTime = now - lastUpdate
                    lastUpdate = now
                    
                    LOCK vehicle.mutex  # Per-vehicle lock
                    error = UpdateVehicle(vehicle, sim.graph, deltaTime)
                    UNLOCK vehicle.mutex
                    
                    IF error != nil:
                        LOG error
                    
                    IF vehicle.Route.CompletedAt != nil:
                        # Vehicle finished, exit goroutine
                        ticker.Stop()
                        RETURN
                
                CASE <-vehicle.stopChan:
                    ticker.Stop()
                    RETURN
```

**Synchronization Strategy:**
- Add `mutex` field to Vehicle struct
- Lock only during update
- Graph is read-only (no lock needed)
- Consider using channels for telemetry instead of shared state

---

## Phase 4: Testing Strategy

### Test File Structure
```
vehicle-update_test.go
├── TestUpdateVehicle_SingleEdge
├── TestUpdateVehicle_MultiEdge
├── TestUpdateVehicle_RouteCompletion
├── TestUpdateVehicle_ZeroLength
└── TestUpdateVehicle_EdgeTransition
```

### Test Cases

#### Test 1: Single Edge Movement
```
SETUP:
    Create graph with 2 nodes (A, B) connected by 1 edge
    Edge length = 100 units
    Edge speed = 10 m/s
    
EXECUTE:
    Assign route A→B to vehicle
    Call UpdateVehicle with deltaTime = 1.0s
    
ASSERT:
    ProgressOnEdge = 0.1 (traveled 10 units of 100)
    Position interpolated 10% from A to B
    Status = Moving
    
EXECUTE:
    Call UpdateVehicle 9 more times (deltaTime = 1.0s each)
    
ASSERT:
    ProgressOnEdge >= 1.0
    CurrentNode = B
    Status = Arrived
    CompletedAt is set
```

#### Test 2: Multi-Edge Route
```
SETUP:
    Create graph A→B→C (2 edges)
    Each edge 100 units, speed 10 m/s
    
EXECUTE:
    Assign route A→C to vehicle
    Update until ProgressOnEdge >= 1.0 on first edge
    
ASSERT:
    CurrentNode = B
    CurrentEdgeIndex = 1
    TargetNode = C
    ProgressOnEdge reset to 0.0
    
EXECUTE:
    Continue updating until route complete
    
ASSERT:
    CurrentNode = C
    Status = Arrived
    CompletedAt is set
```

#### Test 3: Edge Transition Accuracy
```
SETUP:
    Graph A→B, edge length 100, speed 15 m/s
    
EXECUTE:
    UpdateVehicle with deltaTime = 7.0s
    # Should travel 105 units (overshoot by 5)
    
ASSERT (Simple version):
    ProgressOnEdge = 1.0 (clamped)
    Route complete
    
ASSERT (Carry-over version):
    CurrentEdgeIndex incremented
    ProgressOnEdge = 0.05 (carried over)
```

---

## Phase 5: Integration Testing

### Simulation Loop Test
```
SETUP:
    Create simulation engine with test graph
    Add 3 vehicles with different routes
    
EXECUTE:
    Run simulation for 10 seconds (simulated time)
    
ASSERT:
    All vehicles moved along their routes
    Positions updated correctly
    At least 1 vehicle completed route
    No errors logged
```

### Concurrent Update Test (Goroutine Pattern)
```
SETUP:
    Create 10 vehicles
    Start goroutine for each
    
EXECUTE:
    Let run for 5 seconds
    Stop all goroutines
    
ASSERT:
    All vehicles updated independently
    No race conditions (run with -race flag)
    All goroutines terminated cleanly
```

---

## Implementation Checklist

### Phase 1: Core Update Logic
- [ ] Create `vehicle-update.go`
- [ ] Implement `UpdateVehicle()` function
- [ ] Implement position interpolation helper
- [ ] Implement velocity calculation helper
- [ ] Handle edge cases (nil checks, zero length, etc.)
- [ ] Write unit tests for UpdateVehicle

### Phase 2: Simulation Engine
- [ ] Create `simulation-engine.go`
- [ ] Define SimulationEngine struct
- [ ] Implement vehicle add/remove
- [ ] Implement single-loop Run() method
- [ ] Add graceful shutdown
- [ ] Write integration tests

### Phase 3: Goroutine Pattern (Optional)
- [ ] Add mutex to Vehicle struct (if needed)
- [ ] Implement RunVehicleGoroutine()
- [ ] Add per-vehicle stop channel
- [ ] Test with race detector
- [ ] Benchmark performance vs single loop

### Phase 4: Telemetry (Optional)
- [ ] Define telemetry event types
- [ ] Emit position updates
- [ ] Emit route completion events
- [ ] Add telemetry channel/callback system

---

## Key Design Decisions

### 1. Progress Tracking
**Decision Point:** How to handle edge transitions?

**Option A - Simple (Recommended):**
- Reset `ProgressOnEdge` to 0.0 when transitioning
- Accept slight inaccuracy at transitions
- Simpler code, easier to debug

**Option B - Accurate:**
- Carry over excess progress: `ProgressOnEdge -= 1.0`
- More accurate but complex
- May need multiple transitions in one update

**Recommendation:** Start with Option A, optimize to B if needed

---

### 2. Update Loop Pattern
**Decision Point:** Single loop or goroutine-per-vehicle?

**Single Loop:**
- ✅ Simpler
- ✅ Deterministic
- ✅ Easier testing
- ❌ Sequential updates

**Goroutine-Per-Vehicle:**
- ✅ Parallel updates
- ✅ Scales better
- ❌ Complex synchronization
- ❌ Harder to debug

**Recommendation:** Implement single loop first, migrate to goroutines if performance requires

---

### 3. Time Management
**Decision Point:** Fixed vs variable deltaTime?

**Fixed Timestep:**
```
deltaTime = 0.1  # Always 100ms
```
- ✅ Deterministic
- ✅ Simpler
- ❌ Doesn't match real time

**Variable Timestep:**
```
deltaTime = actualElapsedTime
```
- ✅ Matches real time
- ✅ Handles lag gracefully
- ❌ Non-deterministic

**Recommendation:** Variable timestep with max cap (e.g., 0.5s)

---

### 4. Synchronization Strategy
**Decision Point:** How to protect shared state?

**Option A - Coarse Lock:**
```
Lock entire simulation during update
```
- ✅ Simple
- ❌ Blocks all access

**Option B - Fine-Grained Locks:**
```
Lock per vehicle during update
```
- ✅ Better concurrency
- ❌ More complex

**Option C - Lock-Free (Channels):**
```
Send updates via channels
```
- ✅ No locks
- ❌ Most complex

**Recommendation:** Start with coarse lock (single loop), move to fine-grained if using goroutines

---

## Performance Considerations

### Optimization Opportunities
1. **Pre-compute edge directions** - Cache normalized direction vectors
2. **Batch updates** - Update multiple vehicles before acquiring lock
3. **Spatial partitioning** - Only update vehicles in active regions
4. **Event-driven updates** - Only update moving vehicles
5. **Delta compression** - Only emit telemetry when position changes significantly

### Profiling Points
- Time spent in UpdateVehicle
- Lock contention (if using goroutines)
- Memory allocations per update
- Telemetry emission overhead

---

## Error Handling Strategy

### Recoverable Errors
- Edge not found → Stop vehicle, log error, continue simulation
- Invalid progress value → Clamp to valid range, log warning
- Missing node → Use last known position, log error

### Fatal Errors
- Nil graph → Panic or return error before starting
- Corrupted vehicle state → Log and remove vehicle
- Channel closed unexpectedly → Shutdown gracefully

### Logging Levels
- **DEBUG**: Every update (too verbose for production)
- **INFO**: Route started, route completed
- **WARN**: Edge not found, invalid state recovered
- **ERROR**: Fatal errors, vehicle removed

---

## Next Steps After Implementation

1. **Visualization** - Add simple console output or web UI to see vehicles moving
2. **Metrics** - Track vehicles in motion, completed routes, average speed
3. **Dynamic routing** - Allow route changes mid-journey
4. **Collision detection** - Prevent vehicles from occupying same position
5. **Traffic simulation** - Adjust speeds based on vehicle density
6. **Weather effects** - Apply weather multipliers to speeds
7. **Energy simulation** - Track battery/fuel consumption

---

## Success Criteria

### Minimum Viable Product
- ✅ Vehicles move along assigned routes
- ✅ Positions update smoothly
- ✅ Routes complete successfully
- ✅ No crashes or panics
- ✅ Tests pass

### Production Ready
- ✅ All of MVP
- ✅ Handles 100+ concurrent vehicles
- ✅ Graceful error handling
- ✅ Telemetry emission working
- ✅ Performance profiled and optimized
- ✅ Race detector clean (if using goroutines)

---

## Appendix: Pseudocode Reference

### Complete UpdateVehicle Flow
```
FUNCTION UpdateVehicle(vehicle, graph, deltaTime):
    # 1. Validation
    IF vehicle.Route == nil OR vehicle.Route.CompletedAt != nil:
        RETURN early
    
    # 2. Route completion check
    IF CurrentEdgeIndex >= EdgeCount:
        Mark route complete
        RETURN
    
    # 3. Get current edge
    edge = graph.Edges[CurrentEdge]
    IF edge == nil:
        RETURN error
    
    # 4. Calculate movement
    speed = edge.Conditions.EffectiveSpeedLimit
    distanceTraveled = speed * deltaTime
    progressIncrement = distanceTraveled / edge.Length
    ProgressOnEdge += progressIncrement
    
    # 5. Check waypoint reached
    IF ProgressOnEdge >= 1.0:
        CurrentNode = TargetNode
        CurrentEdgeIndex++
        
        IF more edges:
            ProgressOnEdge = 0.0
            Update CurrentEdge and TargetNode
        ELSE:
            ProgressOnEdge = 1.0
    
    # 6. Interpolate position
    fromNode = graph.Nodes[edge.From]
    toNode = graph.Nodes[edge.To]
    CurrentPosition = LERP(fromNode.Position, toNode.Position, ProgressOnEdge)
    
    # 7. Calculate velocity
    direction = NORMALIZE(toNode.Position - fromNode.Position)
    Velocity = direction * speed
    
    # 8. Update metadata
    LastUpdateTime = now
    Status = Moving
    
    RETURN nil
```

---

**End of Implementation Plan**

This plan provides the complete roadmap for implementing vehicle updates and simulation loops. Follow the phases sequentially, test thoroughly at each step, and iterate based on performance requirements.
