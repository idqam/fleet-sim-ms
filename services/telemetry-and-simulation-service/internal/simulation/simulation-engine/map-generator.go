package simulationengine

import (
	"fmt"
	"sort"

	"github.com/fogleman/delaunay"
	"github.com/m/internal/simulation/entities"
)

type MapBounds struct {
	Height int
	Width  int
}

type RadiusMode int

const (
	Sparse RadiusMode = iota
	Connected
)

type Algorithm string

const (
	AlgoRGG      Algorithm = "rgg"
	AlgoKNN      Algorithm = "knn"
	AlgoDelaunay Algorithm = "delaunay"
)

type MapGeneratorConfig struct {
	Bounds     MapBounds
	Seed       int64
	Algorithm  Algorithm
	N          int
	K          int
	RadiusMode RadiusMode
}

func NewMapGenerator(height int, width int, seed int64, algorithm Algorithm, n, k int) *MapGeneratorConfig {
	return &MapGeneratorConfig{
		Bounds: MapBounds{
			Height: height,
			Width:  width,
		},
		Seed:      seed,
		Algorithm: algorithm,
		N:         n,
		K:         k,
	}
}

func (cfg *MapGeneratorConfig) Generate() *entities.MapGraph {
	switch cfg.Algorithm {
	case AlgoRGG:
		return RandomGeometricGraph(cfg.N, cfg.Bounds.Height, cfg.Bounds.Width, cfg.RadiusMode)
	case AlgoKNN:
		return KNNGraph(cfg.N, cfg.Bounds.Height, cfg.Bounds.Width, cfg.K)
	case AlgoDelaunay:
		return DelaunayGraph(cfg.N, cfg.Bounds.Height, cfg.Bounds.Width)
	default:
		return &entities.MapGraph{Nodes: map[string]*entities.MapNode{}}
	}
}

func RandomGeometricGraph(N int, heightBound int, widthBound int, mode RadiusMode) *entities.MapGraph {
	nodes := generateNodes(N, heightBound, widthBound)
	area := float64(heightBound) * float64(widthBound)
	r := OptimalRadius(N, area)
	if mode == Sparse {
		r *= 0.6
	}
	if mode == Connected {
		r *= 1.4
	}

	ids := collectIDs(nodes)
	edges := make(map[string]*entities.MapEdge)

	for i := 0; i < len(ids); i++ {
		for j := i + 1; j < len(ids); j++ {
			a := nodes[ids[i]]
			b := nodes[ids[j]]
			dist := distance(a.Position, b.Position)
			if dist <= r {
				a.Connections[b.ID] = true
				b.Connections[a.ID] = true
				edge := createEdge(a.ID, b.ID, dist)
				edges[edge.ID] = edge
			}
		}
	}

	return &entities.MapGraph{Nodes: nodes, Edges: edges}
}

func KNNGraph(N int, heightBound int, widthBound int, K int) *entities.MapGraph {
	nodes := generateNodes(N, heightBound, widthBound)
	ids := collectIDs(nodes)
	edges := make(map[string]*entities.MapEdge)

	for _, id := range ids {
		cur := nodes[id]

		type pair struct {
			id   string
			dist float64
		}

		distances := make([]pair, 0, N-1)

		for _, other := range ids {
			if other == id {
				continue
			}
			distances = append(distances, pair{
				id:   other,
				dist: distance(cur.Position, nodes[other].Position),
			})
		}

		sort.Slice(distances, func(i, j int) bool {
			return distances[i].dist < distances[j].dist
		})

		limit := K
		if limit > len(distances) {
			limit = len(distances)
		}

		for i := 0; i < limit; i++ {
			other := nodes[distances[i].id]
			cur.Connections[other.ID] = true
			other.Connections[cur.ID] = true
			edgeDist := distances[i].dist
			edge := createEdge(cur.ID, other.ID, edgeDist)
			edges[edge.ID] = edge
		}
	}

	return &entities.MapGraph{Nodes: nodes, Edges: edges}
}

func DelaunayGraph(N int, heightBound int, widthBound int) *entities.MapGraph {
	nodes := generateNodes(N, heightBound, widthBound)
	points := make([]delaunay.Point, 0, N)
	indexMap := make(map[int]string)

	i := 0
	for id, n := range nodes {
		points = append(points, delaunay.Point{X: n.Position.X, Y: n.Position.Y})
		indexMap[i] = id
		i++
	}

	tri, _ := delaunay.Triangulate(points)
	edges := make(map[string]*entities.MapEdge)
	edgeSet := make(map[string]bool)

	for i := 0; i < len(tri.Triangles); i += 3 {
		a := tri.Triangles[i]
		b := tri.Triangles[i+1]
		c := tri.Triangles[i+2]
		addEdge(edgeSet, a, b)
		addEdge(edgeSet, b, c)
		addEdge(edgeSet, c, a)
	}

	for edge := range edgeSet {
		var a, b int
		fmt.Sscanf(edge, "%d-%d", &a, &b)
		na := nodes[indexMap[a]]
		nb := nodes[indexMap[b]]
		na.Connections[nb.ID] = true
		nb.Connections[na.ID] = true
		edgeDist := distance(na.Position, nb.Position)
		mapEdge := createEdge(na.ID, nb.ID, edgeDist)
		edges[mapEdge.ID] = mapEdge
	}

	return &entities.MapGraph{Nodes: nodes, Edges: edges}
}
