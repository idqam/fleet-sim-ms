package entities

import (
	"sync"
	"time"
)

type Vehicle struct {
	ID              string         `json:"id"`
	Type            VehicleType    `json:"type"`
	State           VehicleState   `json:"state"`
	Route           *AssignedRoute `json:"route,omitempty"`
	AssignedFleetID string         `json:"assigned_fleet_id"`
	Mutex           sync.Mutex     `json:"-"`
	StopChan        chan struct{}  `json:"-"`
}

type VehicleState struct {
	CurrentPosition Vector2D      `json:"current_position"`
	Velocity        Vector2D      `json:"velocity"`
	CurrentEdge     string        `json:"current_edge"`
	ProgressOnEdge  float64       `json:"progress_on_edge"`
	Status          VehicleStatus `json:"status"`
	LastUpdateTime  time.Time     `json:"last_update_time"`
}

type AssignedRoute struct {
	Edges            []string   `json:"edges"`
	CurrentEdgeIndex int        `json:"current_edge_index"`
	CurrentNode      string     `json:"current_node"`
	TargetNode       string     `json:"target_node"`
	StartNode        string     `json:"start_node"`
	EndNode          string     `json:"end_node"`
	StartedAt        time.Time  `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}

type VehiclePosition struct {
	VehicleID string    `json:"vehicle_id"`
	Position  Vector2D  `json:"position"`
	Velocity  Vector2D  `json:"velocity"`
	Timestamp time.Time `json:"timestamp"`
}
