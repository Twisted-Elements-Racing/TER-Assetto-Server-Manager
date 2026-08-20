package servermanager

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

type TERStoredEvent struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Track       string   `json:"track"`
	TrackLayout string   `json:"trackLayout"`
	Cars        []string `json:"cars"`

	MaxClients int  `json:"maxClients"`
	LoopMode   int  `json:"loopMode"`
	AutoLoop   bool `json:"autoLoop"`

	Sessions []TERSession `json:"sessions"`
}

type TEREventsResponse struct {
	Events []TERStoredEvent `json:"events"`
}

type TERStartRequest struct {
	EventID string `json:"eventId"`
}

type TERStartResponse struct {
	OK            bool           `json:"ok"`
	Action        string         `json:"action"`
	ServerRunning bool           `json:"serverRunning"`
	Event         TERStoredEvent `json:"event"`
}

func terStoredEventFromCustomRace(
	race *CustomRace,
) TERStoredEvent {
	trackLayout := race.RaceConfig.TrackLayout

	if trackLayout == defaultLayoutName {
		trackLayout = ""
	}

	sessionConfigs, sessionTypes :=
		race.RaceConfig.Sessions.AsSliceWithSessionTypes()

	sessions := make(
		[]TERSession,
		0,
		len(sessionConfigs),
	)

	for index, session := range sessionConfigs {
		sessionType := sessionTypes[index]

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

	return TERStoredEvent{
		ID: race.UUID.String(),

		Name: race.Name,

		Track: race.RaceConfig.Track,

		TrackLayout: trackLayout,

		Cars: splitACList(
			race.RaceConfig.Cars,
		),

		MaxClients: race.RaceConfig.MaxClients,

		LoopMode: race.RaceConfig.LoopMode,

		AutoLoop: race.IsLooping(),

		Sessions: sessions,
	}
}

func (crh *CustomRaceHandler) ServeTEREvents(
	w http.ResponseWriter,
	r *http.Request,
) {
	recent, _, _, _, err :=
		crh.raceManager.ListCustomRaces()

	if err != nil {
		logrus.WithError(err).Error(
			"TER API could not list custom races",
		)

		writeTERJSON(
			w,
			http.StatusInternalServerError,
			TERControlErrorResponse{
				OK:    false,
				Error: "failed to list events",
			},
		)
		return
	}

	events := make(
		[]TERStoredEvent,
		0,
		len(recent),
	)

	for _, race := range recent {
		events = append(
			events,
			terStoredEventFromCustomRace(race),
		)
	}

	writeTERJSON(
		w,
		http.StatusOK,
		TEREventsResponse{
			Events: events,
		},
	)
}

func (crh *CustomRaceHandler) ServeTERStart(
	w http.ResponseWriter,
	r *http.Request,
) {
	if crh.raceManager.process.IsRunning() {
		writeTERJSON(
			w,
			http.StatusConflict,
			TERControlErrorResponse{
				OK:    false,
				Error: "server is already running",
			},
		)
		return
	}

	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		4096,
	)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var request TERStartRequest

	if err := decoder.Decode(&request); err != nil {
		writeTERJSON(
			w,
			http.StatusBadRequest,
			TERControlErrorResponse{
				OK:    false,
				Error: "invalid request body",
			},
		)
		return
	}

	request.EventID = strings.TrimSpace(
		request.EventID,
	)

	if request.EventID == "" {
		writeTERJSON(
			w,
			http.StatusBadRequest,
			TERControlErrorResponse{
				OK:    false,
				Error: "eventId is required",
			},
		)
		return
	}

	race, err := crh.store.FindCustomRaceByID(
		request.EventID,
	)

	if err != nil {
		if errors.Is(
			err,
			ErrCustomRaceNotFound,
		) {
			writeTERJSON(
				w,
				http.StatusNotFound,
				TERControlErrorResponse{
					OK:    false,
					Error: "event not found",
				},
			)
			return
		}

		logrus.WithError(err).Error(
			"TER API could not load custom race",
		)

		writeTERJSON(
			w,
			http.StatusInternalServerError,
			TERControlErrorResponse{
				OK:    false,
				Error: "failed to load event",
			},
		)
		return
	}

	/*
		TER controls whether the server itself is running.

		Server Manager Auto Loop must therefore remain OFF
		for TER-started events.

		This does NOT modify Assetto Corsa LOOP_MODE.
	*/
	if race.LoopServer != nil &&
		race.LoopServer[serverID] {

		race.LoopServer[serverID] = false

		if err := crh.store.UpsertCustomRace(
			race,
		); err != nil {
			logrus.WithError(err).Error(
				"TER API could not disable custom race auto loop",
			)

			writeTERJSON(
				w,
				http.StatusInternalServerError,
				TERControlErrorResponse{
					OK:    false,
					Error: "failed to prepare event",
				},
			)
			return
		}
	}

	startedRace, err :=
		crh.raceManager.StartCustomRace(
			request.EventID,
			false,
		)

	if err != nil {
		logrus.WithError(err).Error(
			"TER API could not start custom race",
		)

		writeTERJSON(
			w,
			http.StatusInternalServerError,
			TERControlErrorResponse{
				OK:    false,
				Error: "failed to start event",
			},
		)
		return
	}

	writeTERJSON(
		w,
		http.StatusOK,
		TERStartResponse{
			OK:            true,
			Action:        "start",
			ServerRunning: crh.raceManager.process.IsRunning(),

			Event: terStoredEventFromCustomRace(
				startedRace,
			),
		},
	)
}
