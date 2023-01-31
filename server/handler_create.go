package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/draganm/bolted"
	"github.com/draganm/event-tap/data"
	"github.com/draganm/event-tap/server/tap"
	"github.com/gofrs/uuid"
)

type createTapOptions data.TapOptions
type createTapResponse data.TapID

func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	log := s.log.WithValues("method", r.Method, "path", r.URL.Path)

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

	tapPath := tapsPath.Append(id.String())

	err = bolted.SugaredWrite(s.db, func(tx bolted.SugaredWriteTx) error {
		tx.CreateMap(tapPath)
		tx.Put(tapPath.Append("options"), tcd)
		return nil
	})

	if err != nil {
		http.Error(w, fmt.Errorf("could not store tap config: %w", err).Error(), http.StatusInternalServerError)
		log.Error(err, "clould not store tap config")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		if err != nil {
			cancel()
		}
	}()

	err = tap.Start(ctx, log, s.db, tapPath, s.bufferClient)
	if err != nil {
		http.Error(w, fmt.Errorf("could start tap: %w", err).Error(), http.StatusInternalServerError)
		log.Error(err, "clould start tap")
		return
	}

	s.mu.Lock()
	s.tapCancels[id.String()] = cancel
	s.mu.Unlock()

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(createTapResponse{ID: id.String()})
}
