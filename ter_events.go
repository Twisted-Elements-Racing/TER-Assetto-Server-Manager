package servermanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/go-chi/chi"
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

type TEREventSessionDetail struct {
	Enabled     bool   `json:"enabled"`
	Name        string `json:"name"`
	TimeMinutes int    `json:"timeMinutes"`
	Laps        int    `json:"laps"`
	IsOpen      int    `json:"isOpen"`
	WaitSeconds int    `json:"waitSeconds"`
}

type TEREventAssistsDetail struct {
	ABS              int `json:"abs"`
	TractionControl  int `json:"tractionControl"`
	StabilityControl int `json:"stabilityControl"`
	AutoClutch       int `json:"autoClutch"`
	TyreBlankets     int `json:"tyreBlankets"`
}

type TEREventRealismDetail struct {
	LegalTyres         []string `json:"legalTyres"`
	FuelRate           int      `json:"fuelRate"`
	DamageMultiplier   int      `json:"damageMultiplier"`
	TyreWearRate       int      `json:"tyreWearRate"`
	ForceVirtualMirror int      `json:"forceVirtualMirror"`
}

type TEREventDetailResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Track       string   `json:"track"`
	TrackLayout string   `json:"trackLayout"`
	Cars        []string `json:"cars"`

	MaxClients int  `json:"maxClients"`
	LoopMode   int  `json:"loopMode"`
	AutoLoop   bool `json:"autoLoop"`

	Practice TEREventSessionDetail `json:"practice"`
	Qualify  TEREventSessionDetail `json:"qualify"`
	Race     TEREventSessionDetail `json:"race"`
	Booking  TEREventSessionDetail `json:"booking"`

	Assists TEREventAssistsDetail `json:"assists"`
	Realism TEREventRealismDetail `json:"realism"`

	AllowedTyresOut         int `json:"allowedTyresOut"`
	MaxContactsPerKilometer int `json:"maxContactsPerKilometer"`
	StartRule               int `json:"startRule"`
	ResultScreenTime        int `json:"resultScreenTime"`

	OverridePassword              bool `json:"overridePassword"`
	ReplacementPasswordConfigured bool `json:"replacementPasswordConfigured"`
	ForceStopTime                 int  `json:"forceStopTime"`
	ForceStopWithDrivers          bool `json:"forceStopWithDrivers"`
}

type TERDeleteEventResponse struct {
	OK      bool   `json:"ok"`
	Action  string `json:"action"`
	EventID string `json:"eventId"`
}

type TEREventSessionRequest struct {
	Enabled     bool   `json:"enabled"`
	Name        string `json:"name"`
	TimeMinutes int    `json:"timeMinutes"`
	Laps        int    `json:"laps"`
	IsOpen      *int   `json:"isOpen"`
	WaitSeconds int    `json:"waitSeconds"`
}

type TEREventAssistsRequest struct {
	ABS              *int `json:"abs"`
	TractionControl  *int `json:"tractionControl"`
	StabilityControl *int `json:"stabilityControl"`
	AutoClutch       *int `json:"autoClutch"`
	TyreBlankets     *int `json:"tyreBlankets"`
}

type TEREventRealismRequest struct {
	FuelRate           *int `json:"fuelRate"`
	DamageMultiplier   *int `json:"damageMultiplier"`
	TyreWearRate       *int `json:"tyreWearRate"`
	ForceVirtualMirror *int `json:"forceVirtualMirror"`
}

type TERCreateEventRequest struct {
	Name        string   `json:"name"`
	Track       string   `json:"track"`
	TrackLayout string   `json:"trackLayout"`
	Cars        []string `json:"cars"`

	MaxClients *int `json:"maxClients"`
	LoopMode   *int `json:"loopMode"`

	Practice TEREventSessionRequest `json:"practice"`
	Qualify  TEREventSessionRequest `json:"qualify"`
	Race     TEREventSessionRequest `json:"race"`
	Booking  TEREventSessionRequest `json:"booking"`

	Assists TEREventAssistsRequest `json:"assists"`
	Realism TEREventRealismRequest `json:"realism"`

	AllowedTyresOut         *int `json:"allowedTyresOut"`
	MaxContactsPerKilometer *int `json:"maxContactsPerKilometer"`
	StartRule               *int `json:"startRule"`
	ResultScreenTime        *int `json:"resultScreenTime"`

	OverridePassword     bool   `json:"overridePassword"`
	ReplacementPassword  string `json:"replacementPassword"`
	ForceStopTime        int    `json:"forceStopTime"`
	ForceStopWithDrivers bool   `json:"forceStopWithDrivers"`
}

