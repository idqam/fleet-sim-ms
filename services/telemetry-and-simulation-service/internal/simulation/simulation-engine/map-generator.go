package simulationengine

import (
	"fmt"
	"math"
	"math/rand/v2"
	"sort"

	"github.com/fogleman/delaunay"
	"github.com/google/uuid"
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
	AlgoRGG       Algorithm = "rgg"
	AlgoKNN       Algorithm = "knn"
	AlgoDelaunay  Algorithm = "delaunay"
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

	for i := 0; i < len(ids); i++ {
		for j := i + 1; j < len(ids); j++ {
			a := nodes[ids[i]]
			b := nodes[ids[j]]
			if distance(a.Position, b.Position) <= r {
				a.Connections = append(a.Connections, b.ID)
				b.Connections = append(b.Connections, a.ID)
			}
		}
	}

	return &entities.MapGraph{Nodes: nodes}
}

func KNNGraph(N int, heightBound int, widthBound int, K int) *entities.MapGraph {
	nodes := generateNodes(N, heightBound, widthBound)
	ids := collectIDs(nodes)

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
			cur.Connections = append(cur.Connections, other.ID)
			other.Connections = append(other.Connections, cur.ID)
		}
	}

	return &entities.MapGraph{Nodes: nodes}
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

	edges := make(map[string]bool)
	for i := 0; i < len(tri.Triangles); i += 3 {
		
		a := tri.Triangles[i]
		b := tri.Triangles[i+1]
		c := tri.Triangles[i+2]
		addEdge(edges, a, b)
		addEdge(edges, b, c)
		addEdge(edges, c, a)
	}

	for edge := range edges {
		
		var a, b int
		fmt.Sscanf(edge, "%d-%d", &a, &b)

		na := nodes[indexMap[a]]
		nb := nodes[indexMap[b]]

		na.Connections = append(na.Connections, nb.ID)
		nb.Connections = append(nb.Connections, na.ID)
	}

	return &entities.MapGraph{Nodes: nodes}
}

func generateNodes(N int, heightBound int, widthBound int) map[string]*entities.MapNode {
	nodes := make(map[string]*entities.MapNode)
	for i := 0; i < N; i++ {
		x, y := UniformRandomDistributionSampler(heightBound, widthBound)
		id := uuid.New().String()
		nodes[id] = &entities.MapNode{
			ID:          id,
			Position:    entities.Vector2D{X: float64(x), Y: float64(y)},
			Type:        entities.NodeTypeIntersection,
			Connections: []string{},
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

func OptimalRadius(N int, area float64) float64 {
	d := math.Log(float64(N))
	return math.Sqrt((d * area) / (math.Pi * float64(N)))
}

func distance(a, b entities.Vector2D) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Sqrt(dx*dx + dy*dy)
}
func UniformRandomDistributionSampler(heightBound int, widthBound int) (int, int) {
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

