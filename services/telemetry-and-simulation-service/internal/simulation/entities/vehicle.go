package entities

import (
	"time"
)

type Vehicle struct {
	ID              string         `json:"id"`
	Type            VehicleType    `json:"type"`
	Physics         VehiclePhysics `json:"physics"`
	State           VehicleState   `json:"state"`
	Route           *AssignedRoute `json:"route,omitempty"`
	AssignedFleetID string         `json:"assigned_fleet_id"`
}

type VehicleState struct {
	CurrentPosition Vector2D      `json:"current_position"`
	Velocity        Vector2D      `json:"velocity"`
	CurrentEdge     string        `json:"current_edge"`
	ProgressOnEdge  float64       `json:"progress_on_edge"`
	Status          VehicleStatus `json:"status"`
	Energy          *EnergySystem `json:"energy,omitempty"`
	LastUpdateTime  time.Time     `json:"last_update_time"`
}

type VehiclePhysics struct {
	MaxSpeed        float64 `json:"max_speed"`
	MaxAcceleration float64 `json:"max_acceleration"`
	MaxDeceleration float64 `json:"max_deceleration"`
}

type AssignedRoute struct {
	Edges            []string  `json:"edges"`
	CurrentEdgeIndex int       `json:"current_edge_index"`
	StartNode        string    `json:"start_node"`
	EndNode          string    `json:"end_node"`
	StartedAt        time.Time `json:"started_at"`
	EstimatedArrival time.Time `json:"estimated_arrival"`
}

type EnergySystem struct {
	MaxCapacity float64 `json:"max_capacity"`
	Current     float64 `json:"current"`
	DrainRate   float64 `json:"drain_rate"`
	Type        string  `json:"type"`
}

type VehiclePosition struct {
	VehicleID string    `json:"vehicle_id"`
	Position  Vector2D  `json:"position"`
	Velocity  Vector2D  `json:"velocity"`
	Heading   float64   `json:"heading"`
	Timestamp time.Time `json:"timestamp"`
}
