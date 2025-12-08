package simulationengine

import (
	"math"
	"testing"
	"time"

	"github.com/m/internal/simulation/entities"
)

func TestUpdateVehiclePosition_NilRoute(t *testing.T) {
	vehicle := &entities.Vehicle{
		ID:    "test-vehicle",
		Route: nil,
	}

	graph := &entities.MapGraph{}

	err := UpdateVehiclePosition(vehicle, graph, 1.0)
	if err != nil {
		t.Errorf("Expected no error for nil route, got: %v", err)
	}
}

func TestUpdateVehiclePosition_CompletedRoute(t *testing.T) {
	now := time.Now()
	vehicle := &entities.Vehicle{
		ID: "test-vehicle",
		Route: &entities.AssignedRoute{
			CompletedAt: &now,
		},
		State: entities.VehicleState{},
	}

	graph := &entities.MapGraph{}

	err := UpdateVehiclePosition(vehicle, graph, 1.0)
	if err != nil {
		t.Errorf("Expected no error for completed route, got: %v", err)
	}

	if vehicle.State.Status != entities.VehicleStatusArrived {
		t.Errorf("Expected status Arrived, got: %v", vehicle.State.Status)
	}
}

func TestUpdateVehiclePosition_SingleEdge_PartialProgress(t *testing.T) {
	nodeA := &entities.MapNode{
		ID:       "A",
		Position: entities.Vector2D{X: 0, Y: 0},
	}
	nodeB := &entities.MapNode{
		ID:       "B",
		Position: entities.Vector2D{X: 100, Y: 0},
	}

	edge := &entities.MapEdge{
		ID:             "A-B",
		From:           "A",
		To:             "B",
		Length:         100.0,
		BaseSpeedLimit: 10.0,
		Conditions: &entities.RoadConditions{
			EffectiveSpeedLimit: 10.0,
		},
	}

	graph := &entities.MapGraph{
		Nodes: map[string]*entities.MapNode{
			"A": nodeA,
			"B": nodeB,
		},
		Edges: map[string]*entities.MapEdge{
			"A-B": edge,
		},
	}

	vehicle := &entities.Vehicle{
		ID: "test-vehicle",
		Route: &entities.AssignedRoute{
			Edges:            []string{"A-B"},
			CurrentEdgeIndex: 0,
			CurrentNode:      "A",
			TargetNode:       "B",
			EndNode:          "B",
		},
		State: entities.VehicleState{
			ProgressOnEdge: 0.0,
		},
	}

	err := UpdateVehiclePosition(vehicle, graph, 1.0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedProgress := 0.1
	if vehicle.State.ProgressOnEdge != expectedProgress {
		t.Errorf("Expected progress %.2f, got %.2f", expectedProgress, vehicle.State.ProgressOnEdge)
	}

	expectedX := 10.0
	if vehicle.State.CurrentPosition.X != expectedX {
		t.Errorf("Expected X position %.2f, got %.2f", expectedX, vehicle.State.CurrentPosition.X)
	}

	if vehicle.State.Status != entities.VehicleStatusMoving {
		t.Errorf("Expected status Moving, got: %v", vehicle.State.Status)
	}
}

func TestUpdateVehiclePosition_SingleEdge_Complete(t *testing.T) {
	nodeA := &entities.MapNode{
		ID:       "A",
		Position: entities.Vector2D{X: 0, Y: 0},
	}
	nodeB := &entities.MapNode{
		ID:       "B",
		Position: entities.Vector2D{X: 100, Y: 0},
	}

	edge := &entities.MapEdge{
		ID:             "A-B",
		From:           "A",
		To:             "B",
		Length:         100.0,
		BaseSpeedLimit: 10.0,
		Conditions: &entities.RoadConditions{
			EffectiveSpeedLimit: 10.0,
		},
	}

	graph := &entities.MapGraph{
		Nodes: map[string]*entities.MapNode{
			"A": nodeA,
			"B": nodeB,
		},
		Edges: map[string]*entities.MapEdge{
			"A-B": edge,
		},
	}

	vehicle := &entities.Vehicle{
		ID: "test-vehicle",
		Route: &entities.AssignedRoute{
			Edges:            []string{"A-B"},
			CurrentEdgeIndex: 0,
			CurrentNode:      "A",
			TargetNode:       "B",
			EndNode:          "B",
		},
		State: entities.VehicleState{
			ProgressOnEdge: 0.0,
		},
	}

	for i := 0; i < 10; i++ {
		err := UpdateVehiclePosition(vehicle, graph, 1.0)
		if err != nil {
			t.Fatalf("Unexpected error on iteration %d: %v", i, err)
		}
	}

	if vehicle.Route.CompletedAt == nil {
		t.Error("Expected route to be completed")
	}

	if vehicle.State.Status != entities.VehicleStatusArrived {
		t.Errorf("Expected status Arrived, got: %v", vehicle.State.Status)
	}

	if vehicle.State.CurrentPosition.X != 100.0 {
		t.Errorf("Expected final X position 100.0, got %.2f", vehicle.State.CurrentPosition.X)
	}

	if vehicle.State.Velocity.X != 0 || vehicle.State.Velocity.Y != 0 {
		t.Errorf("Expected zero velocity, got (%.2f, %.2f)", vehicle.State.Velocity.X, vehicle.State.Velocity.Y)
	}
}

func TestUpdateVehiclePosition_MultiEdge_Transition(t *testing.T) {
	nodeA := &entities.MapNode{
		ID:       "A",
		Position: entities.Vector2D{X: 0, Y: 0},
	}
	nodeB := &entities.MapNode{
		ID:       "B",
		Position: entities.Vector2D{X: 100, Y: 0},
	}
	nodeC := &entities.MapNode{
		ID:       "C",
		Position: entities.Vector2D{X: 100, Y: 100},
	}

	edgeAB := &entities.MapEdge{
		ID:             "A-B",
		From:           "A",
		To:             "B",
		Length:         100.0,
		BaseSpeedLimit: 10.0,
		Conditions: &entities.RoadConditions{
			EffectiveSpeedLimit: 10.0,
		},
	}

	edgeBC := &entities.MapEdge{
		ID:             "B-C",
		From:           "B",
		To:             "C",
		Length:         100.0,
		BaseSpeedLimit: 10.0,
		Conditions: &entities.RoadConditions{
			EffectiveSpeedLimit: 10.0,
		},
	}

	graph := &entities.MapGraph{
		Nodes: map[string]*entities.MapNode{
			"A": nodeA,
			"B": nodeB,
			"C": nodeC,
		},
		Edges: map[string]*entities.MapEdge{
			"A-B": edgeAB,
			"B-C": edgeBC,
		},
	}

	vehicle := &entities.Vehicle{
		ID: "test-vehicle",
		Route: &entities.AssignedRoute{
			Edges:            []string{"A-B", "B-C"},
			CurrentEdgeIndex: 0,
			CurrentNode:      "A",
			TargetNode:       "B",
			EndNode:          "C",
		},
		State: entities.VehicleState{
			ProgressOnEdge: 0.0,
			CurrentEdge:    "A-B",
		},
	}

	for i := 0; i < 10; i++ {
		err := UpdateVehiclePosition(vehicle, graph, 1.0)
		if err != nil {
			t.Fatalf("Unexpected error on iteration %d: %v", i, err)
		}
	}

	if vehicle.Route.CurrentEdgeIndex != 1 {
		t.Errorf("Expected CurrentEdgeIndex 1, got %d", vehicle.Route.CurrentEdgeIndex)
	}

	if vehicle.Route.CurrentNode != "B" {
		t.Errorf("Expected CurrentNode B, got %s", vehicle.Route.CurrentNode)
	}

	if vehicle.Route.TargetNode != "C" {
		t.Errorf("Expected TargetNode C, got %s", vehicle.Route.TargetNode)
	}

	if vehicle.State.ProgressOnEdge != 0.0 {
		t.Errorf("Expected ProgressOnEdge reset to 0.0, got %.2f", vehicle.State.ProgressOnEdge)
	}

	if vehicle.State.CurrentEdge != "B-C" {
		t.Errorf("Expected CurrentEdge B-C, got %s", vehicle.State.CurrentEdge)
	}

	for i := 0; i < 10; i++ {
		err := UpdateVehiclePosition(vehicle, graph, 1.0)
		if err != nil {
			t.Fatalf("Unexpected error on iteration %d: %v", i, err)
		}
	}

	if vehicle.Route.CompletedAt == nil {
		t.Error("Expected route to be completed")
	}

	if vehicle.State.Status != entities.VehicleStatusArrived {
		t.Errorf("Expected status Arrived, got: %v", vehicle.State.Status)
	}
}

func TestUpdateVehiclePosition_EdgeNotFound(t *testing.T) {
	graph := &entities.MapGraph{
		Nodes: map[string]*entities.MapNode{},
		Edges: map[string]*entities.MapEdge{},
	}

	vehicle := &entities.Vehicle{
		ID: "test-vehicle",
		Route: &entities.AssignedRoute{
			Edges:            []string{"nonexistent-edge"},
			CurrentEdgeIndex: 0,
		},
		State: entities.VehicleState{},
	}

	err := UpdateVehiclePosition(vehicle, graph, 1.0)
	if err == nil {
		t.Error("Expected error for nonexistent edge, got nil")
	}
}

func TestUpdateVehiclePosition_VelocityCalculation(t *testing.T) {
	nodeA := &entities.MapNode{
		ID:       "A",
		Position: entities.Vector2D{X: 0, Y: 0},
	}
	nodeB := &entities.MapNode{
		ID:       "B",
		Position: entities.Vector2D{X: 30, Y: 40},
	}

	edge := &entities.MapEdge{
		ID:             "A-B",
		From:           "A",
		To:             "B",
		Length:         50.0,
		BaseSpeedLimit: 10.0,
		Conditions: &entities.RoadConditions{
			EffectiveSpeedLimit: 10.0,
		},
	}

	graph := &entities.MapGraph{
		Nodes: map[string]*entities.MapNode{
			"A": nodeA,
			"B": nodeB,
		},
		Edges: map[string]*entities.MapEdge{
			"A-B": edge,
		},
	}

	vehicle := &entities.Vehicle{
		ID: "test-vehicle",
		Route: &entities.AssignedRoute{
			Edges:            []string{"A-B"},
			CurrentEdgeIndex: 0,
			CurrentNode:      "A",
			TargetNode:       "B",
			EndNode:          "B",
		},
		State: entities.VehicleState{
			ProgressOnEdge: 0.0,
		},
	}

	err := UpdateVehiclePosition(vehicle, graph, 1.0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedVelX := 6.0
	expectedVelY := 8.0

	tolerance := 0.001
	if math.Abs(vehicle.State.Velocity.X-expectedVelX) > tolerance {
		t.Errorf("Expected velocity X %.2f, got %.2f", expectedVelX, vehicle.State.Velocity.X)
	}
	if math.Abs(vehicle.State.Velocity.Y-expectedVelY) > tolerance {
		t.Errorf("Expected velocity Y %.2f, got %.2f", expectedVelY, vehicle.State.Velocity.Y)
	}
}

func TestUpdateVehiclePosition_UseBaseSpeedWhenEffectiveIsZero(t *testing.T) {
	nodeA := &entities.MapNode{
		ID:       "A",
		Position: entities.Vector2D{X: 0, Y: 0},
	}
	nodeB := &entities.MapNode{
		ID:       "B",
		Position: entities.Vector2D{X: 100, Y: 0},
	}

	edge := &entities.MapEdge{
		ID:             "A-B",
		From:           "A",
		To:             "B",
		Length:         100.0,
		BaseSpeedLimit: 15.0,
		Conditions: &entities.RoadConditions{
			EffectiveSpeedLimit: 0.0,
		},
	}

	graph := &entities.MapGraph{
		Nodes: map[string]*entities.MapNode{
			"A": nodeA,
			"B": nodeB,
		},
		Edges: map[string]*entities.MapEdge{
			"A-B": edge,
		},
	}

	vehicle := &entities.Vehicle{
		ID: "test-vehicle",
		Route: &entities.AssignedRoute{
			Edges:            []string{"A-B"},
			CurrentEdgeIndex: 0,
			CurrentNode:      "A",
			TargetNode:       "B",
			EndNode:          "B",
		},
		State: entities.VehicleState{
			ProgressOnEdge: 0.0,
		},
	}

	err := UpdateVehiclePosition(vehicle, graph, 1.0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedProgress := 0.15
	if vehicle.State.ProgressOnEdge != expectedProgress {
		t.Errorf("Expected progress %.2f, got %.2f", expectedProgress, vehicle.State.ProgressOnEdge)
	}
}

func TestInterpolatePosition(t *testing.T) {
	from := &entities.MapNode{
		Position: entities.Vector2D{X: 0, Y: 0},
	}
	to := &entities.MapNode{
		Position: entities.Vector2D{X: 100, Y: 50},
	}

	tests := []struct {
		progress  float64
		expectedX float64
		expectedY float64
	}{
		{0.0, 0.0, 0.0},
		{0.5, 50.0, 25.0},
		{1.0, 100.0, 50.0},
		{0.25, 25.0, 12.5},
	}

	for _, tt := range tests {
		result := interpolatePosition(from, to, tt.progress)
		if result.X != tt.expectedX || result.Y != tt.expectedY {
			t.Errorf("interpolatePosition(%.2f) = (%.2f, %.2f), expected (%.2f, %.2f)",
				tt.progress, result.X, result.Y, tt.expectedX, tt.expectedY)
		}
	}
}

func TestCalculateVelocity(t *testing.T) {
	from := &entities.MapNode{
		Position: entities.Vector2D{X: 0, Y: 0},
	}
	to := &entities.MapNode{
		Position: entities.Vector2D{X: 30, Y: 40},
	}

	speed := 10.0
	result := calculateVelocity(from, to, speed)

	expectedX := 6.0
	expectedY := 8.0

	tolerance := 0.001
	if math.Abs(result.X-expectedX) > tolerance || math.Abs(result.Y-expectedY) > tolerance {
		t.Errorf("calculateVelocity() = (%.2f, %.2f), expected (%.2f, %.2f)",
			result.X, result.Y, expectedX, expectedY)
	}
}

func TestCalculateVelocity_ZeroDistance(t *testing.T) {
	from := &entities.MapNode{
		Position: entities.Vector2D{X: 10, Y: 20},
	}
	to := &entities.MapNode{
		Position: entities.Vector2D{X: 10, Y: 20},
	}

	speed := 10.0
	result := calculateVelocity(from, to, speed)

	if result.X != 0 || result.Y != 0 {
		t.Errorf("Expected zero velocity for zero distance, got (%.2f, %.2f)", result.X, result.Y)
	}
}
