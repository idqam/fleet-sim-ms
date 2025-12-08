package simulationengine

import (
	"fmt"
	"math"
	"math/rand/v2"

	"github.com/google/uuid"
	"github.com/m/internal/simulation/entities"
)

func generateNodes(N int, heightBound int, widthBound int) map[string]*entities.MapNode {
	nodes := make(map[string]*entities.MapNode)
	for i := 0; i < N; i++ {
		x, y := uniformRandomDistributionSampler(heightBound, widthBound)
		id := uuid.New().String()
		nodes[id] = &entities.MapNode{
			ID:          id,
			Position:    entities.Vector2D{X: float64(x), Y: float64(y)},
			Type:        entities.NodeTypeIntersection,
			Connections: make(map[string]bool),
		}
	}
	return nodes
}

func collectIDs(nodes map[string]*entities.MapNode) []string {
	ids := make([]string, 0, len(nodes))
	for id := range nodes {
		ids = append(ids, id)
	}
	return ids
}

func distance(a, b entities.Vector2D) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func optimalRadius(N int, area float64) float64 {
	d := math.Log(float64(N))
	return math.Sqrt((d * area) / (math.Pi * float64(N)))
}

func uniformRandomDistributionSampler(heightBound int, widthBound int) (int, int) {
	x := rand.IntN(widthBound)
	y := rand.IntN(heightBound)
	return x, y
}

func addEdge(edges map[string]bool, a, b int) {
	if a > b {
		a, b = b, a
	}
	key := fmt.Sprintf("%d-%d", a, b)
	edges[key] = true
}

func findEdge(g *entities.MapGraph, from, to string) *entities.MapEdge {
	for _, edge := range g.Edges {
		if (edge.From == from && edge.To == to) || (edge.Bidirectional && edge.From == to && edge.To == from) {
			return edge
		}
	}
	return nil
}

func BuildEdgesFromConnections(mg *entities.MapGraph) error {
	if mg.Edges == nil {
		mg.Edges = make(map[string]*entities.MapEdge)
	}

	created := make(map[string]bool)

	for fromID, fromNode := range mg.Nodes {
		for toID := range fromNode.Connections {
			toNode, exists := mg.Nodes[toID]
			if !exists {
				continue
			}

			var edgeKey string
			if fromID < toID {
				edgeKey = fromID + "-" + toID
			} else {
				edgeKey = toID + "-" + fromID
			}

			if created[edgeKey] {
				continue
			}

			dist := distance(fromNode.Position, toNode.Position)

			var baseSpeedLimit float64
			if dist < 100 {
				baseSpeedLimit = 13.4
			} else if dist < 300 {
				baseSpeedLimit = 22.2
			} else {
				baseSpeedLimit = 33.3
			}

			edge := &entities.MapEdge{
				ID:             edgeKey,
				From:           fromID,
				To:             toID,
				Length:         dist,
				BaseSpeedLimit: baseSpeedLimit,
				SurfaceQuality: 0.95 + (0.05 * float64(len(edgeKey)%100) / 100.0),
				Bidirectional:  true,
				Conditions: &entities.RoadConditions{
					Congestion:          0.0,
					WeatherMultiplier:   1.0,
					EffectiveSpeedLimit: baseSpeedLimit,
				},
			}

			mg.Edges[edgeKey] = edge
			created[edgeKey] = true
		}
	}

	return nil
}

func EnsureGraphConnectivity(g *entities.MapGraph) {
	if len(g.Nodes) == 0 {
		return
	}

	components := findConnectedComponents(g)

	if len(components) <= 1 {
		return
	}

	for i := 0; i < len(components)-1; i++ {
		node1 := findClosestNodeInComponent(g, components[i], components[i+1])
		node2 := findClosestNodeInComponent(g, components[i+1], components[i])

		if node1 != nil && node2 != nil {
			connectNodes(g, node1, node2)
		}
	}
}

