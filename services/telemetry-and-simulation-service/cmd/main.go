package cmd

import (
	"fmt"
	"net/http"
	"time"

	"github.com/m/internal/ext"
	"github.com/m/internal/simulation/entities"
	simulationengine "github.com/m/internal/simulation/simulation-engine"
)

func main() {

	config := simulationengine.NewMapGenerator(2000, 2000, 12345, simulationengine.AlgoDelaunay, 100, 0)
	graph := config.Generate()

	engine := simulationengine.NewSimulationEngine(graph, 100*time.Millisecond)

	spawnConfig := &simulationengine.VehicleSpawnConfig{
		SpawnStrategy:  simulationengine.SpawnRandom,
		TargetStrategy: simulationengine.TargetRandom,
		AllowSameNode:  false,
	}

	for i := 0; i < 10; i++ {
		vehicle := &entities.Vehicle{
			ID:    fmt.Sprintf("vehicle-%d", i),
			Type:  entities.VehicleTypSedan,
			State: entities.VehicleState{Status: entities.VehicleStatusIdle},
		}
		simulationengine.AssignVehicleRoute(vehicle, graph, spawnConfig)
		engine.AddVehicle(vehicle)
	}

	engine.Start()

	api := &ext.SimulationAPI{Engine: engine}
	http.HandleFunc("/api/simulation/start", api.StartSimulation)
	http.HandleFunc("/api/simulation/stop", api.StopSimulation)
	http.HandleFunc("/api/simulation/vehicles", api.GetVehicles)

	http.ListenAndServe(":8081", nil)
}