type TEREventSessionUpdateRequest struct {
	Enabled     *bool   `json:"enabled"`
	Name        *string `json:"name"`
	TimeMinutes *int    `json:"timeMinutes"`
	Laps        *int    `json:"laps"`
	IsOpen      *int    `json:"isOpen"`
	WaitSeconds *int    `json:"waitSeconds"`
}

type TERUpdateEventRequest struct {
	Name        *string   `json:"name"`
	Track       *string   `json:"track"`
	TrackLayout *string   `json:"trackLayout"`
	Cars        *[]string `json:"cars"`

	MaxClients *int `json:"maxClients"`
	LoopMode   *int `json:"loopMode"`

	Practice *TEREventSessionUpdateRequest `json:"practice"`
	Qualify  *TEREventSessionUpdateRequest `json:"qualify"`
	Race     *TEREventSessionUpdateRequest `json:"race"`
	Booking  *TEREventSessionUpdateRequest `json:"booking"`

	Assists *TEREventAssistsRequest `json:"assists"`
	Realism *TEREventRealismRequest `json:"realism"`

	AllowedTyresOut         *int `json:"allowedTyresOut"`
	MaxContactsPerKilometer *int `json:"maxContactsPerKilometer"`
	StartRule               *int `json:"startRule"`
	ResultScreenTime        *int `json:"resultScreenTime"`

	OverridePassword     *bool   `json:"overridePassword"`
	ReplacementPassword  *string `json:"replacementPassword"`
	ForceStopTime        *int    `json:"forceStopTime"`
	ForceStopWithDrivers *bool   `json:"forceStopWithDrivers"`
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

func terEventSessionDetail(
	race *CustomRace,
	sessionType SessionType,
) TEREventSessionDetail {
	session, ok := race.RaceConfig.Sessions[sessionType]

	if !ok {
		return TEREventSessionDetail{
			Enabled: false,
		}
	}

	return TEREventSessionDetail{
		Enabled:     true,
		Name:        session.Name,
		TimeMinutes: session.Time,
		Laps:        session.Laps,
		IsOpen:      int(session.IsOpen),
		WaitSeconds: session.WaitTime,
	}
}

func terEventDetailFromCustomRace(
	race *CustomRace,
) TEREventDetailResponse {
	trackLayout := race.RaceConfig.TrackLayout

	if trackLayout == defaultLayoutName {
		trackLayout = ""
	}

	return TEREventDetailResponse{
		ID: race.UUID.String(),

		Name: race.Name,

		Track:       race.RaceConfig.Track,
		TrackLayout: trackLayout,

		Cars: splitACList(
			race.RaceConfig.Cars,
		),

		MaxClients: race.RaceConfig.MaxClients,
		LoopMode:   race.RaceConfig.LoopMode,
		AutoLoop:   race.IsLooping(),

		Practice: terEventSessionDetail(
			race,
			SessionTypePractice,
		),

		Qualify: terEventSessionDetail(
			race,
			SessionTypeQualifying,
		),

		Race: terEventSessionDetail(
			race,
			SessionTypeRace,
		),

		Booking: terEventSessionDetail(
			race,
			SessionTypeBooking,
		),

		Assists: TEREventAssistsDetail{
			ABS: int(
				race.RaceConfig.ABSAllowed,
			),

			TractionControl: int(
				race.RaceConfig.
					TractionControlAllowed,
			),

			StabilityControl: race.RaceConfig.
				StabilityControlAllowed,

			AutoClutch: race.RaceConfig.
				AutoClutchAllowed,

			TyreBlankets: race.RaceConfig.
				TyreBlanketsAllowed,
		},

		Realism: TEREventRealismDetail{
			LegalTyres: splitACList(
				race.RaceConfig.LegalTyres,
			),

			FuelRate: race.RaceConfig.FuelRate,

			DamageMultiplier: race.RaceConfig.
				DamageMultiplier,

			TyreWearRate: race.RaceConfig.TyreWearRate,

			ForceVirtualMirror: race.RaceConfig.
				ForceVirtualMirror,
		},

		AllowedTyresOut: race.RaceConfig.AllowedTyresOut,

		MaxContactsPerKilometer: race.RaceConfig.
			MaxContactsPerKilometer,

		StartRule: int(race.RaceConfig.StartRule),

		ResultScreenTime: race.RaceConfig.ResultScreenTime,

		OverridePassword: race.OverridePassword,

		ReplacementPasswordConfigured: race.ReplacementPassword != "",

		ForceStopTime: race.ForceStopTime,

		ForceStopWithDrivers: race.ForceStopWithDrivers,
	}
}

func applyTERSession(
	cfg *CurrentRaceConfig,
	sessionType SessionType,
	request TEREventSessionRequest,
) {
	cfg.RemoveSession(sessionType)

	if !request.Enabled {
		return
	}

	name := strings.TrimSpace(request.Name)

	if name == "" {
		name = sessionType.String()
	}

	isOpen := SessionOpenness(1)

	if request.IsOpen != nil {
		isOpen = SessionOpenness(
			*request.IsOpen,
		)
	}

	cfg.AddSession(
		sessionType,
		&SessionConfig{
			Name:     name,
			Time:     request.TimeMinutes,
			Laps:     request.Laps,
			IsOpen:   isOpen,
			WaitTime: request.WaitSeconds,
		},
	)
}

func applyTERSessionUpdate(
	cfg *CurrentRaceConfig,
	sessionType SessionType,
	request *TEREventSessionUpdateRequest,
) {
	if request == nil {
		return
	}

	if request.Enabled != nil &&
		!*request.Enabled {
		cfg.RemoveSession(sessionType)
		return
	}

	session, ok := cfg.Sessions[sessionType]

	if !ok {
		session = &SessionConfig{
			Name:   sessionType.String(),
			IsOpen: SessionOpennessFreeJoin,
		}

		cfg.AddSession(
			sessionType,
			session,
		)
	}

	if request.Name != nil {
		name := strings.TrimSpace(
			*request.Name,
		)

		if name == "" {
			name = sessionType.String()
		}

		session.Name = name
	}

	if request.TimeMinutes != nil {
		session.Time =
			*request.TimeMinutes
	}

	if request.Laps != nil {
		session.Laps =
			*request.Laps
	}

	if request.IsOpen != nil {
		session.IsOpen =
			SessionOpenness(
				*request.IsOpen,
			)
	}

	if request.WaitSeconds != nil {
		session.WaitTime =
			*request.WaitSeconds
	}
}

func (crh *CustomRaceHandler) buildTEROpenEntryList(
	cars []string,
	maxClients int,
) (EntryList, error) {
	allCars, err :=
		crh.raceManager.carManager.ListCars()

	if err != nil {
		return nil, err
	}

	carMap := allCars.AsMap()
	entryList := make(EntryList)

	for i := 0; i < maxClients; i++ {
		carID := strings.TrimSpace(
			cars[i%len(cars)],
		)

		skins, ok := carMap[carID]

		if !ok {
			return nil, fmt.Errorf(
				"unknown car: %s",
				carID,
			)
		}

		entrant := NewEntrant()

		entrant.Model = carID

		if len(skins) > 0 {
			entrant.Skin =
				skins[i%len(skins)]
		} else {
			entrant.Skin = "random_skin"
		}

		entryList.AddInPitBox(
			entrant,
			i,
		)
	}

	return entryList, nil
}

func (crh *CustomRaceHandler) resizeTEREntryList(
	entryList EntryList,
	cars []string,
	maxClients int,
) (EntryList, error) {
	resized, err :=
		crh.buildTEROpenEntryList(
			cars,
			maxClients,
		)

	if err != nil {
		return nil, err
	}

	for _, entrant := range entryList.AsSlice() {
		if entrant.PitBox < 0 ||
			entrant.PitBox >= maxClients {
			continue
		}

		delete(
			resized,
			fmt.Sprintf(
				"CAR_%d",
				entrant.PitBox,
			),
		)

		resized.AddInPitBox(
			entrant,
			entrant.PitBox,
		)
	}

	return resized, nil
}

func buildTERLegalTyres(
	cars []string,
) (string, error) {
	tyres, err := ListTyres()

	if err != nil {
		return "", err
	}

	seen := make(map[string]bool)
	var legalTyres []string

	for _, carID := range cars {
		carID = strings.TrimSpace(carID)

		carTyres, ok := tyres[carID]

		if !ok {
			continue
		}

		for tyre := range carTyres {
			if seen[tyre] {
				continue
			}

			seen[tyre] = true
			legalTyres = append(
				legalTyres,
				tyre,
			)
		}
	}

	sort.Strings(legalTyres)

	return strings.Join(
		legalTyres,
		";",
	), nil
}

func (crh *CustomRaceHandler) ServeTERCreateEvent(
	w http.ResponseWriter,
	r *http.Request,
) {
	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		1024*1024,
	)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var request TERCreateEventRequest

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

	request.Name = strings.TrimSpace(request.Name)
	request.Track = strings.TrimSpace(request.Track)
	request.TrackLayout = strings.TrimSpace(request.TrackLayout)

	if request.Name == "" {
		writeTERJSON(
			w,
			http.StatusBadRequest,
			TERControlErrorResponse{
				OK:    false,
				Error: "name is required",
			},
		)
		return
	}

	if request.Track == "" {
		writeTERJSON(
			w,
			http.StatusBadRequest,
			TERControlErrorResponse{
				OK:    false,
				Error: "track is required",
			},
		)
		return
	}

	if len(request.Cars) == 0 {
		writeTERJSON(
			w,
			http.StatusBadRequest,
			TERControlErrorResponse{
				OK:    false,
				Error: "at least one car is required",
			},
		)
		return
	}

	defaults := ConfigIniDefault()
	cfg := defaults.CurrentRaceConfig

	// TER-created races default Stability Control allowed to on.
	cfg.StabilityControlAllowed = 1
	// TER-created races default Virtual Mirror to off.
	cfg.ForceVirtualMirror = 0

	cfg.Cars = strings.Join(request.Cars, ";")
	cfg.Track = request.Track
	cfg.TrackLayout = request.TrackLayout

	legalTyres, err :=
		buildTERLegalTyres(
			request.Cars,
		)

	if err != nil {
		logrus.WithError(err).Error(
			"TER API could not determine legal tyres",
		)

		writeTERJSON(
			w,
			http.StatusInternalServerError,
			TERControlErrorResponse{
				OK:    false,
				Error: "failed to determine legal tyres",
			},
		)
		return
	}

	cfg.LegalTyres = legalTyres

	if request.MaxClients != nil &&
		*request.MaxClients > 0 {
		cfg.MaxClients = *request.MaxClients
	}

	if request.LoopMode != nil {
		cfg.LoopMode = *request.LoopMode
	}

	if request.Assists.ABS != nil {
		cfg.ABSAllowed = FactoryAssist(
			*request.Assists.ABS,
		)
	}

	if request.Assists.TractionControl != nil {
		cfg.TractionControlAllowed = FactoryAssist(
			*request.Assists.TractionControl,
		)
	}

	if request.Assists.StabilityControl != nil {
		cfg.StabilityControlAllowed =
			*request.Assists.StabilityControl
	}

	if request.Assists.AutoClutch != nil {
		cfg.AutoClutchAllowed =
			*request.Assists.AutoClutch
	}

	if request.Assists.TyreBlankets != nil {
		cfg.TyreBlanketsAllowed =
			*request.Assists.TyreBlankets
	}

	if request.Realism.FuelRate != nil {
		cfg.FuelRate =
			*request.Realism.FuelRate
	}

	if request.Realism.DamageMultiplier != nil {
		cfg.DamageMultiplier =
			*request.Realism.DamageMultiplier
	}

	if request.Realism.TyreWearRate != nil {
		cfg.TyreWearRate =
			*request.Realism.TyreWearRate
	}

	if request.Realism.ForceVirtualMirror != nil {
		cfg.ForceVirtualMirror =
			*request.Realism.ForceVirtualMirror
	}

	if request.AllowedTyresOut != nil {
		cfg.AllowedTyresOut =
			*request.AllowedTyresOut
	}

	if request.MaxContactsPerKilometer != nil {
		cfg.MaxContactsPerKilometer =
			*request.MaxContactsPerKilometer
	}

	if request.StartRule != nil {
		cfg.StartRule =
			StartRule(*request.StartRule)
	}

	if request.ResultScreenTime != nil {
		cfg.ResultScreenTime =
			*request.ResultScreenTime
	}

	applyTERSession(
		&cfg,
		SessionTypePractice,
		request.Practice,
	)

	applyTERSession(
		&cfg,
		SessionTypeQualifying,
		request.Qualify,
	)

	applyTERSession(
		&cfg,
		SessionTypeRace,
		request.Race,
	)

	applyTERSession(
		&cfg,
		SessionTypeBooking,
		request.Booking,
	)

	if len(cfg.Sessions) == 0 {
		writeTERJSON(
			w,
			http.StatusBadRequest,
			TERControlErrorResponse{
				OK:    false,
				Error: "at least one session must be enabled",
			},
		)
		return
	}

	entryList, err :=
		crh.buildTEROpenEntryList(
			request.Cars,
			cfg.MaxClients,
		)

	if err != nil {
		logrus.WithError(err).Error(
			"TER API could not build entry list",
		)

		writeTERJSON(
			w,
			http.StatusBadRequest,
			TERControlErrorResponse{
				OK:    false,
				Error: "failed to build entry list",
			},
		)
		return
	}

	race, err := crh.raceManager.SaveCustomRace(
		request.Name,
		request.OverridePassword,
		request.ReplacementPassword,
		cfg,
		entryList,
		false,
		request.ForceStopTime,
		request.ForceStopWithDrivers,
	)

	if err != nil {
		logrus.WithError(err).Error(
			"TER API could not create custom race",
		)

		writeTERJSON(
			w,
			http.StatusInternalServerError,
			TERControlErrorResponse{
				OK:    false,
				Error: "failed to create event",
			},
		)
		return
	}

	writeTERJSON(
		w,
		http.StatusCreated,
		terStoredEventFromCustomRace(race),
	)
}

