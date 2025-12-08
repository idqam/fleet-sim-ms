package simulationengine_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/m/internal/simulation/entities"
	simulationengine "github.com/m/internal/simulation/simulation-engine"
)

func TestDijkstraRoutes_PrintPrecomputed(t *testing.T) {
	graph := &entities.MapGraph{
		Nodes: make(map[string]*entities.MapNode),
		Edges: make(map[string]*entities.MapEdge),
	}

	nodes := map[string]entities.Vector2D{
		"A": {X: 0, Y: 0},
		"B": {X: 100, Y: 0},
		"C": {X: 200, Y: 0},
		"D": {X: 0, Y: 100},
		"E": {X: 100, Y: 100},
		"F": {X: 200, Y: 100},
		"G": {X: 0, Y: 200},
		"H": {X: 100, Y: 200},
		"I": {X: 200, Y: 200},
	}

	for id, pos := range nodes {
		graph.Nodes[id] = &entities.MapNode{
			ID:          id,
			Position:    pos,
			Type:        entities.NodeTypeIntersection,
			Connections: make(map[string]bool),
		}
	}

	connections := map[string][]string{
		"A": {"B", "D"},
		"B": {"A", "C", "E"},
		"C": {"B", "F"},
		"D": {"A", "E", "G"},
		"E": {"B", "D", "F", "H"},
		"F": {"C", "E", "I"},
		"G": {"D", "H"},
		"H": {"G", "E", "I"},
		"I": {"F", "H"},
	}

	for nodeID, neighbors := range connections {
		for _, neighbor := range neighbors {
			graph.Nodes[nodeID].Connections[neighbor] = true
		}
	}

	err := simulationengine.BuildEdgesFromConnections(graph)
	if err != nil {
		t.Fatalf("Failed to build edges: %v", err)
	}

	testRoutes := []struct {
		start string
		end   string
		desc  string
	}{
		{"A", "I", "Top-left to bottom-right (diagonal)"},
		{"A", "C", "Top-left to top-right (horizontal)"},
		{"A", "G", "Top-left to bottom-left (vertical)"},
		{"B", "H", "Top-middle to bottom-middle (vertical)"},
		{"C", "G", "Top-right to bottom-left (diagonal)"},
		{"E", "I", "Center to bottom-right"},
		{"A", "E", "Top-left to center"},
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("PRECOMPUTED DIJKSTRA ROUTES")
	fmt.Println(strings.Repeat("=", 80) + "\n")

	for i, test := range testRoutes {
		fmt.Printf("Route %d: %s\n", i+1, test.desc)
		fmt.Printf("From: %s → To: %s\n", test.start, test.end)
		fmt.Println(strings.Repeat("-", 80))

		routes := simulationengine.Dijkstra(graph, test.start, test.end)

		if len(routes) == 0 {
			fmt.Printf("❌ No route found from %s to %s\n\n", test.start, test.end)
			t.Errorf("Expected route from %s to %s, but got empty result", test.start, test.end)
			continue
		}

		route := routes[0]

		fmt.Printf("✓ Route found!\n")
		fmt.Printf("  Start Node:     %s\n", route.StartNode)
		fmt.Printf("  End Node:       %s\n", route.EndNode)
		fmt.Printf("  Total Distance: %.2f units\n", route.TotalDistance)
		fmt.Printf("  Number of Edges: %d\n", len(route.Edges))
		fmt.Printf("\n")

		fmt.Printf("  Path Visualization:\n")
		fmt.Printf("  %s", route.StartNode)

		for j, edgeID := range route.Edges {
			edge := graph.Edges[edgeID]
			if edge != nil {
				fmt.Printf(" --[%.1f]--> %s", edge.Length, edge.To)
			} else {
				fmt.Printf(" --[?]--> ?")
			}

			if (j+1)%3 == 0 && j < len(route.Edges)-1 {
				fmt.Printf("\n  ")
			}
		}
		fmt.Printf("\n\n")

		fmt.Printf("  Detailed Edge Information:\n")
		for j, edgeID := range route.Edges {
			edge := graph.Edges[edgeID]
			if edge != nil {
				fmt.Printf("    %d. Edge %s: %s → %s (%.2f units, speed limit: %.1f m/s)\n",
					j+1, edge.ID, edge.From, edge.To, edge.Length, edge.BaseSpeedLimit)
			}
		}

		fmt.Println("\n" + strings.Repeat("=", 80) + "\n")

		if route.StartNode != test.start {
			t.Errorf("Route %d: Expected start node %s, got %s", i+1, test.start, route.StartNode)
		}
		if route.EndNode != test.end {
			t.Errorf("Route %d: Expected end node %s, got %s", i+1, test.end, route.EndNode)
		}
		if route.TotalDistance <= 0 {
			t.Errorf("Route %d: Invalid total distance: %.2f", i+1, route.TotalDistance)
		}
		if len(route.Edges) == 0 {
			t.Errorf("Route %d: Route has no edges", i+1)
		}
	}

	fmt.Println("GRAPH STATISTICS")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Total Nodes: %d\n", len(graph.Nodes))
	fmt.Printf("Total Edges: %d\n", len(graph.Edges))
	fmt.Printf("\nNode Positions:\n")
	for id, node := range graph.Nodes {
		fmt.Printf("  %s: (%.0f, %.0f) - %d connections\n",
			id, node.Position.X, node.Position.Y, len(node.Connections))
	}
	fmt.Println(strings.Repeat("=", 80) + "\n")
}

func TestDijkstraRoutes_NoPath(t *testing.T) {
	graph := &entities.MapGraph{
		Nodes: make(map[string]*entities.MapNode),
		Edges: make(map[string]*entities.MapEdge),
	}

	graph.Nodes["A"] = &entities.MapNode{
		ID:          "A",
		Position:    entities.Vector2D{X: 0, Y: 0},
		Type:        entities.NodeTypeIntersection,
		Connections: make(map[string]bool),
	}

	graph.Nodes["B"] = &entities.MapNode{
		ID:          "B",
		Position:    entities.Vector2D{X: 100, Y: 0},
		Type:        entities.NodeTypeIntersection,
		Connections: make(map[string]bool),
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("TEST: Disconnected Graph (No Path Expected)")
	fmt.Println(strings.Repeat("=", 80))

	routes := simulationengine.Dijkstra(graph, "A", "B")

	if len(routes) != 0 {
		t.Errorf("Expected no route for disconnected nodes, but got %d routes", len(routes))
	} else {
		fmt.Println("✓ Correctly returned no route for disconnected nodes")
	}

	fmt.Println(strings.Repeat("=", 80) + "\n")
}

func TestDijkstraRoutes_SameNode(t *testing.T) {
	graph := &entities.MapGraph{
		Nodes: make(map[string]*entities.MapNode),
		Edges: make(map[string]*entities.MapEdge),
	}

	graph.Nodes["A"] = &entities.MapNode{
		ID:          "A",
		Position:    entities.Vector2D{X: 0, Y: 0},
		Type:        entities.NodeTypeIntersection,
		Connections: make(map[string]bool),
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("TEST: Same Start and End Node")
	fmt.Println(strings.Repeat("=", 80))

	routes := simulationengine.Dijkstra(graph, "A", "A")

	if len(routes) == 0 {
		fmt.Println("✓ No route returned for same start/end node")
	} else {
		route := routes[0]
		fmt.Printf("Route Details:\n")
		fmt.Printf("  Start: %s, End: %s\n", route.StartNode, route.EndNode)
		fmt.Printf("  Distance: %.2f\n", route.TotalDistance)
		fmt.Printf("  Edges: %d\n", len(route.Edges))

		if route.TotalDistance != 0 {
			t.Errorf("Expected distance 0 for same node route, got %.2f", route.TotalDistance)
		}
	}

	fmt.Println(strings.Repeat("=", 80) + "\n")
}
