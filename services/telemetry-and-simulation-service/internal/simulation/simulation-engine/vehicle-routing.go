package simulationengine

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/m/internal/simulation/entities"
)

type VehicleSpawnConfig struct {
	SpawnStrategy  SpawnStrategy
	TargetStrategy TargetStrategy
	AllowSameNode  bool
}

type SpawnStrategy string
type TargetStrategy string

const (
	SpawnRandom      SpawnStrategy = "random"
	SpawnSpecific    SpawnStrategy = "specific"
	SpawnDistributed SpawnStrategy = "distributed"
)

const (
	TargetRandom   TargetStrategy = "random"
	TargetSpecific TargetStrategy = "specific"
	TargetFarthest TargetStrategy = "farthest"
)

func AssignVehicleRoute(vehicle *entities.Vehicle, graph *entities.MapGraph, config *VehicleSpawnConfig) error {
	if len(graph.Nodes) == 0 {
		return fmt.Errorf("graph has no nodes")
	}

	nodeIDs := make([]string, 0, len(graph.Nodes))
	for id := range graph.Nodes {
		nodeIDs = append(nodeIDs, id)
	}

	spawnNode := selectSpawnNode(nodeIDs, config.SpawnStrategy)
	targetNode := selectTargetNode(nodeIDs, graph, spawnNode, config.TargetStrategy, config.AllowSameNode)

	if spawnNode == "" || targetNode == "" {
		return fmt.Errorf("failed to select spawn or target node")
	}

	routes := Dijkstra(graph, spawnNode, targetNode)
	if len(routes) == 0 {
		return fmt.Errorf("no route found from %s to %s", spawnNode, targetNode)
	}

	route := routes[0]

	if len(route.Edges) == 0 {
		if !config.AllowSameNode {
			return fmt.Errorf("route has no edges (same node)")
		}
	}

	now := time.Now()
	vehicle.Route = &entities.AssignedRoute{
		Edges:            route.Edges,
		CurrentEdgeIndex: 0,
		CurrentNode:      spawnNode,
		TargetNode:       getFirstTargetNode(route, graph),
		StartNode:        spawnNode,
		EndNode:          targetNode,
		StartedAt:        now,
		CompletedAt:      nil,
	}

	spawnPos := graph.Nodes[spawnNode].Position
	vehicle.State.CurrentPosition = spawnPos
	vehicle.State.Velocity = entities.Vector2D{X: 0, Y: 0}
	vehicle.State.ProgressOnEdge = 0.0
	vehicle.State.LastUpdateTime = now

	if len(route.Edges) > 0 {
		vehicle.State.CurrentEdge = route.Edges[0]
		vehicle.State.Status = entities.VehicleStatusMoving
	} else {
		vehicle.State.CurrentEdge = ""
		vehicle.State.Status = entities.VehicleStatusIdle
	}

	return nil
}

func AssignVehicleRouteWithNodes(vehicle *entities.Vehicle, graph *entities.MapGraph, startNode, endNode string) error {
	if _, exists := graph.Nodes[startNode]; !exists {
		return fmt.Errorf("start node %s not found in graph", startNode)
	}
	if _, exists := graph.Nodes[endNode]; !exists {
		return fmt.Errorf("end node %s not found in graph", endNode)
	}

	routes := Dijkstra(graph, startNode, endNode)
	if len(routes) == 0 {
		return fmt.Errorf("no route found from %s to %s", startNode, endNode)
	}

	route := routes[0]
	now := time.Now()

	vehicle.Route = &entities.AssignedRoute{
		Edges:            route.Edges,
		CurrentEdgeIndex: 0,
		CurrentNode:      startNode,
		TargetNode:       getFirstTargetNode(route, graph),
		StartNode:        startNode,
		EndNode:          endNode,
		StartedAt:        now,
		CompletedAt:      nil,
	}

	spawnPos := graph.Nodes[startNode].Position
	vehicle.State.CurrentPosition = spawnPos
	vehicle.State.Velocity = entities.Vector2D{X: 0, Y: 0}
	vehicle.State.ProgressOnEdge = 0.0
	vehicle.State.LastUpdateTime = now

	if len(route.Edges) > 0 {
		vehicle.State.CurrentEdge = route.Edges[0]
		vehicle.State.Status = entities.VehicleStatusMoving
	} else {
		vehicle.State.CurrentEdge = ""
		vehicle.State.Status = entities.VehicleStatusIdle
	}

	return nil
}

func selectSpawnNode(nodeIDs []string, strategy SpawnStrategy) string {
	if len(nodeIDs) == 0 {
		return ""
	}

	switch strategy {
	case SpawnRandom:
		return nodeIDs[rand.IntN(len(nodeIDs))]
	case SpawnDistributed:
		return selectDistributedNode(nodeIDs)
	default:
		return nodeIDs[rand.IntN(len(nodeIDs))]
	}
}

func selectTargetNode(nodeIDs []string, graph *entities.MapGraph, spawnNode string, strategy TargetStrategy, allowSame bool) string {
	if len(nodeIDs) == 0 {
		return ""
	}

	switch strategy {
	case TargetRandom:
		for i := 0; i < 10; i++ {
			target := nodeIDs[rand.IntN(len(nodeIDs))]
			if allowSame || target != spawnNode {
				return target
			}
		}
		return nodeIDs[rand.IntN(len(nodeIDs))]

	case TargetFarthest:
		return selectFarthestNode(nodeIDs, graph, spawnNode, allowSame)

	default:
		for i := 0; i < 10; i++ {
			target := nodeIDs[rand.IntN(len(nodeIDs))]
			if allowSame || target != spawnNode {
				return target
			}
		}
		return nodeIDs[rand.IntN(len(nodeIDs))]
	}
}

func selectDistributedNode(nodeIDs []string) string {
	if len(nodeIDs) == 0 {
		return ""
	}

	candidates := make([]string, 0, 5)
	for i := 0; i < 5 && i < len(nodeIDs); i++ {
		idx := rand.IntN(len(nodeIDs))
		candidates = append(candidates, nodeIDs[idx])
	}

	if len(candidates) == 0 {
		return nodeIDs[0]
	}

	return candidates[rand.IntN(len(candidates))]
}

func selectFarthestNode(nodeIDs []string, graph *entities.MapGraph, fromNode string, allowSame bool) string {
	if len(nodeIDs) == 0 {
		return ""
	}

	fromPos := graph.Nodes[fromNode].Position
	maxDist := 0.0
	farthest := nodeIDs[0]

	for _, nodeID := range nodeIDs {
		if !allowSame && nodeID == fromNode {
			continue
		}

		nodePos := graph.Nodes[nodeID].Position
		dist := distance(fromPos, nodePos)

		if dist > maxDist {
			maxDist = dist
			farthest = nodeID
		}
	}

	return farthest
}

func getFirstTargetNode(route *entities.Route, graph *entities.MapGraph) string {
	if len(route.Edges) == 0 {
		return route.EndNode
	}

	firstEdge := graph.Edges[route.Edges[0]]
	if firstEdge != nil {
		return firstEdge.To
	}

	return route.EndNode
}


