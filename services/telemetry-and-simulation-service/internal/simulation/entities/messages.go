package entities

import (
	"time"
)

type MapQuery interface {
	IsQuery()
}

type QueryNearbyVehicles struct {
	VehicleID    string                      `json:"vehicle_id"`
	Position     Vector2D                    `json:"position"`
	Radius       float64                     `json:"radius"`
	ResponseChan chan NearbyVehiclesResponse `json:"-"`
}

func (q QueryNearbyVehicles) IsQuery() {}

type QueryEdgeConditions struct {
	EdgeID       string                      `json:"edge_id"`
	ResponseChan chan EdgeConditionsResponse `json:"-"`
}

func (q QueryEdgeConditions) IsQuery() {}

type QueryRoute struct {
	StartNode    string             `json:"start_node"`
	EndNode      string             `json:"end_node"`
	ResponseChan chan RouteResponse `json:"-"`
}

func (q QueryRoute) IsQuery() {}

type QueryWeather struct {
	ResponseChan chan GlobalWeather `json:"-"`
}

func (q QueryWeather) IsQuery() {}

type VehicleUpdate struct {
	VehicleID      string    `json:"vehicle_id"`
	OldPosition    Vector2D  `json:"old_position"`
	NewPosition    Vector2D  `json:"new_position"`
	Velocity       Vector2D  `json:"velocity"`
	CurrentEdge    string    `json:"current_edge"`
	ProgressOnEdge float64   `json:"progress_on_edge"`
	Timestamp      time.Time `json:"timestamp"`
}

type CongestionUpdate struct {
	EdgeID       string    `json:"edge_id"`
	VehicleCount int       `json:"vehicle_count"`
	Timestamp    time.Time `json:"timestamp"`
}

type WeatherChanged struct {
	NewWeather    GlobalWeather `json:"new_weather"`
	AffectedEdges []string      `json:"affected_edges"`
	Timestamp     time.Time     `json:"timestamp"`
}

type TrafficAlert struct {
	EdgeID    string   `json:"edge_id"`
	AlertType string   `json:"alert_type"`
	Severity  Severity `json:"severity"`
}

type SimulationCommand struct {
	Command string                 `json:"command"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

type NearbyVehiclesResponse struct {
	Vehicles  []VehiclePosition `json:"vehicles"`
	Timestamp time.Time         `json:"timestamp"`
}

type EdgeConditionsResponse struct {
	Conditions    RoadConditions `json:"conditions"`
	WeatherEffect WeatherEffects `json:"weather_effect"`
}

type RouteResponse struct {
	Route         *Route        `json:"route,omitempty"`
	EstimatedTime time.Duration `json:"estimated_time"`
	Success       bool          `json:"success"`
	Error         string        `json:"error,omitempty"`
}
