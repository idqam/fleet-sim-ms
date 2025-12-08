package simulationengine_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/m/internal/simulation/entities"
	simulationengine "github.com/m/internal/simulation/simulation-engine"
)

func TestVehicleRouteAssignment_Random(t *testing.T) {
	config := simulationengine.NewMapGenerator(1000, 1000, 42, simulationengine.AlgoKNN, 20, 4)
	graph := config.Generate()

	spawnConfig := &simulationengine.VehicleSpawnConfig{
		SpawnStrategy:  simulationengine.SpawnRandom,
		TargetStrategy: simulationengine.TargetRandom,
		AllowSameNode:  false,
	}

	vehicle := &entities.Vehicle{
		ID:   uuid.New().String(),
		Type: entities.VehicleTypSedan,
		State: entities.VehicleState{
			Status: entities.VehicleStatusIdle,
		},
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("TEST: Random Vehicle Route Assignment")
	fmt.Println(strings.Repeat("=", 80))

	err := simulationengine.AssignVehicleRoute(vehicle, graph, spawnConfig)
	if err != nil {
		t.Fatalf("Failed to assign route: %v", err)
	}

	fmt.Printf("Vehicle ID: %s\n", vehicle.ID)
	fmt.Printf("Spawn Node: %s\n", vehicle.Route.StartNode)
	fmt.Printf("Target Node: %s\n", vehicle.Route.EndNode)
	fmt.Printf("Current Node: %s\n", vehicle.Route.CurrentNode)
	fmt.Printf("Next Waypoint: %s\n", vehicle.Route.TargetNode)
	fmt.Printf("Route Edges: %d\n", len(vehicle.Route.Edges))
	fmt.Printf("Current Position: (%.2f, %.2f)\n", vehicle.State.CurrentPosition.X, vehicle.State.CurrentPosition.Y)
	fmt.Printf("Status: %s\n", vehicle.State.Status)
	fmt.Printf("Progress on Edge: %.2f\n", vehicle.State.ProgressOnEdge)

	fmt.Println(strings.Repeat("=", 80) + "\n")

	if vehicle.Route == nil {
		t.Error("Route should not be nil")
	}
	if vehicle.Route.StartNode == "" {
		t.Error("StartNode should not be empty")
	}
	if vehicle.Route.EndNode == "" {
		t.Error("EndNode should not be empty")
	}
	if vehicle.Route.CurrentNode != vehicle.Route.StartNode {
		t.Errorf("CurrentNode should equal StartNode, got %s != %s", vehicle.Route.CurrentNode, vehicle.Route.StartNode)
	}
	if vehicle.State.ProgressOnEdge != 0.0 {
		t.Errorf("ProgressOnEdge should be 0.0, got %.2f", vehicle.State.ProgressOnEdge)
	}
}

func TestVehicleRouteAssignment_Farthest(t *testing.T) {
	config := simulationengine.NewMapGenerator(1000, 1000, 123, simulationengine.AlgoDelaunay, 15, 0)
	graph := config.Generate()

	spawnConfig := &simulationengine.VehicleSpawnConfig{
		SpawnStrategy:  simulationengine.SpawnRandom,
		TargetStrategy: simulationengine.TargetFarthest,
		AllowSameNode:  false,
	}

	vehicle := &entities.Vehicle{
		ID:   uuid.New().String(),
		Type: entities.VehicleTypeTruck,
		State: entities.VehicleState{
			Status: entities.VehicleStatusIdle,
		},
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("TEST: Farthest Target Route Assignment")
	fmt.Println(strings.Repeat("=", 80))

	err := simulationengine.AssignVehicleRoute(vehicle, graph, spawnConfig)
	if err != nil {
		t.Fatalf("Failed to assign route: %v", err)
	}

	spawnPos := graph.Nodes[vehicle.Route.StartNode].Position
	targetPos := graph.Nodes[vehicle.Route.EndNode].Position

	dx := targetPos.X - spawnPos.X
	dy := targetPos.Y - spawnPos.Y
	straightLineDist := (dx*dx + dy*dy)

	fmt.Printf("Vehicle ID: %s\n", vehicle.ID)
	fmt.Printf("Spawn: %s at (%.0f, %.0f)\n", vehicle.Route.StartNode, spawnPos.X, spawnPos.Y)
	fmt.Printf("Target: %s at (%.0f, %.0f)\n", vehicle.Route.EndNode, targetPos.X, targetPos.Y)
	fmt.Printf("Straight-line distance²: %.2f\n", straightLineDist)
	fmt.Printf("Route edges: %d\n", len(vehicle.Route.Edges))
	fmt.Printf("Status: %s\n", vehicle.State.Status)

	fmt.Println(strings.Repeat("=", 80) + "\n")

	if len(vehicle.Route.Edges) == 0 {
		t.Error("Route should have edges for farthest target")
	}
}

func TestVehicleRouteAssignment_SpecificNodes(t *testing.T) {
	config := simulationengine.NewMapGenerator(800, 800, 456, simulationengine.AlgoRGG, 25, 0)
	config.RadiusMode = simulationengine.Connected
	graph := config.Generate()

	nodeIDs := make([]string, 0, len(graph.Nodes))
	for id := range graph.Nodes {
		nodeIDs = append(nodeIDs, id)
		if len(nodeIDs) >= 2 {
			break
		}
	}

	if len(nodeIDs) < 2 {
		t.Skip("Not enough nodes in graph")
	}

	startNode := nodeIDs[0]
	endNode := nodeIDs[1]

	vehicle := &entities.Vehicle{
		ID:   uuid.New().String(),
		Type: entities.VehicleTypSedan,
		State: entities.VehicleState{
			Status: entities.VehicleStatusIdle,
		},
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("TEST: Specific Nodes Route Assignment")
	fmt.Println(strings.Repeat("=", 80))

	err := simulationengine.AssignVehicleRouteWithNodes(vehicle, graph, startNode, endNode)
	if err != nil {
		t.Fatalf("Failed to assign route: %v", err)
	}

	fmt.Printf("Vehicle ID: %s\n", vehicle.ID)
	fmt.Printf("Assigned Start: %s\n", startNode)
	fmt.Printf("Assigned End: %s\n", endNode)
	fmt.Printf("Route Start: %s\n", vehicle.Route.StartNode)
	fmt.Printf("Route End: %s\n", vehicle.Route.EndNode)
	fmt.Printf("Current Node: %s\n", vehicle.Route.CurrentNode)
	fmt.Printf("Target Node: %s\n", vehicle.Route.TargetNode)
	fmt.Printf("Edges: %d\n", len(vehicle.Route.Edges))

	fmt.Println(strings.Repeat("=", 80) + "\n")

	if vehicle.Route.StartNode != startNode {
		t.Errorf("Expected start node %s, got %s", startNode, vehicle.Route.StartNode)
	}
	if vehicle.Route.EndNode != endNode {
		t.Errorf("Expected end node %s, got %s", endNode, vehicle.Route.EndNode)
	}
}

func TestMultipleVehicleAssignments(t *testing.T) {
	config := simulationengine.NewMapGenerator(1500, 1500, 789, simulationengine.AlgoDelaunay, 30, 0)
	graph := config.Generate()

	spawnConfig := &simulationengine.VehicleSpawnConfig{
		SpawnStrategy:  simulationengine.SpawnRandom,
		TargetStrategy: simulationengine.TargetRandom,
		AllowSameNode:  false,
	}

	numVehicles := 5
	vehicles := make([]*entities.Vehicle, numVehicles)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("TEST: Multiple Vehicle Route Assignments")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Graph: %d nodes, %d edges\n\n", len(graph.Nodes), len(graph.Edges))

	for i := 0; i < numVehicles; i++ {
		vehicles[i] = &entities.Vehicle{
			ID:   fmt.Sprintf("vehicle-%d", i+1),
			Type: entities.VehicleTypSedan,
			State: entities.VehicleState{
				Status: entities.VehicleStatusIdle,
			},
		}

		err := simulationengine.AssignVehicleRoute(vehicles[i], graph, spawnConfig)
		if err != nil {
			t.Fatalf("Failed to assign route to vehicle %d: %v", i+1, err)
		}

		fmt.Printf("Vehicle %d:\n", i+1)
		fmt.Printf("  Route: %s → %s\n", vehicles[i].Route.StartNode[:8], vehicles[i].Route.EndNode[:8])
		fmt.Printf("  Edges: %d\n", len(vehicles[i].Route.Edges))
		fmt.Printf("  Status: %s\n\n", vehicles[i].State.Status)
	}

	fmt.Println(strings.Repeat("=", 80) + "\n")

	for i, v := range vehicles {
		if v.Route == nil {
			t.Errorf("Vehicle %d has no route", i+1)
		}
		if v.State.Status != entities.VehicleStatusMoving && v.State.Status != entities.VehicleStatusIdle {
			t.Errorf("Vehicle %d has unexpected status: %s", i+1, v.State.Status)
		}
	}
}
