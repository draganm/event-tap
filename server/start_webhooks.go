package server

import (
	"context"
	"fmt"

	"github.com/draganm/bolted"
	"github.com/draganm/event-tap/server/tap"
)

func (s *Server) startTaps(ctx context.Context) error {
	return bolted.SugaredRead(s.db, func(tx bolted.SugaredReadTx) (err error) {
		for it := tx.Iterator(tapsPath); !it.IsDone(); it.Next() {
			ctx, cancel := context.WithCancel(context.Background())
			defer func() {
				if err != nil {
					cancel()
				}
			}()

			err := tap.Start(ctx, s.log, s.db, tapsPath.Append(it.GetKey()), s.bufferClient)
			if err != nil {
				return fmt.Errorf("could not start tap for %s: %w", it.GetKey(), err)
			}

			s.mu.Lock()
			s.tapCancels[it.GetKey()] = cancel
			s.mu.Unlock()

		}
		return nil
	})
}
