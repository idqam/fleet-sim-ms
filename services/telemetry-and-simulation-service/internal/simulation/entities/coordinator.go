package entities

import (
	"time"
)

type MapCoordinator struct {
	Graph         *MapGraph
	Weather       *GlobalWeather
	SpatialIndex  SpatialIndex
	VehicleStates map[string]VehiclePosition
	Config        CoordinatorConfig
	Metrics       CoordinatorMetrics

	QueryChannel    chan MapQuery
	UpdateChannel   chan VehicleUpdate
	WeatherChannel  chan WeatherChanged
	ShutdownChannel chan struct{}
}

type CoordinatorConfig struct {
	TickRate                 time.Duration `json:"tick_rate"`
	SpatialIndexCellSize     float64       `json:"spatial_index_cell_size"`
	MaxVehicles              int           `json:"max_vehicles"`
	CongestionUpdateInterval time.Duration `json:"congestion_update_interval"`
}

type CoordinatorMetrics struct {
	TotalQueries     int64         `json:"total_queries"`
	QueryLatency     time.Duration `json:"query_latency"`
	UpdatesProcessed int64         `json:"updates_processed"`
	ActiveVehicles   int           `json:"active_vehicles"`
}
