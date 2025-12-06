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
