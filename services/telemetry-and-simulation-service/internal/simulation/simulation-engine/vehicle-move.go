package simulationengine

import (
	"errors"
	"time"

	"github.com/m/internal/simulation/entities"
)

func UpdateVehiclePosition(vehicle *entities.Vehicle, graph *entities.MapGraph, delta float64) error {

	if vehicle.Route == nil {
		return nil
	}

	if vehicle.Route.CompletedAt != nil {
		vehicle.State.Status = entities.VehicleStatusArrived
		return nil
	}

	if vehicle.Route.CurrentEdgeIndex >= len(vehicle.Route.Edges) {
		now := time.Now()
		vehicle.Route.CompletedAt = &now
		vehicle.State.Status = entities.VehicleStatusArrived
		vehicle.State.Velocity = entities.Vector2D{X: 0, Y: 0}
		vehicle.State.CurrentPosition = entities.Vector2D{X: graph.Nodes[vehicle.Route.EndNode].Position.X, Y: graph.Nodes[vehicle.Route.EndNode].Position.Y}
		return nil
	}

	edge := graph.Edges[vehicle.Route.Edges[vehicle.Route.CurrentEdgeIndex]]
	if edge == nil {
		return errors.New("Edge not found")
	}

	speed := edge.Conditions.EffectiveSpeedLimit

	if speed <= 0 {
		speed = edge.BaseSpeedLimit
	}

	displacement := speed * delta

	progressIncrement := displacement / edge.Length

	vehicle.State.ProgressOnEdge += progressIncrement

	if vehicle.State.ProgressOnEdge >= 0.999999 {
		vehicle.Route.CurrentNode = vehicle.Route.TargetNode
		vehicle.Route.CurrentEdgeIndex += 1

		if vehicle.Route.CurrentEdgeIndex < len(vehicle.Route.Edges) {

			vehicle.State.ProgressOnEdge = 0.0

			vehicle.State.CurrentEdge = vehicle.Route.Edges[vehicle.Route.CurrentEdgeIndex]
			nextEdge := graph.Edges[vehicle.State.CurrentEdge]

			if nextEdge != nil {
				vehicle.Route.TargetNode = nextEdge.To
				edge = nextEdge
			} else {
				vehicle.State.ProgressOnEdge = 1.0
				vehicle.State.CurrentEdge = ""
			}
		} else {
			now := time.Now()
			vehicle.Route.CompletedAt = &now
			vehicle.State.Status = entities.VehicleStatusArrived
			vehicle.State.Velocity = entities.Vector2D{X: 0, Y: 0}
			vehicle.State.ProgressOnEdge = 1.0
			vehicle.State.CurrentPosition = entities.Vector2D{
				X: graph.Nodes[vehicle.Route.EndNode].Position.X,
				Y: graph.Nodes[vehicle.Route.EndNode].Position.Y,
			}
			return nil
		}
	}

	fromNode := graph.Nodes[edge.From]
	toNode := graph.Nodes[edge.To]

	progress := clamp(vehicle.State.ProgressOnEdge, 0.0, 1.0)

	vehicle.State.CurrentPosition = interpolatePosition(fromNode, toNode, progress)
	vehicle.State.Velocity = calculateVelocity(fromNode, toNode, speed)

	vehicle.State.LastUpdateTime = time.Now()

	if vehicle.State.Status != entities.VehicleStatusArrived {
		vehicle.State.Status = entities.VehicleStatusMoving
	}

	return nil
}

func interpolatePosition(from, to *entities.MapNode, progress float64) entities.Vector2D {

	resX := from.Position.X + (to.Position.X-from.Position.X)*progress
	resY := from.Position.Y + (to.Position.Y-from.Position.Y)*progress

	return entities.Vector2D{X: resX, Y: resY}

}

func calculateVelocity(from, to *entities.MapNode, speed float64) entities.Vector2D {

	dx := to.Position.X - from.Position.X
	dy := to.Position.Y - from.Position.Y

	d := distance(from.Position, to.Position)

	if d == 0 {
		return entities.Vector2D{X: 0, Y: 0}
	}

	normalizedX := dx / d
	normalizedY := dy / d

	return entities.Vector2D{
		X: normalizedX * speed,
		Y: normalizedY * speed,
	}
}
