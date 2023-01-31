package server

import (
	"fmt"
	"net/http"

	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/draganm/event-buffer/client"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

type Server struct {
	db  bolted.Database
	log logr.Logger
	http.Handler

	bufferClient *client.Client
}

var tapsPath = dbpath.ToPath("taps")

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

	prometheus.Register(newStatsCollector(db, log))

	s := &Server{
		Handler:      r,
		db:           db,
		log:          log,
		bufferClient: bufferClient,
	}

	r.Methods("POST").Path("/taps").HandlerFunc(s.create)
	r.Methods("GET").Path("/taps").HandlerFunc(s.list)

	return s, nil
}
