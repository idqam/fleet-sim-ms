package entities

type SpatialIndex interface {
	Insert(vehicleID string, position Vector2D) error
	Remove(vehicleID string) error
	Query(center Vector2D, radius float64) ([]string, error)
	Update(vehicleID string, oldPos, newPos Vector2D) error
}

type GridIndex struct {
	CellSize float64
	Cells    map[CellKey]*Cell
}

type CellKey struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Cell struct {
	Bounds     Bounds   `json:"bounds"`
	VehicleIDs []string `json:"vehicle_ids"`
}

type Bounds struct {
	MinX float64 `json:"min_x"`
	MinY float64 `json:"min_y"`
	MaxX float64 `json:"max_x"`
	MaxY float64 `json:"max_y"`
}

type QuadTree struct {
	Bounds      Bounds
	MaxVehicles int
	Children    [4]*QuadTree
	Vehicles    map[string]Vector2D
}

type QueryResult struct {
	VehiclePositions []VehiclePosition `json:"vehicle_positions"`
	QueryTime        int64             `json:"query_time"`
	CellsChecked     int               `json:"cells_checked"`
}
