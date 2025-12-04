package entities

import (
	"time"
)

type GlobalWeather struct {
	Condition     WeatherCondition `json:"condition"`
	Intensity     float64          `json:"intensity"`
	WindSpeed     float64          `json:"wind_speed"`
	WindDirection float64          `json:"wind_direction"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

type WeatherEffects struct {
	SpeedMultiplier        float64 `json:"speed_multiplier"`
	BrakingMultiplier      float64 `json:"braking_multiplier"`
	VisibilityRange        float64 `json:"visibility_range"`
	AccelerationMultiplier float64 `json:"acceleration_multiplier"`
}

type WeatherTransition struct {
	FromCondition WeatherCondition `json:"from_condition"`
	ToCondition   WeatherCondition `json:"to_condition"`
	Progress      float64          `json:"progress"`
	Duration      time.Duration    `json:"duration"`
}
