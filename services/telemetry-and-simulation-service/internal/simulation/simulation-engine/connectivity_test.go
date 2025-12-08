package simulationengine_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/m/internal/simulation/entities"
	simulationengine "github.com/m/internal/simulation/simulation-engine"
)

func TestGraphConnectivity_DisconnectedComponents(t *testing.T) {
	config := simulationengine.NewMapGenerator(1000, 1000, 42, simulationengine.AlgoRGG, 50, 0)
	config.RadiusMode = simulationengine.Sparse
	config.EnsureConnectivity = false

	graph := config.Generate()

	components := countComponents(graph)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("TEST: Graph Connectivity - Without Connectivity Enforcement")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Nodes: %d, Edges: %d, Components: %d\n", len(graph.Nodes), len(graph.Edges), components)

	if components > 1 {
		fmt.Printf("✓ Graph has %d disconnected components (as expected)\n", components)
	}

	fmt.Println(strings.Repeat("=", 80) + "\n")
}

func TestGraphConnectivity_EnsureConnected(t *testing.T) {
	config := simulationengine.NewMapGenerator(1000, 1000, 42, simulationengine.AlgoRGG, 50, 0)
	config.RadiusMode = simulationengine.Sparse
	config.EnsureConnectivity = true

	graph := config.Generate()

	components := countComponents(graph)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("TEST: Graph Connectivity - With Connectivity Enforcement")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Nodes: %d, Edges: %d, Components: %d\n", len(graph.Nodes), len(graph.Edges), components)

	if components == 1 {
		fmt.Println("✓ Graph is fully connected!")
	} else {
		t.Errorf("Expected 1 component, got %d", components)
	}

	fmt.Println(strings.Repeat("=", 80) + "\n")
}

func TestWeightVariation_DefaultConfig(t *testing.T) {
	config := simulationengine.NewMapGenerator(1000, 1000, 123, simulationengine.AlgoKNN, 30, 4)

	graph := config.Generate()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("TEST: Weight Variation - Default Configuration")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Nodes: %d, Edges: %d\n\n", len(graph.Nodes), len(graph.Edges))

	var minLength, maxLength float64 = 1e9, 0
	var minSpeed, maxSpeed float64 = 1e9, 0
	var minQuality, maxQuality float64 = 1.0, 0
	var totalQuality float64

	for _, edge := range graph.Edges {
		if edge.Length < minLength {
			minLength = edge.Length
		}
		if edge.Length > maxLength {
			maxLength = edge.Length
		}
		if edge.BaseSpeedLimit < minSpeed {
			minSpeed = edge.BaseSpeedLimit
		}
		if edge.BaseSpeedLimit > maxSpeed {
			maxSpeed = edge.BaseSpeedLimit
		}
		if edge.SurfaceQuality < minQuality {
			minQuality = edge.SurfaceQuality
		}
		if edge.SurfaceQuality > maxQuality {
			maxQuality = edge.SurfaceQuality
		}
		totalQuality += edge.SurfaceQuality
	}

	avgQuality := totalQuality / float64(len(graph.Edges))

	fmt.Printf("Edge Length Range:       %.2f - %.2f units\n", minLength, maxLength)
	fmt.Printf("Speed Limit Range:       %.2f - %.2f m/s\n", minSpeed, maxSpeed)
	fmt.Printf("Surface Quality Range:   %.3f - %.3f\n", minQuality, maxQuality)
	fmt.Printf("Average Surface Quality: %.3f\n\n", avgQuality)

	fmt.Println("Sample Edges:")
	count := 0
	for id, edge := range graph.Edges {
		if count >= 5 {
			break
		}
		fmt.Printf("  %s: Length=%.2f, Speed=%.2f, Quality=%.3f\n",
			id, edge.Length, edge.BaseSpeedLimit, edge.SurfaceQuality)
		count++
	}

	fmt.Println(strings.Repeat("=", 80) + "\n")

	if minQuality < 0.5 || maxQuality > 1.0 {
		t.Errorf("Surface quality out of bounds: %.3f - %.3f", minQuality, maxQuality)
	}
}

