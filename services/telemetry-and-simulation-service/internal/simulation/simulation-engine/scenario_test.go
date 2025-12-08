package simulationengine_test

import (
	"fmt"
	"strings"
	"testing"

	simulationengine "github.com/m/internal/simulation/simulation-engine"
)

func TestRealWorldScenario_CitySimulation(t *testing.T) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("REAL-WORLD SCENARIO: City Fleet Simulation Map")
	fmt.Println(strings.Repeat("=", 80) + "\n")

	config := simulationengine.NewMapGenerator(2000, 2000, 12345, simulationengine.AlgoDelaunay, 100, 0)

	config.WeightVariation = &simulationengine.WeightVariationConfig{
		CurvatureMin:          1.05,
		CurvatureMax:          1.25,
		SpeedVariation:        0.15,
		QualityMean:           0.85,
		QualityStdDev:         0.10,
		UseDistanceFromCenter: true,
	}

	config.EnsureConnectivity = true

	fmt.Println("Configuration:")
	fmt.Println("  Map Size: 2000x2000 units")
	fmt.Println("  Nodes: 100 (intersections)")
	fmt.Println("  Algorithm: Delaunay Triangulation")
	fmt.Println("  Curvature: 1.05-1.25 (realistic road curves)")
	fmt.Println("  Speed Variation: ±15%")
	fmt.Println("  Quality: 0.85 ± 0.10 (better in city center)")
	fmt.Println("  Connectivity: Enforced")

	graph := config.Generate()

	components := countComponents(graph)

	fmt.Printf("Generated Graph Statistics:\n")
	fmt.Printf("  Total Nodes: %d\n", len(graph.Nodes))
	fmt.Printf("  Total Edges: %d\n", len(graph.Edges))
	fmt.Printf("  Connected Components: %d\n", components)
	fmt.Printf("  Avg Edges per Node: %.2f\n\n", float64(len(graph.Edges)*2)/float64(len(graph.Nodes)))

	var minLength, maxLength, totalLength float64 = 1e9, 0, 0
	var minSpeed, maxSpeed, totalSpeed float64 = 1e9, 0, 0
	var minQuality, maxQuality, totalQuality float64 = 1.0, 0, 0

	urbanCount := 0
	suburbanCount := 0
	highwayCount := 0

	for _, edge := range graph.Edges {
		if edge.Length < minLength {
			minLength = edge.Length
		}
		if edge.Length > maxLength {
			maxLength = edge.Length
		}
		totalLength += edge.Length

		if edge.BaseSpeedLimit < minSpeed {
			minSpeed = edge.BaseSpeedLimit
		}
		if edge.BaseSpeedLimit > maxSpeed {
			maxSpeed = edge.BaseSpeedLimit
		}
		totalSpeed += edge.BaseSpeedLimit

		if edge.SurfaceQuality < minQuality {
			minQuality = edge.SurfaceQuality
		}
		if edge.SurfaceQuality > maxQuality {
			maxQuality = edge.SurfaceQuality
		}
		totalQuality += edge.SurfaceQuality

		if edge.BaseSpeedLimit < 18 {
			urbanCount++
		} else if edge.BaseSpeedLimit < 28 {
			suburbanCount++
		} else {
			highwayCount++
		}
	}

	avgLength := totalLength / float64(len(graph.Edges))
	avgSpeed := totalSpeed / float64(len(graph.Edges))
	avgQuality := totalQuality / float64(len(graph.Edges))

	fmt.Println("Edge Statistics:")
	fmt.Printf("  Length:  min=%.2f, max=%.2f, avg=%.2f units\n", minLength, maxLength, avgLength)
	fmt.Printf("  Speed:   min=%.2f, max=%.2f, avg=%.2f m/s\n", minSpeed, maxSpeed, avgSpeed)
	fmt.Printf("  Quality: min=%.3f, max=%.3f, avg=%.3f\n\n", minQuality, maxQuality, avgQuality)

	fmt.Println("Road Type Distribution:")
	fmt.Printf("  Urban (< 18 m/s):      %d edges (%.1f%%)\n", urbanCount, 100.0*float64(urbanCount)/float64(len(graph.Edges)))
	fmt.Printf("  Suburban (18-28 m/s):  %d edges (%.1f%%)\n", suburbanCount, 100.0*float64(suburbanCount)/float64(len(graph.Edges)))
	fmt.Printf("  Highway (> 28 m/s):    %d edges (%.1f%%)\n\n", highwayCount, 100.0*float64(highwayCount)/float64(len(graph.Edges)))

	fmt.Println("Sample Routes for Fleet Vehicles:")

	nodeIDs := make([]string, 0, len(graph.Nodes))
	for id := range graph.Nodes {
		nodeIDs = append(nodeIDs, id)
		if len(nodeIDs) >= 6 {
			break
		}
	}

	if len(nodeIDs) >= 6 {
		routes := []struct {
			start int
			end   int
			desc  string
		}{
			{0, 3, "Short delivery"},
			{1, 5, "Cross-city route"},
			{2, 4, "Medium distance"},
		}

		for i, route := range routes {
			startID := nodeIDs[route.start]
			endID := nodeIDs[route.end]

			result := simulationengine.Dijkstra(graph, startID, endID)

			if len(result) > 0 {
				r := result[0]
				fmt.Printf("  Route %d (%s):\n", i+1, route.desc)
				fmt.Printf("    Distance: %.2f units\n", r.TotalDistance)
				fmt.Printf("    Segments: %d edges\n", len(r.Edges))

				var totalTime float64
				for _, edgeID := range r.Edges {
					edge := graph.Edges[edgeID]
					if edge != nil && edge.BaseSpeedLimit > 0 {
						totalTime += edge.Length / edge.BaseSpeedLimit
					}
				}
				fmt.Printf("    Est. Time: %.2f seconds\n\n", totalTime)
			}
		}
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("✓ City simulation map ready for fleet operations")
	fmt.Println(strings.Repeat("=", 80) + "\n")

	if components != 1 {
		t.Errorf("Expected fully connected graph, got %d components", components)
	}

	if avgQuality < 0.75 || avgQuality > 0.95 {
		t.Errorf("Average quality out of expected range: %.3f", avgQuality)
	}

	if len(graph.Edges) < len(graph.Nodes)-1 {
		t.Errorf("Not enough edges for connectivity: %d edges for %d nodes", len(graph.Edges), len(graph.Nodes))
	}
}
