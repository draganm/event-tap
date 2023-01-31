package server

import (
	"context"
	"fmt"

	"github.com/draganm/bolted"
	"github.com/draganm/event-tap/server/webhook"
)

func (s *Server) startWebhooks(ctx context.Context) error {
	return bolted.SugaredRead(s.db, func(tx bolted.SugaredReadTx) error {
		for it := tx.Iterator(tapsPath); !it.IsDone(); it.Next() {
			err := webhook.Start(ctx, s.log, s.db, tapsPath.Append(it.GetKey()), s.bufferClient)
			if err != nil {
				return fmt.Errorf("could not start webhook for %s: %w", it.GetKey(), err)
			}
		}
		return nil
	})
}
