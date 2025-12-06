# MVP Event Schema - Phase 1

## Overview

This document defines the **minimal viable event schema** for the fleet simulation's first phase. The goal is to establish end-to-end data flow (Go → Message Queue → Java → React) with simple position events before adding complexity like weather and physics.

---

## BasicVehiclePosEvent

### Purpose
Minimal telemetry event to track vehicle position on the road network graph.

### Schema

```go
type BasicVehiclePosEvent struct {
    VehicleID        string    `json:"vehicle_id"`
    EdgeID           string    `json:"edge_id"`
    MostRecentNodeID string    `json:"recent_node_id"`
    Progress         float64   `json:"progress"`
    Timestamp        time.Time `json:"timestamp"`
}
```

### Field Descriptions

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `VehicleID` | `string` | Unique identifier for the vehicle | `"vehicle_123"` |
| `EdgeID` | `string` | Current road edge the vehicle is on | `"edge_A_to_B"` |
| `MostRecentNodeID` | `string` | Last node the vehicle passed (start of current edge) | `"node_A"` |
| `Progress` | `float64` | Distance along the edge (0.0 = start, 1.0 = end) | `0.75` (75% complete) |
| `Timestamp` | `time.Time` | When this position was recorded | `"2024-12-05T19:37:00Z"` |

---

## Design Review ✅

### What's Good

✅ **Minimal but Complete**: Contains exactly what's needed to visualize vehicle movement  
✅ **Graph-Based**: Uses EdgeID + Progress instead of raw coordinates (matches your architecture)  
✅ **MostRecentNodeID**: Smart addition! Helps with:
- Determining direction of travel
- Calculating next edge when Progress reaches 1.0
- Debugging route progression

✅ **JSON Tags**: Proper serialization for cross-service communication  
✅ **Timestamp**: Essential for time-series storage and event ordering

### Suggestions

#### 1. Consider Adding `FleetID` (Optional)
```go
type BasicVehiclePosEvent struct {
    VehicleID        string    `json:"vehicle_id"`
    FleetID          string    `json:"fleet_id"`  // NEW
    EdgeID           string    `json:"edge_id"`
    MostRecentNodeID string    `json:"recent_node_id"`
    Progress         float64   `json:"progress"`
    Timestamp        time.Time `json:"timestamp"`
}
```

**Why?**: Enables fleet-level filtering in React dashboard ("show only Fleet A vehicles")

#### 2. Consider Renaming `MostRecentNodeID` → `FromNodeID`
```go
FromNodeID string `json:"from_node_id"`  // Clearer intent
```

**Why?**: More explicit that this is the edge's start node

---

## Event Flow

```
┌──────────────────┐
│  VehicleAgent    │
│  (Go Goroutine)  │
└────────┬─────────┘
         │ Emits BasicVehiclePosEvent every tick
         ▼
┌──────────────────┐
│ TelemetryEmitter │
│   (Interface)    │
└────────┬─────────┘
         │ Serializes to JSON
         ▼
┌──────────────────┐
│  Message Queue   │
│ (Kafka/RabbitMQ) │
│  Topic: vehicle- │
│    positions     │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Java Spring     │
│  Boot Consumer   │
│  - Stores in DB  │
│  - Broadcasts    │
│    via WebSocket │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  React Frontend  │
│  - Receives via  │
│    WebSocket     │
│  - Updates map   │
│    visualization │
└──────────────────┘
```

---

## Example Event (JSON)

```json
{
  "vehicle_id": "vehicle_001",
  "edge_id": "edge_depot_to_intersection_1",
  "recent_node_id": "node_depot",
  "progress": 0.42,
  "timestamp": "2024-12-05T19:37:15.123Z"
}
```

**Interpretation**: Vehicle 001 is 42% of the way from the depot to intersection 1.

---

## Database Schema (Java/PostgreSQL)

### Table: `vehicle_positions`

