package ext

import (
	"encoding/json"
	"net/http"

	"github.com/m/internal/simulation/entities"
	simulationengine "github.com/m/internal/simulation/simulation-engine"
)

type SimulationAPI struct {
	Engine *simulationengine.SimulationEngine
}

func (api *SimulationAPI) StartSimulation(w http.ResponseWriter, r *http.Request) {
	api.Engine.Start()
	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func (api *SimulationAPI) StopSimulation(w http.ResponseWriter, r *http.Request) {
	api.Engine.Stop()
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

func (api *SimulationAPI) GetVehicles(w http.ResponseWriter, r *http.Request) {
	api.Engine.Mutex.RLock()
	defer api.Engine.Mutex.RUnlock()

	vehicles := make([]*entities.Vehicle, 0, len(api.Engine.Vehicles))
	for _, v := range api.Engine.Vehicles {
		vehicles = append(vehicles, v)
	}

	json.NewEncoder(w).Encode(vehicles)
}
