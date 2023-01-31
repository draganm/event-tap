package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/draganm/bolted"
	"github.com/gorilla/mux"
)

var ErrNotFound = errors.New("not found")

func (s *Server) delete(w http.ResponseWriter, r *http.Request) {

	tapID := mux.Vars(r)["tapID"]
	log := s.log.WithValues("method", r.Method, "path", r.URL.Path, "tapID", tapID)

	err := bolted.SugaredWrite(s.db, func(tx bolted.SugaredWriteTx) error {
		webhookPath := tapsPath.Append(tapID)
		if !tx.Exists(webhookPath) {
			return ErrNotFound
		}
		tx.Delete(webhookPath)
		return nil
	})

	if errors.Is(err, ErrNotFound) {
		http.Error(w, "not found", http.StatusNotFound)
		log.Error(err, "tap not found")
		return
	}

	if err != nil {
		http.Error(w, fmt.Errorf("could not delete tap: %w", err).Error(), http.StatusInternalServerError)
		log.Error(err, "could not delete tap")
		return
	}

	s.mu.Lock()
	cancel, found := s.tapCancels[tapID]
	if found {
		cancel()
	}
	s.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}
