package entities

type CollisionRadius struct {
	Center Vector2D `json:"center"`
	Radius float64  `json:"radius"`
}

type MovementResult struct {
	NewPosition      Vector2D `json:"new_position"`
	NewVelocity      Vector2D `json:"new_velocity"`
	DistanceTraveled float64  `json:"distance_traveled"`
}
