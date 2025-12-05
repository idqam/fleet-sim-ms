# Routing Architecture Design

## Overview

This document outlines the routing system architecture for FleetSim, a distributed event-driven fleet simulation platform. The design emphasizes separation of concerns, computational efficiency, and real-time adaptability.

---

## Core Design Principles

### Separation of Concerns

The routing system is built on a fundamental separation between **data** and **algorithms**:

**Map Generation** (`map-generator.go`)
- **Responsibility**: Create graph structures (nodes + edges)
- **Output**: Static data representing the road network
- **Analogy**: Building the road network of a city

**Pathfinding/Routing** (`map-routing.go`)
- **Responsibility**: Find paths through existing graphs
- **Input**: Graph data + start/end points
- **Analogy**: GPS navigation on existing roads

### Why Separate?

**Data vs. Algorithm Independence:**
- The same map can be searched using different algorithms (Dijkstra, A*, BFS, DFS)
- The same algorithm can work on maps from different generators (RGG, KNN, Delaunay)

**Single Responsibility Principle:**
- `MapGeneratorConfig`: "How do I create a graph?"
- `RoutingConfig`: "How do I search through a graph?"

**Composability:**
- Mix and match: Delaunay map + Dijkstra routing, or KNN map + A* routing
- Each component does one thing well

**Key Insight:** Routing algorithms are **stateless transformations** applied to data, not containers for data. Similar to how:
- A sorting algorithm doesn't own the array it sorts
- A search algorithm doesn't own the data structure it searches
- A compression algorithm doesn't own the file it compresses

---

## Architectural Components

### 1. RoutingConfig Structure

```
RoutingConfig:
  - Algo: Algorithm (enum: djk, bia, dfs, bfs)
```

**Design Decision:** `RoutingConfig` stores only the algorithm choice, not the graph data.

**Rationale:**
- Configuration is lightweight and portable
- Graph data (edges) is passed as parameters to routing functions
- Enables reuse of the same config across different graphs

### 2. Routing Function Signature Pattern

```
func AlgorithmName(graph *MapGraph, start, end string) *Route
```

**Parameters:**
- `graph`: The complete graph structure with edges
- `start`, `end`: Node IDs for route calculation

**Returns:**
- `Route`: Complete path with metadata

**Benefits:**
- Algorithms are pure functions (no side effects)
- Easy to test with mock data
- Thread-safe by default

---

## Event-Driven Routing Architecture

### Concept: Reactive Pathfinding

Instead of computing routes once, the system **reacts to events** that invalidate current paths.

### State Management

**Per-Vehicle State:**
- `CurrentPath`: List of nodes representing planned route
- `CurrentPosition`: Vehicle's position on the path
- `PathVersion`: Timestamp/version for cache invalidation

**Global State:**
- `WorldState`: Traffic conditions, road closures, weather
- `EventQueue`: Pending events that may affect routes

### Event Types

**Static Events** (Permanent Changes)
- Road closures
- Construction zones
- Map topology changes

**Dynamic Events** (Temporary Changes)
- Traffic congestion
- Accidents
- Weather conditions

**Vehicle Events**
- Destination changes
- Fuel low (need detour to station)
- Passenger pickup/dropoff

### Routing Strategy: Local Rerouting

**Concept:** Only recalculate the affected portion of the path.

**Process:**
1. Event occurs (e.g., road closure)
2. Define "radius of influence" around the event
3. Identify vehicles within this radius
4. Trigger reroute only for affected vehicles
5. Use incremental algorithms that reuse previous computation

**Efficiency Gains:**
- **Complexity**: O(affected vehicles) instead of O(all vehicles)
- **Realism**: Vehicles react to real-time conditions like real drivers
- **Scalability**: System can handle large fleets efficiently

---

## Algorithm Selection Strategy

### Hybrid Approach: Static Precomputation + Dynamic Adaptation

#### Phase 1: Initialization (Precomputation)

**Algorithms:** Dijkstra or Bidirectional A*

**When:**
- Simulation start
- Vehicle receives new destination
- Complete map refresh

**Process:**
- Compute full optimal path
- Store in vehicle state
- Transition to FOLLOWING state

**Characteristics:**
- Complexity: O(E + V log V)
- Optimal on static graphs
- Simple implementation
- No state management overhead

#### Phase 2: Runtime (Event-Driven)

**Algorithms:** D* Lite or LPA* (Lifelong Planning A*)

**When:**
- Event affects current path
- Map topology changes
- Edge weights update (traffic)

**Process:**
- Detect event impact on current path
- Run incremental algorithm on affected region
- Update path seamlessly
- Continue FOLLOWING state

**Characteristics:**
- Complexity: O(affected nodes)
- Reuses previous computation
- Maintains search tree
- Efficient for dynamic environments

### Vehicle State Machine

