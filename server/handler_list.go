package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/draganm/bolted"
	"github.com/draganm/event-tap/data"
)

func (s *Server) list(w http.ResponseWriter, r *http.Request) {

	limit := 100

	cursor := r.URL.Query().Get("cursor")

	log := s.log.WithValues("method", r.Method, "path", r.URL.Path)

	page := &data.TapListPage{
		Entries: []data.TapListEntry{},
	}

	err := bolted.SugaredRead(s.db, func(tx bolted.SugaredReadTx) error {

		it := tx.Iterator(tapsPath)
		if cursor != "" {
			it.Seek(cursor)
			if !it.IsDone() && it.GetKey() == cursor {
				it.Next()
			}
		}
		for ; !it.IsDone(); it.Next() {
			opts := &data.TapOptions{}
			optsPath := tapsPath.Append(it.GetKey(), "options")
			err := json.Unmarshal(tx.Get(optsPath), &opts)
			if err != nil {
				return fmt.Errorf("could not parse %s: %w", optsPath.String(), err)
			}

			page.Entries = append(page.Entries, data.TapListEntry{
				Name:       opts.Name,
				ID:         it.GetKey(),
				WebhookURL: opts.WebhookURL,
			})

			page.Cursor = it.GetKey()

			if len(page.Entries) >= limit {
				break
			}
		}

		if it.IsDone() {
			page.Cursor = ""
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Errorf("could not list taps: %w", err).Error(), http.StatusInternalServerError)
		log.Error(err, "could not list taps")
		return
	}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(page)

}
