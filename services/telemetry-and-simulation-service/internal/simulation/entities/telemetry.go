package entities

import (
	"time"
)

type TelemetryBatch struct {
	VehicleID string           `json:"vehicle_id"`
	Positions []PositionRecord `json:"positions"`
	Events    []VehicleEvent   `json:"events"`
	StartTime time.Time        `json:"start_time"`
	EndTime   time.Time        `json:"end_time"`
}

type PositionRecord struct {
	Timestamp time.Time `json:"timestamp"`
	Lat       float64   `json:"lat"`
	Lon       float64   `json:"lon"`
	Speed     float64   `json:"speed"`
	Heading   float64   `json:"heading"`
	EdgeID    string    `json:"edge_id"`
}

type VehicleEvent struct {
	VehicleID string                 `json:"vehicle_id"`
	EventType EventType              `json:"event_type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Severity  Severity               `json:"severity"`
}

type TelemetryMetrics struct {
	TotalUpdates     int64   `json:"total_updates"`
	AverageSpeed     float64 `json:"average_speed"`
	DistanceTraveled float64 `json:"distance_traveled"`
	EventCount       int     `json:"event_count"`
}