```sql
CREATE TABLE vehicle_positions (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id VARCHAR(50) NOT NULL,
    edge_id VARCHAR(100) NOT NULL,
    recent_node_id VARCHAR(100) NOT NULL,
    progress DOUBLE PRECISION NOT NULL CHECK (progress >= 0 AND progress <= 1),
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for time-series queries
CREATE INDEX idx_vehicle_positions_timestamp ON vehicle_positions(timestamp DESC);

-- Index for vehicle-specific queries
CREATE INDEX idx_vehicle_positions_vehicle_id ON vehicle_positions(vehicle_id, timestamp DESC);
```

**Alternative**: Use TimescaleDB for better time-series performance:
```sql
SELECT create_hypertable('vehicle_positions', 'timestamp');
```

---

## React Visualization

### Converting to Map Coordinates

```typescript
interface VehiclePosition {
  vehicleId: string;
  edgeId: string;
  recentNodeId: string;
  progress: number;
  timestamp: string;
}

function calculateMapPosition(
  event: VehiclePosition,
  mapGraph: MapGraph
): [number, number] {
  const edge = mapGraph.edges[event.edgeId];
  const fromNode = mapGraph.nodes[edge.from];
  const toNode = mapGraph.nodes[edge.to];
  
  // Linear interpolation
  const x = fromNode.x + (toNode.x - fromNode.x) * event.progress;
  const y = fromNode.y + (toNode.y - fromNode.y) * event.progress;
  
  return [x, y];
}
```

---

## Emission Frequency

### Recommended Settings

```go
type AgentConfig struct {
    PhysicsTickRate   time.Duration  // 60 FPS = 16ms
    TelemetryInterval time.Duration  // 1 second (emit every 60 ticks)
}
```

**Why 1 second?**
- ✅ Smooth visualization (1 update/sec is plenty for human perception)
- ✅ Reduces message queue load (60x fewer events than per-tick)
- ✅ Reasonable database write rate

**For 1000 vehicles**: 1000 events/second = very manageable

---

## Future Enhancements (Phase 2+)

Once the MVP works, enrich the event:

```go
type EnrichedVehiclePosEvent struct {
    // Phase 1 fields
    VehicleID        string    `json:"vehicle_id"`
    EdgeID           string    `json:"edge_id"`
    FromNodeID       string    `json:"from_node_id"`
    Progress         float64   `json:"progress"`
    Timestamp        time.Time `json:"timestamp"`
    
    // Phase 2: Add velocity
    Speed            float64   `json:"speed"`           // m/s
    Heading          float64   `json:"heading"`         // degrees
    
    // Phase 3: Add energy
    EnergyLevel      float64   `json:"energy_level"`    // 0.0 to 1.0
    
    // Phase 4: Add weather effects
    WeatherCondition string    `json:"weather_condition"` // "clear", "rain", etc.
    
    // Phase 5: Add status
    Status           string    `json:"status"`          // "moving", "idle", etc.
}
```

---

## Testing Strategy

### Unit Test (Go)
```go
func TestBasicVehiclePosEvent_Serialization(t *testing.T) {
    event := BasicVehiclePosEvent{
        VehicleID:        "v1",
        EdgeID:           "e1",
        MostRecentNodeID: "n1",
        Progress:         0.5,
        Timestamp:        time.Now(),
    }
    
    data, err := json.Marshal(event)
    assert.NoError(t, err)
    
    var decoded BasicVehiclePosEvent
    err = json.Unmarshal(data, &decoded)
    assert.NoError(t, err)
    assert.Equal(t, event.VehicleID, decoded.VehicleID)
}
```

### Integration Test
1. Start Go simulation with 1 vehicle
2. Verify events appear in message queue
3. Verify Java consumer receives events
4. Verify events stored in database
5. Verify React receives WebSocket updates

---

## Summary

Your `BasicVehiclePosEvent` is **excellent for MVP**! It's:
- ✅ Simple enough to implement quickly
- ✅ Complete enough to visualize vehicle movement
- ✅ Extensible for future phases
- ✅ Graph-based (matches your architecture)

**Recommendation**: Ship this as-is for Phase 1. The only optional addition is `FleetID` if you want fleet-level filtering in the UI.

🚀 **Ready to implement!**