func TestWeightVariation_CustomConfig(t *testing.T) {
	config := simulationengine.NewMapGenerator(1000, 1000, 456, simulationengine.AlgoDelaunay, 25, 0)

	config.WeightVariation = &simulationengine.WeightVariationConfig{
		CurvatureMin:          1.1,
		CurvatureMax:          1.5,
		SpeedVariation:        0.2,
		QualityMean:           0.80,
		QualityStdDev:         0.08,
		UseDistanceFromCenter: true,
	}

	graph := config.Generate()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("TEST: Weight Variation - Custom Configuration")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Config: Curvature=[1.1-1.5], SpeedVar=0.2, Quality=0.80±0.08\n")
	fmt.Printf("Nodes: %d, Edges: %d\n\n", len(graph.Nodes), len(graph.Edges))

	var minLength, maxLength float64 = 1e9, 0
	var minSpeed, maxSpeed float64 = 1e9, 0
	var minQuality, maxQuality float64 = 1.0, 0
	var totalQuality float64

	for _, edge := range graph.Edges {
		if edge.Length < minLength {
			minLength = edge.Length
		}
		if edge.Length > maxLength {
			maxLength = edge.Length
		}
		if edge.BaseSpeedLimit < minSpeed {
			minSpeed = edge.BaseSpeedLimit
		}
		if edge.BaseSpeedLimit > maxSpeed {
			maxSpeed = edge.BaseSpeedLimit
		}
		if edge.SurfaceQuality < minQuality {
			minQuality = edge.SurfaceQuality
		}
		if edge.SurfaceQuality > maxQuality {
			maxQuality = edge.SurfaceQuality
		}
		totalQuality += edge.SurfaceQuality
	}

	avgQuality := totalQuality / float64(len(graph.Edges))

	fmt.Printf("Edge Length Range:       %.2f - %.2f units\n", minLength, maxLength)
	fmt.Printf("Speed Limit Range:       %.2f - %.2f m/s\n", minSpeed, maxSpeed)
	fmt.Printf("Surface Quality Range:   %.3f - %.3f\n", minQuality, maxQuality)
	fmt.Printf("Average Surface Quality: %.3f\n", avgQuality)

	fmt.Println(strings.Repeat("=", 80) + "\n")
}

func TestAllAlgorithms_WithConnectivityAndVariation(t *testing.T) {
	algorithms := []struct {
		name string
		algo simulationengine.Algorithm
	}{
		{"Random Geometric Graph", simulationengine.AlgoRGG},
		{"K-Nearest Neighbors", simulationengine.AlgoKNN},
		{"Delaunay Triangulation", simulationengine.AlgoDelaunay},
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("TEST: All Algorithms with Connectivity & Weight Variation")
	fmt.Println(strings.Repeat("=", 80) + "\n")

	for _, test := range algorithms {
		config := simulationengine.NewMapGenerator(800, 800, 789, test.algo, 40, 5)
		config.RadiusMode = simulationengine.Sparse

		graph := config.Generate()
		components := countComponents(graph)

		var totalQuality float64
		for _, edge := range graph.Edges {
			totalQuality += edge.SurfaceQuality
		}
		avgQuality := totalQuality / float64(len(graph.Edges))

		fmt.Printf("Algorithm: %s\n", test.name)
		fmt.Printf("  Nodes: %d, Edges: %d, Components: %d\n",
			len(graph.Nodes), len(graph.Edges), components)
		fmt.Printf("  Avg Surface Quality: %.3f\n", avgQuality)

		if components != 1 {
			t.Errorf("%s: Expected 1 component, got %d", test.name, components)
		}

		fmt.Println()
	}

	fmt.Println(strings.Repeat("=", 80) + "\n")
}

func countComponents(graph *entities.MapGraph) int {
	visited := make(map[string]bool)
	components := 0

	for nodeID := range graph.Nodes {
		if !visited[nodeID] {
			components++
			queue := []string{nodeID}
			visited[nodeID] = true

			for len(queue) > 0 {
				current := queue[0]
				queue = queue[1:]

				for neighbor := range graph.Nodes[current].Connections {
					if !visited[neighbor] {
						visited[neighbor] = true
						queue = append(queue, neighbor)
					}
				}
			}
		}
	}

	return components
}
