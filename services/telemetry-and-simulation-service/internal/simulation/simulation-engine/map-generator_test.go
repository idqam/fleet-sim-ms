package simulationengine_test

import (
	"testing"

	"github.com/m/internal/simulation/entities"
	simulationengine "github.com/m/internal/simulation/simulation-engine"
)

func TestBuildEdgesFromConnections(t *testing.T) {
	graph := &entities.MapGraph{
		Nodes: make(map[string]*entities.MapNode),
		Edges: make(map[string]*entities.MapEdge),
	}

	nodeA := &entities.MapNode{
		ID:          "node-a",
		Position:    entities.Vector2D{X: 0, Y: 0},
		Type:        entities.NodeTypeIntersection,
		Connections: make(map[string]bool),
	}
	nodeB := &entities.MapNode{
		ID:          "node-b",
		Position:    entities.Vector2D{X: 100, Y: 0},
		Type:        entities.NodeTypeIntersection,
		Connections: make(map[string]bool),
	}
	nodeC := &entities.MapNode{
		ID:          "node-c",
		Position:    entities.Vector2D{X: 50, Y: 86.6},
		Type:        entities.NodeTypeIntersection,
		Connections: make(map[string]bool),
	}

	nodeA.Connections["node-b"] = true
	nodeB.Connections["node-a"] = true
	nodeB.Connections["node-c"] = true
	nodeC.Connections["node-b"] = true
	nodeC.Connections["node-a"] = true
	nodeA.Connections["node-c"] = true

	graph.Nodes["node-a"] = nodeA
	graph.Nodes["node-b"] = nodeB
	graph.Nodes["node-c"] = nodeC

	err := simulationengine.BuildEdgesFromConnections(graph)
	if err != nil {
		t.Fatalf("BuildEdgesFromConnections failed: %v", err)
	}

	if len(graph.Edges) != 3 {
		t.Errorf("Expected 3 edges, got %d", len(graph.Edges))
	}

	for edgeID, edge := range graph.Edges {
		t.Logf("Edge %s: From=%s, To=%s, Length=%.2f, BaseSpeedLimit=%.2f, SurfaceQuality=%.2f",
			edgeID, edge.From, edge.To, edge.Length, edge.BaseSpeedLimit, edge.SurfaceQuality)

		if edge.Length <= 0 {
			t.Errorf("Edge %s has invalid length: %f", edgeID, edge.Length)
		}

		if edge.BaseSpeedLimit <= 0 {
			t.Errorf("Edge %s has invalid base speed limit: %f", edgeID, edge.BaseSpeedLimit)
		}

		if edge.SurfaceQuality < 0.95 || edge.SurfaceQuality > 1.0 {
			t.Errorf("Edge %s has invalid surface quality: %f (expected 0.95-1.0)", edgeID, edge.SurfaceQuality)
		}

		if !edge.Bidirectional {
			t.Errorf("Edge %s should be bidirectional", edgeID)
		}

		if edge.Conditions == nil {
			t.Errorf("Edge %s has nil conditions", edgeID)
			continue
		}

		if edge.Conditions.WeatherMultiplier != 1.0 {
			t.Errorf("Edge %s has invalid weather multiplier: %f (expected 1.0)", edgeID, edge.Conditions.WeatherMultiplier)
		}

		if edge.Conditions.EffectiveSpeedLimit != edge.BaseSpeedLimit {
			t.Errorf("Edge %s has mismatched effective speed limit: %f (expected %f)",
				edgeID, edge.Conditions.EffectiveSpeedLimit, edge.BaseSpeedLimit)
		}

		if edge.Conditions.Congestion != 0.0 {
			t.Errorf("Edge %s has non-zero initial congestion: %f", edgeID, edge.Conditions.Congestion)
		}
	}
}

func TestBuildEdgesFromConnections_EmptyGraph(t *testing.T) {
	graph := &entities.MapGraph{
		Nodes: make(map[string]*entities.MapNode),
	}

	err := simulationengine.BuildEdgesFromConnections(graph)
	if err != nil {
		t.Fatalf("BuildEdgesFromConnections failed on empty graph: %v", err)
	}

	if len(graph.Edges) != 0 {
		t.Errorf("Expected 0 edges for empty graph, got %d", len(graph.Edges))
	}
}

func TestBuildEdgesFromConnections_InvalidConnection(t *testing.T) {
	graph := &entities.MapGraph{
		Nodes: make(map[string]*entities.MapNode),
		Edges: make(map[string]*entities.MapEdge),
	}

	nodeA := &entities.MapNode{
		ID:          "node-a",
		Position:    entities.Vector2D{X: 0, Y: 0},
		Type:        entities.NodeTypeIntersection,
		Connections: make(map[string]bool),
	}
	nodeA.Connections["non-existent-node"] = true

	graph.Nodes["node-a"] = nodeA

	err := simulationengine.BuildEdgesFromConnections(graph)
	if err != nil {
		t.Fatalf("BuildEdgesFromConnections failed: %v", err)
	}

	if len(graph.Edges) != 0 {
		t.Errorf("Expected 0 edges (invalid connection should be skipped), got %d", len(graph.Edges))
	}
}

func TestBuildEdgesFromConnections_SpeedLimitCategories(t *testing.T) {
	graph := &entities.MapGraph{
		Nodes: make(map[string]*entities.MapNode),
		Edges: make(map[string]*entities.MapEdge),
	}

	tests := []struct {
		name          string
		distance      float64
		expectedSpeed float64
		category      string
	}{
		{"urban", 50, 13.4, "urban"},
		{"suburban", 200, 22.2, "suburban"},
		{"highway", 400, 33.3, "highway"},
	}

	for i, tt := range tests {
		nodeA := &entities.MapNode{
			ID:          "node-" + tt.name + "-a",
			Position:    entities.Vector2D{X: 0, Y: 0},
			Type:        entities.NodeTypeIntersection,
			Connections: make(map[string]bool),
		}
		nodeB := &entities.MapNode{
			ID:          "node-" + tt.name + "-b",
			Position:    entities.Vector2D{X: tt.distance, Y: 0},
			Type:        entities.NodeTypeIntersection,
			Connections: make(map[string]bool),
		}

		nodeA.Connections[nodeB.ID] = true
		nodeB.Connections[nodeA.ID] = true

		graph.Nodes[nodeA.ID] = nodeA
		graph.Nodes[nodeB.ID] = nodeB

		t.Logf("Test %d: %s - distance=%.2f, expected speed=%.2f", i, tt.category, tt.distance, tt.expectedSpeed)
	}

	err := simulationengine.BuildEdgesFromConnections(graph)
	if err != nil {
		t.Fatalf("BuildEdgesFromConnections failed: %v", err)
	}

	if len(graph.Edges) != 3 {
		t.Errorf("Expected 3 edges, got %d", len(graph.Edges))
	}

	for edgeID, edge := range graph.Edges {
		t.Logf("Edge %s: Length=%.2f, BaseSpeedLimit=%.2f", edgeID, edge.Length, edge.BaseSpeedLimit)

		var expectedSpeed float64
		if edge.Length < 100 {
			expectedSpeed = 13.4
		} else if edge.Length < 300 {
			expectedSpeed = 22.2
		} else {
			expectedSpeed = 33.3
		}

		if edge.BaseSpeedLimit != expectedSpeed {
			t.Errorf("Edge %s has incorrect speed limit: got %.2f, expected %.2f (length=%.2f)",
				edgeID, edge.BaseSpeedLimit, expectedSpeed, edge.Length)
		}
	}
}
