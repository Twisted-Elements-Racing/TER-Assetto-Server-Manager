package servermanager

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
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

type TERSession struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	TimeMinutes int    `json:"timeMinutes"`
	Laps        int    `json:"laps"`
	WaitSeconds int    `json:"waitSeconds"`
	IsOpen      int    `json:"isOpen"`
}

type TERCurrentEvent struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	Track       string   `json:"track"`
	TrackLayout string   `json:"trackLayout"`
	Cars        []string `json:"cars"`

	MaxClients       int `json:"maxClients"`
	FuelRate         int `json:"fuelRate"`
	DamageMultiplier int `json:"damageMultiplier"`
	TyreWearRate     int `json:"tyreWearRate"`
	LoopMode         int `json:"loopMode"`

	ABSAllowed              int `json:"absAllowed"`
	TractionControlAllowed  int `json:"tractionControlAllowed"`
	StabilityControlAllowed int `json:"stabilityControlAllowed"`
	AutoClutchAllowed       int `json:"autoClutchAllowed"`
	TyreBlanketsAllowed     int `json:"tyreBlanketsAllowed"`

	IsPractice     bool `json:"isPractice"`
	IsTimeAttack   bool `json:"isTimeAttack"`
	IsChampionship bool `json:"isChampionship"`
	IsRaceWeekend  bool `json:"isRaceWeekend"`

	Sessions []TERSession `json:"sessions"`
}

type TERCurrentEventResponse struct {
	Running bool             `json:"running"`
	Event   *TERCurrentEvent `json:"event,omitempty"`
}

type TERSessionStateResponse struct {
	Running bool `json:"running"`
	Ready   bool `json:"ready"`

	SessionIndex        int `json:"sessionIndex"`
	CurrentSessionIndex int `json:"currentSessionIndex"`
	SessionCount        int `json:"sessionCount"`

	Name string `json:"name"`
	Type string `json:"type"`

	Track       string `json:"track"`
	TrackLayout string `json:"trackLayout"`

	TimeMinutes int `json:"timeMinutes"`
	Laps        int `json:"laps"`
	WaitSeconds int `json:"waitSeconds"`

	SessionStartedAt string `json:"sessionStartedAt,omitempty"`

	ElapsedMilliseconds        int64 `json:"elapsedMilliseconds"`
	AssettoElapsedMilliseconds int64 `json:"assettoElapsedMilliseconds"`

	Timed            bool  `json:"timed"`
	RemainingSeconds int64 `json:"remainingSeconds"`
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

func splitACList(value string) []string {
	parts := strings.Split(value, ";")

	result := make(
		[]string,
		0,
		len(parts),
	)

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if part == "" {
			continue
		}

		result = append(
			result,
			part,
		)
	}

	return result
}

func (h *HealthCheck) ServeTERCurrentEvent(
	w http.ResponseWriter,
	r *http.Request,
) {
	response := TERCurrentEventResponse{
		Running: h.process.IsRunning(),
	}

	if response.Running {
		event := h.process.Event()

		if event != nil {
			cfg := event.GetRaceConfig()

			trackLayout := cfg.TrackLayout

			if trackLayout == defaultLayoutName {
				trackLayout = ""
			}

			sessions := make(
				[]TERSession,
				0,
				len(cfg.Sessions),
			)

			sessionConfigs, sessionTypes :=
				cfg.Sessions.AsSliceWithSessionTypes()

			for index, session := range sessionConfigs {

				sessionType :=
					sessionTypes[index]

				sessions = append(
					sessions,
					TERSession{
						Type: sessionType.OriginalString(),

						Name: session.Name,

						TimeMinutes: session.Time,

						Laps: session.Laps,

						WaitSeconds: session.WaitTime,

						IsOpen: int(session.IsOpen),
					},
				)
			}

			response.Event = &TERCurrentEvent{
				Name: event.EventName(),

				Description: event.EventDescription(),

				Track: cfg.Track,

				TrackLayout: trackLayout,

				Cars: splitACList(cfg.Cars),

				MaxClients: cfg.MaxClients,

				LoopMode: cfg.LoopMode,

				FuelRate: cfg.FuelRate,

				DamageMultiplier: cfg.DamageMultiplier,

				TyreWearRate: cfg.TyreWearRate,

				ABSAllowed: int(cfg.ABSAllowed),

				TractionControlAllowed: int(
					cfg.TractionControlAllowed,
				),

				StabilityControlAllowed: cfg.StabilityControlAllowed,

				AutoClutchAllowed: cfg.AutoClutchAllowed,

				TyreBlanketsAllowed: cfg.TyreBlanketsAllowed,

				IsPractice: event.IsPractice(),

				IsTimeAttack: event.IsTimeAttack(),

				IsChampionship: event.IsChampionship(),

				IsRaceWeekend: event.IsRaceWeekend(),

				Sessions: sessions,
			}
		}
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	_ = json.NewEncoder(w).Encode(response)
}

func (h *HealthCheck) ServeTERSessionState(
	w http.ResponseWriter,
	r *http.Request,
) {
	response := TERSessionStateResponse{
		Running: h.process.IsRunning(),
	}

	if response.Running &&
		!h.raceControl.SessionStartTime.IsZero() {

		info := h.raceControl.SessionInfo

		response.Ready = true

		response.SessionIndex = int(info.SessionIndex)
		response.CurrentSessionIndex = int(info.CurrentSessionIndex)
		response.SessionCount = int(info.SessionCount)

		response.Name = info.Name
		response.Type = info.Type.String()

		response.Track = info.Track

		trackLayout := info.TrackConfig

		if trackLayout == defaultLayoutName {
			trackLayout = ""
		}

		response.TrackLayout = trackLayout

		response.TimeMinutes = int(info.Time)
		response.Laps = int(info.Laps)
		response.WaitSeconds = int(info.WaitTime)

		response.SessionStartedAt = h.raceControl.
			SessionStartTime.
			UTC().
			Format(time.RFC3339Nano)

		elapsed := time.Since(
			h.raceControl.SessionStartTime,
		).Milliseconds()

		if elapsed < 0 {
			elapsed = 0
		}

		response.ElapsedMilliseconds = elapsed

		assettoElapsed := int64(
			info.ElapsedMilliseconds,
		)

		if assettoElapsed < 0 {
			assettoElapsed = 0
		}

		response.AssettoElapsedMilliseconds = assettoElapsed

		if info.Time > 0 {
			response.Timed = true

			totalSeconds := int64(info.Time) * 60
			elapsedSeconds := elapsed / 1000
			remaining := totalSeconds - elapsedSeconds

			if remaining < 0 {
				remaining = 0
			}

			response.RemainingSeconds = remaining
		}
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	_ = json.NewEncoder(w).Encode(
		response,
	)
}