```
┌──────────┐
│ PLANNING │ ──► Run Dijkstra/Bi-A*
└────┬─────┘     Store complete path
     │
     ▼
┌───────────┐
│ FOLLOWING │ ──► Move along precomputed path
└────┬──────┘     Listen for events
     │
     │ (Event affects path)
     ▼
┌────────────┐
│ REROUTING  │ ──► Run D* Lite on affected region
└────┬───────┘     Update path
     │
     └──► Back to FOLLOWING
```

### Event Classification & Decision Logic

**Minor Events** (Continue with current path)
- Small traffic fluctuations
- Events far from vehicle's route
- Temporary delays that will clear before arrival

**Major Events** (Trigger D* Lite rerouting)
- Road closures on current path
- Severe traffic on upcoming segments
- Destination changes
- Large-scale map updates

**Decision Algorithm:**
```
if event.affectsCurrentPath(vehicle.path) AND event.severity > threshold:
    switch to D* Lite rerouting
else:
    continue following current path
```

---

## Multi-Layer Routing System

### Layer 1: Backend Route Planner

**Purpose:** Precomputation and complex route optimization

**Algorithms:**

**Contraction Hierarchies (CCH)**
- **Use Case**: Static maps with millions of queries
- **Preprocessing**: Build hierarchy of "important" nodes
- **Query Time**: Search only the hierarchy (much smaller graph)
- **Trade-off**: Slow preprocessing, lightning-fast queries
- **Mental Model**: Building an express highway system on top of local roads

**K-Shortest Paths**
- **Use Case**: Alternative route generation
- **Output**: Top K alternative routes, not just optimal
- **Applications**:
  - Show users multiple route options
  - Fallback routes when primary is blocked
  - Fleet load balancing (avoid all vehicles on same route)

### Layer 2: Core Simulation Engine

**Purpose:** Real-time pathfinding during simulation

**Algorithms:**

**Dijkstra's Algorithm**
- **Use Case**: Initial path computation, fresh queries
- **Characteristics**: Simple, optimal, well-understood
- **Complexity**: O(E + V log V)

**Bidirectional A***
- **Use Case**: Faster initial computation with heuristics
- **Characteristics**: Searches from both start and goal
- **Complexity**: Better than Dijkstra in practice

**D* Lite**
- **Use Case**: Map structure changes (road closures)
- **Characteristics**: Incremental, maintains search tree
- **Complexity**: O(affected nodes)

**Incremental Dijkstra**
- **Use Case**: Edge weight changes (traffic updates)
- **Characteristics**: Propagates weight changes efficiently
- **Complexity**: O(affected edges)

### Layer 3: Frontend Visualization

**Purpose:** Smooth, real-time path rendering

**Strategy:** Precomputed + Dynamic Updates

**Phase 1: Precomputation (Offline)**
- Calculate base paths for common routes
- Store in cache/database
- Use fast algorithms (CCH, Arc Flags)

**Phase 2: Runtime Adjustments (Online)**
- Start with precomputed path
- Apply incremental updates based on current conditions
- Use D* Lite for map changes
- Use Incremental Dijkstra for edge weight changes

**Why Separate?**
- Frontend needs sub-millisecond response for smooth animation
- Backend can afford seconds/minutes for preprocessing
- Hybrid approach: Fast initial render + real-time corrections

---

## System Architecture Diagram

```
┌─────────────────────────────────────────────────┐
│  Event System (Redis Streams/Pub-Sub)          │
│  - Road closures, traffic updates, alerts      │
└──────────────────┬──────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────┐
│  Routing Coordinator                            │
│  - Decides which vehicles need reroute          │
│  - Manages event → algorithm mapping            │
│  - Throttles reroute requests                   │
└──────────────────┬──────────────────────────────┘
                   │
            ┌──────┴──────┐
            ▼             ▼
    ┌─────────────┐ ┌──────────────┐
    │ Static      │ │ Dynamic      │
    │ Algorithms  │ │ Algorithms   │
    │ - Dijkstra  │ │ - D* Lite    │
    │ - A*        │ │ - LPA*       │
    │ - CCH       │ │ - Incremental│
    └─────────────┘ └──────────────┘
            │             │
            └──────┬──────┘
                   ▼
┌─────────────────────────────────────────────────┐
│  Path Cache / State Manager                     │
│  - Current paths for all vehicles               │
│  - Precomputed routes                           │
│  - Search tree state (for D* Lite)              │
└─────────────────────────────────────────────────┘
```

---

## Algorithm Selection Decision Tree

```
New route request?
├─ Yes → Use Dijkstra or Bi-A* (precomputation)
└─ No → Event occurred?
    ├─ Map structure changed? → Use D* Lite
    ├─ Edge weights changed? → Use Incremental Dijkstra
    ├─ Need alternatives? → Use K-Shortest Paths
    └─ Fast query needed? → Use CCH (if preprocessed)
```

---

## Benefits of Hybrid Architecture

### 1. Simplicity First
- Start with simple algorithms (Dijkstra)
- Add complexity (D* Lite) only when needed
- Easier to debug and maintain
- Lower barrier to entry for new developers