func findConnectedComponents(g *entities.MapGraph) [][]string {
	visited := make(map[string]bool)
	components := [][]string{}

	for nodeID := range g.Nodes {
		if !visited[nodeID] {
			component := []string{}
			queue := []string{nodeID}
			visited[nodeID] = true

			for len(queue) > 0 {
				current := queue[0]
				queue = queue[1:]
				component = append(component, current)

				for neighbor := range g.Nodes[current].Connections {
					if !visited[neighbor] {
						visited[neighbor] = true
						queue = append(queue, neighbor)
					}
				}
			}

			components = append(components, component)
		}
	}

	return components
}

func findClosestNodeInComponent(g *entities.MapGraph, component1, component2 []string) *entities.MapNode {
	var closest *entities.MapNode
	minDist := math.Inf(1)

	for _, id1 := range component1 {
		node1 := g.Nodes[id1]
		for _, id2 := range component2 {
			node2 := g.Nodes[id2]
			dist := distance(node1.Position, node2.Position)
			if dist < minDist {
				minDist = dist
				closest = node1
			}
		}
	}

	return closest
}

func connectNodes(g *entities.MapGraph, node1, node2 *entities.MapNode) {
	node1.Connections[node2.ID] = true
	node2.Connections[node1.ID] = true

	dist := distance(node1.Position, node2.Position)

	var baseSpeedLimit float64
	if dist < 100 {
		baseSpeedLimit = 13.4
	} else if dist < 300 {
		baseSpeedLimit = 22.2
	} else {
		baseSpeedLimit = 33.3
	}

	edgeKey := node1.ID + "-" + node2.ID
	if node1.ID > node2.ID {
		edgeKey = node2.ID + "-" + node1.ID
	}

	edge := &entities.MapEdge{
		ID:             edgeKey,
		From:           node1.ID,
		To:             node2.ID,
		Length:         dist,
		BaseSpeedLimit: baseSpeedLimit,
		SurfaceQuality: 0.90,
		Bidirectional:  true,
		Conditions: &entities.RoadConditions{
			Congestion:          0.0,
			WeatherMultiplier:   1.0,
			EffectiveSpeedLimit: baseSpeedLimit,
		},
	}

	g.Edges[edgeKey] = edge
}

func ApplyWeightVariation(g *entities.MapGraph, config *WeightVariationConfig, bounds MapBounds) {
	centerX := float64(bounds.Width) / 2.0
	centerY := float64(bounds.Height) / 2.0
	maxDistFromCenter := math.Sqrt(centerX*centerX + centerY*centerY)

	for _, edge := range g.Edges {
		curvature := config.CurvatureMin + rand.Float64()*(config.CurvatureMax-config.CurvatureMin)
		edge.Length *= curvature

		speedVariation := 1.0 + (rand.Float64()*2.0-1.0)*config.SpeedVariation
		edge.BaseSpeedLimit *= speedVariation
		if edge.Conditions != nil {
			edge.Conditions.EffectiveSpeedLimit = edge.BaseSpeedLimit
		}

		quality := rand.NormFloat64()*config.QualityStdDev + config.QualityMean
		quality = math.Max(0.5, math.Min(1.0, quality))

		if config.UseDistanceFromCenter {
			fromNode := g.Nodes[edge.From]
			toNode := g.Nodes[edge.To]

			if fromNode != nil && toNode != nil {
				midX := (fromNode.Position.X + toNode.Position.X) / 2.0
				midY := (fromNode.Position.Y + toNode.Position.Y) / 2.0

				distFromCenter := math.Sqrt((midX-centerX)*(midX-centerX) + (midY-centerY)*(midY-centerY))
				normalizedDist := distFromCenter / maxDistFromCenter

				centerBonus := (1.0 - normalizedDist) * 0.1
				quality += centerBonus
				quality = math.Max(0.5, math.Min(1.0, quality))
			}
		}

		edge.SurfaceQuality = quality
	}
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
