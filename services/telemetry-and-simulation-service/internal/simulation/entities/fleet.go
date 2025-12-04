package entities

import (
	"time"
)

type FleetStatus string

const (
	FleetStatusActive   FleetStatus = "active"
	FleetStatusInactive FleetStatus = "inactive"
	FleetStatusPaused   FleetStatus = "paused"
)

type Fleet struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Status      FleetStatus            `json:"status"`
	Vehicles    []Vehicle              `json:"vehicles"`
	VehicleIDs  []string               `json:"vehicle_ids"`
	Config      FleetConfig            `json:"config"`
	Metrics     FleetMetrics           `json:"metrics"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type FleetConfig struct {
	MaxVehicles          int           `json:"max_vehicles"`
	DefaultVehicleType   VehicleType   `json:"default_vehicle_type"`
	EnergyThresholdAlert float64       `json:"energy_threshold_alert"`
}

type FleetMetrics struct {
	TotalVehicles         int       `json:"total_vehicles"`
	ActiveVehicles        int       `json:"active_vehicles"`
	IdleVehicles          int       `json:"idle_vehicles"`
	VehiclesInTransit     int       `json:"vehicles_in_transit"`
	VehiclesBreakdown     int       `json:"vehicles_breakdown"`
	TotalDistanceTraveled float64   `json:"total_distance_traveled"`
	AverageSpeed          float64   `json:"average_speed"`
	TotalEnergyConsumed   float64   `json:"total_energy_consumed"`
	CompletedRoutes       int       `json:"completed_routes"`
	ActiveRoutes          int       `json:"active_routes"`
	TotalEvents           int       `json:"total_events"`
	CriticalEvents        int       `json:"critical_events"`
	LastUpdated           time.Time `json:"last_updated"`
}

type FleetAssignment struct {
	FleetID    string    `json:"fleet_id"`
	VehicleID  string    `json:"vehicle_id"`
	AssignedAt time.Time `json:"assigned_at"`
	AssignedBy string    `json:"assigned_by"`
	Priority   int       `json:"priority"`
}
