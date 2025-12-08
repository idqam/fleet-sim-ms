# Simulation & Telemetry Implementation Plan

This document outlines the detailed architectural and implementation plan for adding telemetry emission, gRPC integration, real-time WebSocket updates, and management controls to the Simulation Engine.

## 1. Architecture Overview

The enhanced simulation engine will operate three concurrent time-domain loops per vehicle/system:
1.  **Physics Loop (Variable/High Freq):** Calculates movement, updates internal state (Already implemented).
2.  **Telemetry Loop (1Hz):** Snapshots state and queues data for persistence via gRPC.
3.  **Broadcast Loop (10Hz / 100ms):** Aggregates positions of all active vehicles for WebSocket streaming.

---

## 2. Component Implementation Details

### Phase 1: Telemetry Emission Logic

**Objective:** Vehicles must emit detailed telemetry every 1 second, independent of their movement interactions.

#### 1.1 Update `Vehicle` Struct
We need a buffered channel to hold telemetry snapshots to avoid blocking the physics loop.

**File:** `internal/simulation/entities/vehicle.go`
```go
type Vehicle struct {
    // ... existing fields
    TelemetryChan chan TelemetrySnapshot `json:"-"`
}

type TelemetrySnapshot struct {
    VehicleID string
    Timestamp time.Time
    Position  Vector2D
    Velocity  Vector2D
    // ... other sensor data (fuel, battery, etc.)
}
```

#### 1.2 Modify `RunVehicleGoroutine`
Introduce a second ticker specifically for telemetry.

**File:** `internal/simulation/simulation-engine/simulation-engine.go`
```go
// Inside the goroutine
physicsTicker := time.NewTicker(s.UpdateRate)
telemetryTicker := time.NewTicker(1 * time.Second)

for {
    select {
    case <-physicsTicker.C:
        // ... existing movement logic ...

    case <-telemetryTicker.C:
        // 1. Lock Mutex (Read Lock preferred)
        // 2. Create Snapshot
        // 3. Non-blocking send to TelemetryChan (drop if full)
    }
}
```

---

### Phase 2: Inter-Service Communication (gRPC)

**Objective:** The simulation service talks to "itself" (or the telemetry ingestion service) via gRPC to persist data.

#### 2.1 Internal gRPC Client
Create a wrapper around the generated gRPC client to handle connection management.

**File:** `internal/integrations/grpc/telemetry_client.go`
*   **Struct:** `TelemetryClient` holding the `grpc.ClientConn`.
*   **Method:** `SendTelemetry(ctx, data)` which maps `TelemetrySnapshot` to `proto.TelemetryRequest`.

#### 2.2 The Dispatcher
A dedicated goroutine in the `SimulationEngine` that aggregates channel data and flushes it to gRPC.

**File:** `internal/simulation/simulation-engine/dispatcher.go`
*   **Role:** Consumes from all `Vehicle.TelemetryChan`s.
*   **Strategy:** 
    *   Could have one dispatcher per vehicle (simple, more goroutines).
    *   Or a central "Telemetry Bus" channel that all vehicles write to (better scaling).
    *   **Selected:** Central buffered channel in `SimulationEngine`.
*   **Action:** When a snapshot arrives, convert to Proto and call `grpcClient.RecordTelemetry()`.

---

### Phase 3: Real-Time WebSockets

**Objective:** Stream position updates to the frontend @ 10Hz (100ms).

#### 3.1 WebSocket Hub
Standard Go Websocket hub pattern to manage connections.

**File:** `internal/websockets/hub.go`
*   **Struct:** `Hub`
*   **Fields:**
    *   `clients map[*Client]bool`
    *   `broadcast chan []byte` (The byte payload is the JSON batch)
*   **Methods:** `Register`, `Unregister`, `Run`.

#### 3.2 The Broadcaster Loop
The Simulation Engine needs a "global" ticker to gather state from all vehicles.

**File:** `internal/simulation/simulation-engine/simulation-engine.go`
*   **Add Field:** `BroadcastTicker *time.Ticker` (100ms).
*   **Update `Start()`:** Launch a `broadcaster` goroutine.
*   **Logic:**
    1.  Wait for 100ms tick.
    2.  `RLock` the engine vehicles map.
    3.  Iterate all active vehicles.
    4.  Create a lightweight struct: `type VisUpdate { ID, Lat, Lng, Heading }`.
    5.  Marshal array `[]VisUpdate` to JSON.
    6.  Send to `Hub.broadcast` channel.

---

### Phase 4: Simulation Manager

**Objective:** Control the lifecycle of vehicles dynamically via API.

#### 4.1 Manager Interface
**File:** `internal/simulation/manager.go`
```go
type Manager interface {
    StartSimulation(config SimConfig) error
    StopSimulation() error
    SpawnVehicle(config VehicleConfig) (string, error)
    RemoveVehicle(id string) error
}
```

#### 4.2 Integration with Handlers
*   **HTTP/gRPC Handlers** will accept external requests (e.g., `POST /simulation/vehicles`).
*   The handlers will invoke the `Manager` methods.
*   The `Manager` interacts directly with the `SimulationEngine`.

---

## 3. Implementation Steps Order

1.  **Infrastructure (Day 1):**
    *   Define `TelemeterySnapshot` structs.
    *   Create the `Hub` for WebSockets.
    *   Set up the `TelemetryClient` stub.

2.  **Engine Wiring (Day 2):**
    *   Add `telemetryTicker` to vehicle loop.
    *   Add `broadcastTicker` to engine start.
    *   Implement the `Dispatcher` to drain telemetry channels.

3.  **Data Flow (Day 3):**
    *   Connect Dispatcher to gRPC client.
    *   Connect Broadcaster to WebSocket Hub.

4.  **Control (Day 4):**
    *   Implement API handlers to call `AddVehicle`/`RemoveVehicle`.

---

## 4. Verification Strategy

### 4.1 Persistance Verification
**Action:** Run simulation for 60 seconds with 1 vehicle.
**Check:** Query TimescaleDB.
```sql
SELECT time, vehicle_id, latitude, longitude 
FROM telemetry 
WHERE vehicle_id = '...' 
ORDER BY time DESC;
```
**Expectation:** ~60 rows, timestamped 1 second apart.

### 4.2 Real-time Verification
**Action:** Connect a simple WS client (e.g., Postman or simple HTML/JS page).
**Check:** Observe message frequency.
**Expectation:** Batches of JSON arrays arriving ~10 times per second.

### 4.3 Simulation Verification
**Action:** Call `POST /vehicles` then `DELETE /vehicles/{id}`.
**Check:** Logs showing "Starting goroutine" and "Stopping goroutine", and "Route Complete" events.
