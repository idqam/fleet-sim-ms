package simulationengine

import (
	"fmt"
	"testing"
	"time"

	"github.com/m/internal/simulation/entities"
	"github.com/stretchr/testify/assert"
)

func TestSimulationEngine_Lifecycle(t *testing.T) {
	graph := &entities.MapGraph{
		Nodes: make(map[string]*entities.MapNode),
		Edges: make(map[string]*entities.MapEdge),
	}
	engine := NewSimulationEngine(graph, 100*time.Millisecond)

	engine.Start()
	assert.True(t, engine.IsRunning)

	v := &entities.Vehicle{ID: "v1"}
	engine.AddVehicle(v)

	engine.mutex.RLock()
	storedV, exists := engine.Vehicles["v1"]
	engine.mutex.RUnlock()
	assert.True(t, exists)
	assert.Equal(t, v, storedV)
	assert.NotNil(t, v.StopChan)

	engine.RemoveVehicle("v1")
	engine.mutex.RLock()
	_, exists = engine.Vehicles["v1"]
	engine.mutex.RUnlock()
	assert.False(t, exists)

	engine.Stop()
	assert.False(t, engine.IsRunning)
}

func TestSimulationEngine_ConcurrentVehicles(t *testing.T) {
	nodeA := &entities.MapNode{ID: "A", Position: entities.Vector2D{X: 0, Y: 0}}
	nodeB := &entities.MapNode{ID: "B", Position: entities.Vector2D{X: 1000, Y: 0}}
	edge := &entities.MapEdge{ID: "A-B", From: "A", To: "B", Length: 1000, Conditions: &entities.RoadConditions{EffectiveSpeedLimit: 100}}

	graph := &entities.MapGraph{
		Nodes: map[string]*entities.MapNode{"A": nodeA, "B": nodeB},
		Edges: map[string]*entities.MapEdge{"A-B": edge},
	}

	engine := NewSimulationEngine(graph, 10*time.Millisecond)
	engine.Start()
	defer engine.Stop()

	vehicleCount := 50

	for i := 0; i < vehicleCount; i++ {
		id := fmt.Sprintf("v%d", i)
		v := &entities.Vehicle{
			ID: id,
			State: entities.VehicleState{
				CurrentPosition: entities.Vector2D{X: 0, Y: 0},
				Status:          entities.VehicleStatusIdle,
			},
			Route: &entities.AssignedRoute{
				Edges:            []string{"A-B"},
				CurrentEdgeIndex: 0,
				StartNode:        "A",
				EndNode:          "B",
				CurrentNode:      "A",
				TargetNode:       "B",
			},
		}
		engine.AddVehicle(v)
	}

	time.Sleep(50 * time.Millisecond)

	engine.mutex.RLock()
	for _, v := range engine.Vehicles {
		v.Mutex.Lock()
		assert.True(t, v.State.ProgressOnEdge > 0 || v.State.Status == entities.VehicleStatusMoving)
		v.Mutex.Unlock()
	}
	engine.mutex.RUnlock()
}
