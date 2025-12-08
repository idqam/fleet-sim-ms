package simulationengine

import (
	"math"

	"github.com/m/internal/simulation/entities"
)

type Algorithm string

const (
	AlgoDijkstra Algorithm = "djk"
	AlgoBiA      Algorithm = "bia"
	AlgoDFS      Algorithm = "dfs"
	AlgoBFS      Algorithm = "bfs"
)

type RoutingConfig struct {
	Algo Algorithm
}

func (c *RoutingConfig) ToString() string {
	return string(c.Algo)
}

func (c *RoutingConfig) ToAlgorithm() Algorithm {
	return c.Algo
}

func NewRoutingConfig(algo Algorithm) *RoutingConfig {
	return &RoutingConfig{Algo: algo}
}

func Dijkstra(g *entities.MapGraph, start, end string) []*entities.Route {
	dist := make(map[string]float64)
	prev := make(map[string]string)
	unvisited := make(map[string]bool)

	for id := range g.Nodes {
		dist[id] = math.Inf(1)
		unvisited[id] = true
	}

	dist[start] = 0

	for len(unvisited) > 0 {
		current := ""
		minDist := math.Inf(1)

		for n := range unvisited {
			if dist[n] < minDist {
				minDist = dist[n]
				current = n
			}
		}

		if current == "" {
			break
		}

		if current == end {
			break
		}

		delete(unvisited, current)

		node := g.Nodes[current]
		for neighbor := range node.Connections {
			edge := findEdge(g, current, neighbor)
			if edge == nil {
				continue
			}

			alt := dist[current] + edge.Length
			if alt < dist[neighbor] {
				dist[neighbor] = alt
				prev[neighbor] = current
			}
		}
	}

	pathNodes := []string{}
	u := end
	for {
		pathNodes = append([]string{u}, pathNodes...)
		if u == start {
			break
		}
		p, ok := prev[u]
		if !ok {
			return []*entities.Route{}
		}
		u = p
	}

	pathEdges := []string{}
	total := 0.0
	for i := 0; i < len(pathNodes)-1; i++ {
		e := findEdge(g, pathNodes[i], pathNodes[i+1])
		if e != nil {
			pathEdges = append(pathEdges, e.ID)
			total += e.Length
		}
	}

	r := entities.Route{
		Edges:         pathEdges,
		StartNode:     start,
		EndNode:       end,
		TotalDistance: total,
	}

	return []*entities.Route{&r}
}
