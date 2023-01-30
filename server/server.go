package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/draganm/event-buffer/client"
	"github.com/draganm/event-tap/data"
	"github.com/draganm/event-tap/server/webhook"
	"github.com/go-logr/logr"
	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

type Server struct {
	db  bolted.Database
	log logr.Logger
	http.Handler
}

var tapsPath = dbpath.ToPath("taps")

type createTapOptions data.TapOptions
type createTapResponse data.TapID

func New(log logr.Logger, db bolted.Database, bufferBaseURL string) (*Server, error) {

	bufferClient, err := client.New(bufferBaseURL)
	if err != nil {
		return nil, fmt.Errorf("could not create buffer client: %w", err)
	}

	err = bolted.SugaredWrite(db, func(tx bolted.SugaredWriteTx) error {
		if !tx.Exists(tapsPath) {
			tx.CreateMap(tapsPath)

		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not initialize db: %w", err)
	}

	r := mux.NewRouter()

	r.Methods("POST").Path("/taps").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := log.WithValues("method", r.Method, "path", r.URL.Path)

		cto := createTapOptions{}

		err := json.NewDecoder(r.Body).Decode(&cto)
		if err != nil {
			http.Error(w, fmt.Errorf("could not decode options: %w", err).Error(), http.StatusBadRequest)
			log.Error(err, "clould not decode tap options")
			return
		}

		id, err := uuid.NewV6()
		if err != nil {
			http.Error(w, fmt.Errorf("could not create tap id: %w", err).Error(), http.StatusBadRequest)
			log.Error(err, "clould not create tap id")
			return
		}

		tcd, err := json.Marshal(cto)
		if err != nil {
			http.Error(w, fmt.Errorf("could not marshal tap config: %w", err).Error(), http.StatusBadRequest)
			log.Error(err, "clould not marshal tap config")
			return
		}

		webhookPath := tapsPath.Append(id.String())

		err = bolted.SugaredWrite(db, func(tx bolted.SugaredWriteTx) error {
			tx.CreateMap(webhookPath)
			tx.Put(webhookPath.Append("options"), tcd)
			return nil
		})

		if err != nil {
			http.Error(w, fmt.Errorf("could not store tap config: %w", err).Error(), http.StatusInternalServerError)
			log.Error(err, "clould not store tap config")
			return
		}

		err = webhook.Start(context.Background(), log, db, webhookPath, bufferClient)
		if err != nil {
			http.Error(w, fmt.Errorf("could start webhook: %w", err).Error(), http.StatusInternalServerError)
			log.Error(err, "clould start webhook")
			return
		}

		w.WriteHeader(http.StatusCreated)

		json.NewEncoder(w).Encode(createTapResponse{ID: id.String()})

	})

	prometheus.Register(newStatsCollector(db, log))

	return &Server{
		Handler: r,
		db:      db,
		log:     log,
	}, nil
}
