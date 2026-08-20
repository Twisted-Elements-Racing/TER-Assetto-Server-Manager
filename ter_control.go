package servermanager

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

type TERStopResponse struct {
	OK             bool   `json:"ok"`
	Action         string `json:"action"`
	ServerRunning  bool   `json:"serverRunning"`
	AlreadyStopped bool   `json:"alreadyStopped"`
}

type TERControlErrorResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

func writeTERJSON(
	w http.ResponseWriter,
	status int,
	value interface{},
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}

func (sah *ServerAdministrationHandler) stopActiveServerEvent() error {
	if !sah.process.IsRunning() {
		return nil
	}

	event := sah.process.Event()

	if event.IsChampionship() && !event.IsPractice() {
		return sah.championshipManager.StopActiveEvent()
	}

	if event.IsRaceWeekend() && !event.IsPractice() {
		return sah.raceWeekendManager.StopActiveSession()
	}

	return sah.process.Stop()
}

func (sah *ServerAdministrationHandler) ServeTERStop(
	w http.ResponseWriter,
	r *http.Request,
) {
	alreadyStopped := !sah.process.IsRunning()

	if !alreadyStopped {
		if err := sah.stopActiveServerEvent(); err != nil {
			logrus.WithError(err).Error(
				"TER API could not stop active server event",
			)

			writeTERJSON(
				w,
				http.StatusInternalServerError,
				TERControlErrorResponse{
					OK:    false,
					Error: "failed to stop server",
				},
			)
			return
		}
	}

	writeTERJSON(
		w,
		http.StatusOK,
		TERStopResponse{
			OK:             true,
			Action:         "stop",
			ServerRunning:  sah.process.IsRunning(),
			AlreadyStopped: alreadyStopped,
		},
	)
}
