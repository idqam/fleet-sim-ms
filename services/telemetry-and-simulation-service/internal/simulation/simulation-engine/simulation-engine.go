package simulationengine

import (
	"fmt"
	"sync"
	"time"

	"github.com/m/internal/simulation/entities"
)

type SimulationEngine struct {
	Graph      *entities.MapGraph
	Vehicles   map[string]*entities.Vehicle
	UpdateRate time.Duration
	Mutex      sync.RWMutex
	IsRunning  bool
	wg         sync.WaitGroup
}

type TelemetryEmitterImpl struct {
	// Will send to message queue later
	Events chan entities.BasicVehiclePosEvent
}

func NewSimulationEngine(graph *entities.MapGraph, updateRate time.Duration) *SimulationEngine {
	return &SimulationEngine{
		Graph:      graph,
		Vehicles:   make(map[string]*entities.Vehicle),
		UpdateRate: updateRate,
		IsRunning:  false,
	}
}

func (s *SimulationEngine) Start() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if s.IsRunning {
		return
	}
	s.IsRunning = true

	for _, vehicle := range s.Vehicles {
		s.RunVehicleGoroutine(vehicle)
	}
}

func (s *SimulationEngine) Stop() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if !s.IsRunning {
		return
	}
	s.IsRunning = false

	for _, vehicle := range s.Vehicles {
		s.stopVehicle(vehicle)
	}
	s.wg.Wait()
}

func (s *SimulationEngine) AddVehicle(vehicle *entities.Vehicle) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if _, exists := s.Vehicles[vehicle.ID]; exists {
		return
	}

	vehicle.StopChan = make(chan struct{})

	s.Vehicles[vehicle.ID] = vehicle

	if s.IsRunning {
		s.RunVehicleGoroutine(vehicle)
	}
}

func (s *SimulationEngine) RemoveVehicle(id string) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	vehicle, exists := s.Vehicles[id]
	if !exists {
		return
	}

	if s.IsRunning {
		s.stopVehicle(vehicle)
	}

	delete(s.Vehicles, id)
}

func (s *SimulationEngine) stopVehicle(vehicle *entities.Vehicle) {
	select {
	case <-vehicle.StopChan:
	default:
		close(vehicle.StopChan)
	}
}

func (s *SimulationEngine) RunVehicleGoroutine(vehicle *entities.Vehicle) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		ticker := time.NewTicker(s.UpdateRate)
		defer ticker.Stop()

		lastUpdate := time.Now()
		lastTelemetryEmit := time.Now()

		for {
			select {
			case <-ticker.C:
				now := time.Now()
				dt := now.Sub(lastUpdate).Seconds()
				lastUpdate = now

				vehicle.Mutex.Lock()
				err := UpdateVehiclePosition(vehicle, s.Graph, dt)
				vehicle.Mutex.Unlock()

				if now.Sub(lastTelemetryEmit) >= 1*time.Second {
					s.emitTelemetry(vehicle)
					lastTelemetryEmit = now
				}

				if err != nil {
				}

				if vehicle.Route != nil && vehicle.Route.CompletedAt != nil {
					return
				}

			case <-vehicle.StopChan:
				return
			}
		}
	}()
}

func (t *TelemetryEmitterImpl) EmitPosition(event entities.BasicVehiclePosEvent) error {
	select {
	case t.Events <- event:
		return nil
	default:
		return fmt.Errorf("telemetry channel full")
	}
}

func (s *SimulationEngine) emitTelemetry(vehicle *entities.Vehicle) {
	if vehicle.Route == nil || len(vehicle.Route.Edges) == 0 {
		return
	}

	currentEdge := vehicle.State.CurrentEdge
	if currentEdge == "" && vehicle.Route.CurrentEdgeIndex > 0 {
		currentEdge = vehicle.Route.Edges[vehicle.Route.CurrentEdgeIndex-1]
	}

	event := entities.BasicVehiclePosEvent{
		VehicleID:  vehicle.ID,
		EdgeID:     currentEdge,
		FromNodeID: vehicle.Route.CurrentNode,
		Progress:   vehicle.State.ProgressOnEdge,
		Timestamp:  time.Now(),
	}

	//  Later: send to Kafka/RabbitMQ/REDIS pub sub
	fmt.Printf("[TELEMETRY] %s | Edge: %s | Progress: %.2f\n",
		vehicle.ID, event.EdgeID, event.Progress)
}
