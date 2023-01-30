package testrig

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/draganm/bolted/embedded"
	"github.com/draganm/event-tap/server"
	"github.com/go-logr/logr"
)

func StartServer(ctx context.Context, log logr.Logger, buferBaseURL string) (string, error) {
	td, err := os.MkdirTemp("", "")
	if err != nil {
		return "", fmt.Errorf("could not create temp dir: %w", err)
	}

	db, err := embedded.Open(filepath.Join(td, "db"), 0700, embedded.Options{})
	if err != nil {
		return "", fmt.Errorf("could not open db: %w", err)
	}

	server, err := server.New(log, db, buferBaseURL)
	if err != nil {
		return "", fmt.Errorf("could not start tap server: %w", err)
	}

	hs := httptest.NewServer(server)

	go func() {
		<-ctx.Done()
		hs.Close()
		db.Close()
		os.RemoveAll(td)
	}()

	return hs.URL, nil
}
