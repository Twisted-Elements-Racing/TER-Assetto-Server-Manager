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

type TERCar struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Brand string `json:"brand"`
	IsMod bool   `json:"isMod"`
	IsDLC bool   `json:"isDlc"`
}

type TERCarsResponse struct {
	Cars []TERCar `json:"cars"`
}

type TERTrackLayout struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TERTrack struct {
	ID      string           `json:"id"`
	Name    string           `json:"name"`
	Layouts []TERTrackLayout `json:"layouts"`
	IsMod   bool             `json:"isMod"`
	IsDLC   bool             `json:"isDlc"`
}

type TERTracksResponse struct {
	Tracks []TERTrack `json:"tracks"`
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

func (ch *CarsHandler) ServeTERCars(
	w http.ResponseWriter,
	r *http.Request,
) {
	cars, err := ch.carManager.ListCars()

	if err != nil {
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}

	response := TERCarsResponse{
		Cars: make([]TERCar, 0, len(cars)),
	}

	for _, car := range cars {
		displayName := car.Details.Name

		if displayName == "" {
			displayName = car.PrettyName()
		}

		response.Cars = append(
			response.Cars,
			TERCar{
				ID:    car.Name,
				Name:  displayName,
				Brand: car.Details.Brand,
				IsMod: car.IsMod(),
				IsDLC: car.IsPaidDLC(),
			},
		)
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	_ = json.NewEncoder(w).Encode(response)
}

func (th *TracksHandler) ServeTERTracks(
	w http.ResponseWriter,
	r *http.Request,
) {
	tracks, err := th.trackManager.ListTracks()

	if err != nil {
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}

	response := TERTracksResponse{
		Tracks: make(
			[]TERTrack,
			0,
			len(tracks),
		),
	}

	for _, track := range tracks {
		layouts := make(
			[]TERTrackLayout,
			0,
			len(track.Layouts),
		)

		if len(track.Layouts) == 0 {
			layouts = append(
				layouts,
				TERTrackLayout{
					ID:   "",
					Name: "Default",
				},
			)
		}

		for _, layout := range track.Layouts {
			layoutID := layout
			layoutName := prettifyName(
				layout,
				false,
			)

			if layout == defaultLayoutName || layout == "" {
				layoutID = ""
				layoutName = "Default"
			}

			info, err := GetTrackInfo(
				track.Name,
				layout,
			)

			if err == nil && info != nil && info.Name != "" {
				layoutName = info.Name
			}

			layouts = append(
				layouts,
				TERTrackLayout{
					ID:   layoutID,
					Name: layoutName,
				},
			)
		}

		response.Tracks = append(
			response.Tracks,
			TERTrack{
				ID:      track.Name,
				Name:    track.PrettyName(),
				Layouts: layouts,
				IsMod:   track.IsMod(),
				IsDLC:   track.IsPaidDLC(),
			},
		)
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	_ = json.NewEncoder(w).Encode(response)
}
