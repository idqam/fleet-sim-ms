package simulationengine

import "github.com/m/internal/simulation/entities"

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

func Dijkstra(g *entities.MapGraph, start, end string) {

}
