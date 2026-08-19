package servermanager

import (
	"encoding/json"
	"net/http"
)

type TERStatusResponse struct {
	Service          string   `json:"service"`
	Version          string   `json:"version"`
	ServerRunning    bool     `json:"serverRunning"`
	EventInProgress  bool     `json:"eventInProgress"`
	ConnectedDrivers int      `json:"connectedDrivers"`
	ServerName       string   `json:"serverName"`
	ServerID         ServerID `json:"serverId"`
	AssettoInstalled bool     `json:"assettoInstalled"`
}

func (h *HealthCheck) ServeTERStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	opts, err := h.store.LoadServerOptions()

	var serverName string

	if err == nil {
		serverName = opts.Name
	}

	running := h.process.IsRunning()

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	_ = json.NewEncoder(w).Encode(
		TERStatusResponse{
			Service:          "ter-ac-server-manager",
			Version:          BuildVersion,
			ServerRunning:    running,
			EventInProgress:  running,
			ConnectedDrivers: h.raceControl.ConnectedDrivers.Len(),
			ServerName:       serverName,
			ServerID:         serverID,
			AssettoInstalled: IsAssettoInstalled(),
		},
	)
}