### 2. Graceful Degradation
- If D* Lite fails, fall back to Dijkstra
- System still works, just less efficiently
- No single point of failure

### 3. Incremental Adoption
- Deploy with just Dijkstra initially
- Add event-driven routing as simulation matures
- No need to build everything at once
- Validate each layer independently

### 4. Resource Management
- **Dijkstra**: Stateless, low memory footprint
- **D* Lite**: Stateful, higher memory but only for actively rerouting vehicles
- **Most vehicles most of the time**: Just following precomputed paths (cheap)

### 5. Computational Efficiency

**Cold Start (No Prior Knowledge):**
- Pay O(E + V log V) cost once per route
- Acceptable for initial computation

**Hot Updates (Path Already Exists):**
- D* Lite: O(affected nodes) - much faster than full recomputation
- Most events affect only a small portion of the map

---

## Real-World Analogy: GPS Navigation

**Before Driving (Dijkstra/Bi-A*):**
- GPS calculates the full route
- Shows entire path
- Optimizes for current conditions

**While Driving (Event-Driven):**
- GPS monitors traffic in real-time
- Only recalculates if there's a problem ahead
- Keeps most of route unchanged
- Updates only affected portions

**Key Insight:** You don't recalculate your entire route every second—you only reroute when something significant changes on your current path.

---

## Implementation Guidelines

### Data Structure Compatibility

**Graph Structure:**
```
MapGraph:
  - Nodes: map[string]*MapNode
  - Edges: map[string]*MapEdge

MapNode:
  - ID: string
  - Position: Point
  - Connections: map[string]bool

MapEdge:
  - ID: string
  - From: string
  - To: string
  - Distance: float64
```

**Route Structure:**
```
Route:
  - Path: []string (node IDs)
  - TotalDistance: float64
  - EstimatedTime: float64
  - Algorithm: Algorithm (which algo generated this)
```

### Algorithm Interface

All routing algorithms should conform to a common interface:

```
type RoutingAlgorithm interface {
    FindPath(graph *MapGraph, start, end string) *Route
}
```

**Benefits:**
- Algorithms are interchangeable
- Easy to add new algorithms
- Vehicle code doesn't care which algorithm is used
- Transparent switching between algorithms

### Event Handling Pattern

**Event Structure:**
```
RoutingEvent:
  - Type: EventType (closure, traffic, etc.)
  - Location: Point or EdgeID
  - Severity: float64
  - Radius: float64 (area of effect)
  - Timestamp: time.Time
```

**Event Processing:**
1. Event published to Redis stream
2. Routing Coordinator receives event
3. Query spatial index for affected vehicles
4. For each affected vehicle:
   - Check if event impacts current path
   - If yes, trigger appropriate rerouting algorithm
   - Update vehicle's path
5. Publish path updates back to event system

---

## Performance Considerations

### Spatial Indexing
- Use R-tree or quadtree for fast "vehicles near event" queries
- O(log n) lookup instead of O(n) linear search
- Critical for large fleets (1000+ vehicles)

### Caching Strategy
- Cache precomputed routes for common origin-destination pairs
- Invalidate cache entries when map changes
- LRU eviction for memory management

### Parallelization
- Route computations are embarrassingly parallel
- Use worker pool for concurrent route calculations
- Goroutines per vehicle for independent processing

### Throttling
- Limit reroute requests per vehicle (e.g., max 1 per second)
- Batch similar events to reduce duplicate work
- Prioritize critical events (road closures > minor traffic)

---

## Future Enhancements

### Advanced Algorithms
- **Arc Flags**: Preprocessing for regional routing
- **Hub Labeling**: Ultra-fast distance queries
- **Time-Dependent Routing**: Account for predictable traffic patterns

### Machine Learning Integration
- Predict traffic patterns from historical data
- Learn optimal rerouting thresholds
- Adaptive algorithm selection based on map characteristics

### Multi-Objective Optimization
- Optimize for distance, time, fuel, and comfort simultaneously
- Pareto-optimal route sets
- User preference weighting

---

## References & Further Reading

### Algorithms
- **Dijkstra's Algorithm**: Classic shortest path
- **A* Search**: Heuristic-guided pathfinding
- **D* Lite**: Incremental replanning for dynamic environments
- **Contraction Hierarchies**: Fast routing on road networks
- **K-Shortest Paths**: Alternative route generation

### Design Patterns
- **Strategy Pattern**: Interchangeable algorithms
- **State Pattern**: Vehicle routing states
- **Observer Pattern**: Event-driven updates
- **Command Pattern**: Route requests as objects

---

## Conclusion

This routing architecture balances **simplicity**, **performance**, and **flexibility**. By separating concerns, using hybrid algorithms, and embracing event-driven design, FleetSim can efficiently simulate large fleets in dynamic environments while maintaining code clarity and extensibility.

The key insight: **Start simple (Dijkstra), add sophistication where it provides value (D* Lite for events), and keep data separate from algorithms.**
