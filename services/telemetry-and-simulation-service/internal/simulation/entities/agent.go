package entities

import (
	"time"
)

type VehicleAgent struct {
	Vehicle           *Vehicle
	CoordinatorQuery  chan<- MapQuery
	CoordinatorUpdate chan<- VehicleUpdate
	WeatherBroadcast  <-chan WeatherChanged
	TelemetryEmitter  TelemetryEmitter
	Config            AgentConfig
	State             AgentState
}

type AgentConfig struct {
	PhysicsTickRate          time.Duration `json:"physics_tick_rate"`
	TelemetryInterval        time.Duration `json:"telemetry_interval"`
	CollisionAvoidanceRadius float64       `json:"collision_avoidance_radius"`
	MaxQueryTimeout          time.Duration `json:"max_query_timeout"`
}

type AgentState struct {
	LastPhysicsUpdate    time.Time         `json:"last_physics_update"`
	LastTelemetryEmit    time.Time         `json:"last_telemetry_emit"`
	CachedWeather        *GlobalWeather    `json:"cached_weather"`
	CachedNearbyVehicles []VehiclePosition `json:"cached_nearby_vehicles"`
}

type DecisionState struct {
	DesiredSpeed   float64 `json:"desired_speed"`
	DesiredHeading float64 `json:"desired_heading"`
	IsAvoiding     bool    `json:"is_avoiding"`
	IsBraking      bool    `json:"is_braking"`
	Reason         string  `json:"reason"`
}

type TelemetryEmitter interface {
	EmitPosition(vehicleID string, position PositionRecord) error
	EmitEvent(vehicleID string, event VehicleEvent) error
	EmitBatch(batch TelemetryBatch) error
}
