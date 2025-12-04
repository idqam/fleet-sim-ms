package entities


type Vector2D struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
}

type VehicleType string

const (
    VehicleTypSedan VehicleType = "car"
    VehicleTypeTruck VehicleType = "truck"
    VehicleTypeDrone VehicleType = "drone"
)

type VehicleStatus string

const (
    VehicleStatusIdle      VehicleStatus = "idle"
    VehicleStatusMoving    VehicleStatus = "moving"
    VehicleStatusStopped   VehicleStatus = "stopped"
    VehicleStatusArrived   VehicleStatus = "arrived"
    VehicleStatusBreakdown VehicleStatus = "breakdown"
)

type WeatherCondition string

const (
    WeatherClear WeatherCondition = "clear"
    WeatherRain  WeatherCondition = "rain"
    WeatherSnow  WeatherCondition = "snow"
    WeatherFog   WeatherCondition = "fog"
)


type NodeType string

const (
    NodeTypeIntersection NodeType = "intersection"
    NodeTypeWaypoint     NodeType = "waypoint"
    NodeTypeParking      NodeType = "parking"
    NodeTypeDepot        NodeType = "depot"
)

type EventType string

const (
    EventPositionUpdate     EventType = "position_update"
    EventRouteStarted       EventType = "route_started"
    EventRouteCompleted     EventType = "route_completed"
    EventWeatherChanged     EventType = "weather_changed"
    EventCollisionAverted   EventType = "collision_averted"
    EventBreakdownOccurred  EventType = "breakdown_occurred"
    EventEnergyLow          EventType = "energy_low"
    EventTrafficCongestion  EventType = "traffic_congestion"
)


type Severity string

const (
    SeverityInfo     Severity = "info"
    SeverityWarning  Severity = "warning"
    SeverityCritical Severity = "critical"
)