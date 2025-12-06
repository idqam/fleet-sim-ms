package entities

import (
	"time"
)

type BasicVehiclePosEvent struct {
	VehicleID  string    `json:"vehicle_id"`
	EdgeID     string    `json:"edge_id"`
	FromNodeID string    `json:"from_node_id"`
	Progress   float64   `json:"progress"`
	Timestamp  time.Time `json:"timestamp"`
}

type VehicleEvent struct {
	VehicleID string                 `json:"vehicle_id"`
	EventType EventType              `json:"event_type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Severity  Severity               `json:"severity"`
}

type TelemetryMetrics struct {
	TotalUpdates     int64     `json:"total_updates"`
	DistanceTraveled float64   `json:"distance_traveled"`
	EventCount       int       `json:"event_count"`
	LastUpdated      time.Time `json:"last_updated"`
}
