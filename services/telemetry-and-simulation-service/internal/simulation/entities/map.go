package entities

import (
	"time"
)

type MapGraph struct {
	Nodes map[string]*MapNode `json:"nodes"`
	Edges map[string]*MapEdge `json:"edges"`
}

type MapNode struct {
	ID          string   `json:"id"`
	Position    Vector2D `json:"position"`
	Type        NodeType `json:"type"`
	Connections []string `json:"connections"`
}

type MapEdge struct {
	ID                string  `json:"id"`
	From              string  `json:"from"`
	To                string  `json:"to"`
	Length            float64 `json:"length"`
	BaseSpeedLimit    float64 `json:"base_speed_limit"`
	WeatherMultiplier float64 `json:"weather_multiplier"`
	SurfaceQuality    float64 `json:"surface_quality"`
	Bidirectional     bool    `json:"bidirectional"`
}

type RoadConditions struct {
	Congestion          float64   `json:"congestion"`
	EffectiveSpeedLimit float64   `json:"effective_speed_limit"`
	LastUpdated         time.Time `json:"last_updated"`
}

type Route struct {
	Edges         []string `json:"edges"`
	StartNode     string   `json:"start_node"`
	EndNode       string   `json:"end_node"`
	TotalDistance float64  `json:"total_distance"`
}
