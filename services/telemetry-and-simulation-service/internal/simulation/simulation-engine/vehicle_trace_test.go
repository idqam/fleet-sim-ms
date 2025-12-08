package simulationengine

import (
	"math"
	"testing"

	"github.com/m/internal/simulation/entities"
)

func TestTraceVehiclePosition_20TimeUnits(t *testing.T) {
	nodeA := &entities.MapNode{ID: "A", Position: entities.Vector2D{X: 0, Y: 0}}
	nodeB := &entities.MapNode{ID: "B", Position: entities.Vector2D{X: 100, Y: 0}}
	nodeC := &entities.MapNode{ID: "C", Position: entities.Vector2D{X: 200, Y: 0}}

	edgeAB := &entities.MapEdge{
		ID:             "A-B",
		From:           "A",
		To:             "B",
		Length:         100.0,
		BaseSpeedLimit: 10.0,
		Conditions:     &entities.RoadConditions{EffectiveSpeedLimit: 10.0},
	}
	edgeBC := &entities.MapEdge{
		ID:             "B-C",
		From:           "B",
		To:             "C",
		Length:         100.0,
		BaseSpeedLimit: 10.0,
		Conditions:     &entities.RoadConditions{EffectiveSpeedLimit: 10.0},
	}

	graph := &entities.MapGraph{
		Nodes: map[string]*entities.MapNode{"A": nodeA, "B": nodeB, "C": nodeC},
		Edges: map[string]*entities.MapEdge{"A-B": edgeAB, "B-C": edgeBC},
	}

	vehicle := &entities.Vehicle{
		ID: "tracer",
		Route: &entities.AssignedRoute{
			Edges:            []string{"A-B", "B-C"},
			CurrentEdgeIndex: 0,
			CurrentNode:      "A",
			TargetNode:       "B",
			EndNode:          "C",
		},
		State: entities.VehicleState{
			CurrentPosition: entities.Vector2D{X: 0, Y: 0},
			CurrentEdge:     "A-B",
			Status:          entities.VehicleStatusMoving,
		},
	}

	t.Logf("Starting Trace for Vehicle %s (Linear X)", vehicle.ID)
	t.Logf("Route: A(0,0) -> B(100,0) -> C(200,0)")
	t.Logf("%-5s | %-10s | %-8s | %s", "Time", "Edge", "Progress", "Position")
	t.Logf("---------------------------------------------------------")

	t.Logf("%-5d | %-10s | %-8.2f | (%.2f, %.2f)",
		0, vehicle.State.CurrentEdge, vehicle.State.ProgressOnEdge,
		vehicle.State.CurrentPosition.X, vehicle.State.CurrentPosition.Y)

	for i := 1; i <= 20; i++ {
		err := UpdateVehiclePosition(vehicle, graph, 1.0)
		if err != nil {
			t.Fatalf("Error at time %d: %v", i, err)
		}

		edge := vehicle.State.CurrentEdge
		if vehicle.State.Status == entities.VehicleStatusArrived {
			edge = "ARRIVED"
		}

		t.Logf("%-5d | %-10s | %-8.2f | (%.2f, %.2f)",
			i, edge, vehicle.State.ProgressOnEdge,
			vehicle.State.CurrentPosition.X, vehicle.State.CurrentPosition.Y)
	}
}

func TestTraceVehiclePosition_Diagonal(t *testing.T) {

	nodeA := &entities.MapNode{ID: "A", Position: entities.Vector2D{X: 0, Y: 0}}
	nodeB := &entities.MapNode{ID: "B", Position: entities.Vector2D{X: 60, Y: 80}}
	nodeC := &entities.MapNode{ID: "C", Position: entities.Vector2D{X: 60, Y: 180}}

	edgeAB := &entities.MapEdge{
		ID:             "A-B",
		From:           "A",
		To:             "B",
		Length:         100.0,
		BaseSpeedLimit: 10.0,
		Conditions:     &entities.RoadConditions{EffectiveSpeedLimit: 10.0},
	}
	edgeBC := &entities.MapEdge{
		ID:             "B-C",
		From:           "B",
		To:             "C",
		Length:         100.0,
		BaseSpeedLimit: 10.0,
		Conditions:     &entities.RoadConditions{EffectiveSpeedLimit: 10.0},
	}

	graph := &entities.MapGraph{
		Nodes: map[string]*entities.MapNode{"A": nodeA, "B": nodeB, "C": nodeC},
		Edges: map[string]*entities.MapEdge{"A-B": edgeAB, "B-C": edgeBC},
	}

	vehicle := &entities.Vehicle{
		ID: "tracer_diag",
		Route: &entities.AssignedRoute{
			Edges:            []string{"A-B", "B-C"},
			CurrentEdgeIndex: 0,
			CurrentNode:      "A",
			TargetNode:       "B",
			EndNode:          "C",
		},
		State: entities.VehicleState{
			CurrentPosition: entities.Vector2D{X: 0, Y: 0},
			CurrentEdge:     "A-B",
			Status:          entities.VehicleStatusMoving,
		},
	}

	t.Logf("Starting Trace for Vehicle %s (Diagonal & Vertical)", vehicle.ID)
	t.Logf("Route: A(0,0) -> B(60,80) -> C(60,180)")
	t.Logf("%-5s | %-10s | %-8s | %s", "Time", "Edge", "Progress", "Position")
	t.Logf("---------------------------------------------------------")

	t.Logf("%-5d | %-10s | %-8.2f | (%.2f, %.2f)",
		0, vehicle.State.CurrentEdge, vehicle.State.ProgressOnEdge,
		vehicle.State.CurrentPosition.X, vehicle.State.CurrentPosition.Y)

	for i := 1; i <= 20; i++ {
		err := UpdateVehiclePosition(vehicle, graph, 1.0)
		if err != nil {
			t.Fatalf("Error at time %d: %v", i, err)
		}

		edge := vehicle.State.CurrentEdge
		if vehicle.State.Status == entities.VehicleStatusArrived {
			edge = "ARRIVED"
		}

		t.Logf("%-5d | %-10s | %-8.2f | (%.2f, %.2f)",
			i, edge, vehicle.State.ProgressOnEdge,
			vehicle.State.CurrentPosition.X, vehicle.State.CurrentPosition.Y)

		if i == 5 {

			if math.Abs(vehicle.State.CurrentPosition.X-30.0) > 0.1 || math.Abs(vehicle.State.CurrentPosition.Y-40.0) > 0.1 {
				t.Errorf("Time 5: Expected (30, 40), got (%.2f, %.2f)", vehicle.State.CurrentPosition.X, vehicle.State.CurrentPosition.Y)
			}
		}
	}
}