func (crh *CustomRaceHandler) ServeTEREvent(
	w http.ResponseWriter,
	r *http.Request,
) {
	eventID := strings.TrimSpace(
		chi.URLParam(r, "id"),
	)

	if eventID == "" {
		writeTERJSON(
			w,
			http.StatusBadRequest,
			TERControlErrorResponse{
				OK:    false,
				Error: "event id is required",
			},
		)
		return
	}

	race, err := crh.findTERActiveEvent(
		eventID,
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

	writeTERJSON(
		w,
		http.StatusOK,
		terEventDetailFromCustomRace(race),
	)
}

func (crh *CustomRaceHandler) ServeTERUpdateEvent(
	w http.ResponseWriter,
	r *http.Request,
) {
	eventID := strings.TrimSpace(
		chi.URLParam(r, "id"),
	)

	if eventID == "" {
		writeTERJSON(
			w,
			http.StatusBadRequest,
			TERControlErrorResponse{
				OK:    false,
				Error: "event id is required",
			},
		)
		return
	}

	race, err := crh.findTERActiveEvent(
		eventID,
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
			"TER API could not load custom race for update",
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

	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		1024*1024,
	)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var request TERUpdateEventRequest

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

	cfg := race.RaceConfig
	entryList := race.EntryList

	carsChanged := false
	maxClientsChanged := false

	if request.Name != nil {
		name := strings.TrimSpace(
			*request.Name,
		)

		if name == "" {
			writeTERJSON(
				w,
				http.StatusBadRequest,
				TERControlErrorResponse{
					OK:    false,
					Error: "name cannot be empty",
				},
			)
			return
		}

		race.Name = name
		race.HasCustomName = true
	}

	if request.Track != nil {
		track := strings.TrimSpace(
			*request.Track,
		)

		if track == "" {
			writeTERJSON(
				w,
				http.StatusBadRequest,
				TERControlErrorResponse{
					OK:    false,
					Error: "track cannot be empty",
				},
			)
			return
		}

		cfg.Track = track
	}

	if request.TrackLayout != nil {
		cfg.TrackLayout =
			strings.TrimSpace(
				*request.TrackLayout,
			)
	}

	if request.Cars != nil {
		if len(*request.Cars) == 0 {
			writeTERJSON(
				w,
				http.StatusBadRequest,
				TERControlErrorResponse{
					OK:    false,
					Error: "at least one car is required",
				},
			)
			return
		}

		cars := make(
			[]string,
			0,
			len(*request.Cars),
		)

		for _, car := range *request.Cars {
			car = strings.TrimSpace(car)

			if car == "" {
				continue
			}

			cars = append(
				cars,
				car,
			)
		}

		if len(cars) == 0 {
			writeTERJSON(
				w,
				http.StatusBadRequest,
				TERControlErrorResponse{
					OK:    false,
					Error: "at least one valid car is required",
				},
			)
			return
		}

		cfg.Cars = strings.Join(
			cars,
			";",
		)

		legalTyres, err :=
			buildTERLegalTyres(cars)

		if err != nil {
			logrus.WithError(err).Error(
				"TER API could not determine legal tyres",
			)

			writeTERJSON(
				w,
				http.StatusInternalServerError,
				TERControlErrorResponse{
					OK:    false,
					Error: "failed to determine legal tyres",
				},
			)
			return
		}

		cfg.LegalTyres = legalTyres
		carsChanged = true
	}

	if request.MaxClients != nil {
		if *request.MaxClients <= 0 {
			writeTERJSON(
				w,
				http.StatusBadRequest,
				TERControlErrorResponse{
					OK:    false,
					Error: "maxClients must be greater than zero",
				},
			)
			return
		}

		if cfg.MaxClients !=
			*request.MaxClients {
			maxClientsChanged = true
		}

		cfg.MaxClients =
			*request.MaxClients
	}

	if request.LoopMode != nil {
		cfg.LoopMode =
			*request.LoopMode
	}

	if request.Assists != nil {
		if request.Assists.ABS != nil {
			cfg.ABSAllowed =
				FactoryAssist(
					*request.Assists.ABS,
				)
		}

		if request.Assists.TractionControl != nil {
			cfg.TractionControlAllowed =
				FactoryAssist(
					*request.Assists.
						TractionControl,
				)
		}

		if request.Assists.StabilityControl != nil {
			cfg.StabilityControlAllowed =
				*request.Assists.
					StabilityControl
		}

		if request.Assists.AutoClutch != nil {
			cfg.AutoClutchAllowed =
				*request.Assists.AutoClutch
		}

		if request.Assists.TyreBlankets != nil {
			cfg.TyreBlanketsAllowed =
				*request.Assists.TyreBlankets
		}
	}

	if request.Realism != nil {
		if request.Realism.FuelRate != nil {
			cfg.FuelRate =
				*request.Realism.FuelRate
		}

		if request.Realism.DamageMultiplier != nil {
			cfg.DamageMultiplier =
				*request.Realism.
					DamageMultiplier
		}

		if request.Realism.TyreWearRate != nil {
			cfg.TyreWearRate =
				*request.Realism.TyreWearRate
		}

		if request.Realism.ForceVirtualMirror != nil {
			cfg.ForceVirtualMirror =
				*request.Realism.
					ForceVirtualMirror
		}
	}

	if request.AllowedTyresOut != nil {
		cfg.AllowedTyresOut =
			*request.AllowedTyresOut
	}

	if request.MaxContactsPerKilometer != nil {
		cfg.MaxContactsPerKilometer =
			*request.MaxContactsPerKilometer
	}

	if request.StartRule != nil {
		cfg.StartRule =
			StartRule(*request.StartRule)
	}

	if request.ResultScreenTime != nil {
		cfg.ResultScreenTime =
			*request.ResultScreenTime
	}

	applyTERSessionUpdate(
		&cfg,
		SessionTypePractice,
		request.Practice,
	)

	applyTERSessionUpdate(
		&cfg,
		SessionTypeQualifying,
		request.Qualify,
	)

	applyTERSessionUpdate(
		&cfg,
		SessionTypeRace,
		request.Race,
	)

	applyTERSessionUpdate(
		&cfg,
		SessionTypeBooking,
		request.Booking,
	)

	if len(cfg.Sessions) == 0 {
		writeTERJSON(
			w,
			http.StatusBadRequest,
			TERControlErrorResponse{
				OK:    false,
				Error: "at least one session must be enabled",
			},
		)
		return
	}

	if carsChanged {
		cars := splitACList(
			cfg.Cars,
		)

		entryList, err =
			crh.buildTEROpenEntryList(
				cars,
				cfg.MaxClients,
			)

		if err != nil {
			logrus.WithError(err).Error(
				"TER API could not rebuild entry list",
			)

			writeTERJSON(
				w,
				http.StatusBadRequest,
				TERControlErrorResponse{
					OK:    false,
					Error: "failed to rebuild entry list",
				},
			)
			return
		}
	} else if maxClientsChanged {
		cars := splitACList(
			cfg.Cars,
		)

		entryList, err =
			crh.resizeTEREntryList(
				entryList,
				cars,
				cfg.MaxClients,
			)

		if err != nil {
			logrus.WithError(err).Error(
				"TER API could not resize entry list",
			)

			writeTERJSON(
				w,
				http.StatusBadRequest,
				TERControlErrorResponse{
					OK:    false,
					Error: "failed to resize entry list",
				},
			)
			return
		}
	}

	if request.OverridePassword != nil {
		race.OverridePassword =
			*request.OverridePassword
	}

	if request.ReplacementPassword != nil {
		race.ReplacementPassword =
			*request.ReplacementPassword
	}

	if request.ForceStopTime != nil {
		race.ForceStopTime =
			*request.ForceStopTime
	}

	if request.ForceStopWithDrivers != nil {
		race.ForceStopWithDrivers =
			*request.ForceStopWithDrivers
	}

	race.RaceConfig = cfg
	race.EntryList = entryList

	if err := crh.store.UpsertCustomRace(
		race,
	); err != nil {
		logrus.WithError(err).Error(
			"TER API could not update custom race",
		)

		writeTERJSON(
			w,
			http.StatusInternalServerError,
			TERControlErrorResponse{
				OK:    false,
				Error: "failed to update event",
			},
		)
		return
	}

	writeTERJSON(
		w,
		http.StatusOK,
		terEventDetailFromCustomRace(race),
	)
}

func (crh *CustomRaceHandler) ServeTERDeleteEvent(
	w http.ResponseWriter,
	r *http.Request,
) {
	eventID := strings.TrimSpace(
		chi.URLParam(r, "id"),
	)

	if eventID == "" {
		writeTERJSON(
			w,
			http.StatusBadRequest,
			TERControlErrorResponse{
				OK:    false,
				Error: "event id is required",
			},
		)
		return
	}

	_, err := crh.findTERActiveEvent(
		eventID,
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
			"TER API could not load custom race for deletion",
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

	err = crh.raceManager.DeleteCustomRace(
		eventID,
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
			"TER API could not delete custom race",
		)

		writeTERJSON(
			w,
			http.StatusInternalServerError,
			TERControlErrorResponse{
				OK:    false,
				Error: "failed to delete event",
			},
		)
		return
	}

	writeTERJSON(
		w,
		http.StatusOK,
		TERDeleteEventResponse{
			OK:      true,
			Action:  "delete",
			EventID: eventID,
		},
	)
}

func (crh *CustomRaceHandler) findTERActiveEvent(
	eventID string,
) (*CustomRace, error) {
	race, err := crh.store.FindCustomRaceByID(
		eventID,
	)

	if err != nil {
		return nil, err
	}

	if !race.Deleted.IsZero() {
		return nil, ErrCustomRaceNotFound
	}

	return race, nil
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
