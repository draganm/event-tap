package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/draganm/bolted"
	"github.com/draganm/event-tap/data"
	"github.com/draganm/event-tap/server/webhook"
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

	webhookPath := tapsPath.Append(id.String())

	err = bolted.SugaredWrite(s.db, func(tx bolted.SugaredWriteTx) error {
		tx.CreateMap(webhookPath)
		tx.Put(webhookPath.Append("options"), tcd)
		return nil
	})

	if err != nil {
		http.Error(w, fmt.Errorf("could not store tap config: %w", err).Error(), http.StatusInternalServerError)
		log.Error(err, "clould not store tap config")
		return
	}

	err = webhook.Start(context.Background(), log, s.db, webhookPath, s.bufferClient)
	if err != nil {
		http.Error(w, fmt.Errorf("could start webhook: %w", err).Error(), http.StatusInternalServerError)
		log.Error(err, "clould start webhook")
		return
	}

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(createTapResponse{ID: id.String()})
}
