package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dop251/goja"
	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/draganm/event-buffer/client"
	"github.com/draganm/event-tap/data"
	"github.com/go-logr/logr"
)

type options data.TapOptions

func Start(ctx context.Context, log logr.Logger, db bolted.Database, path dbpath.Path, bufferClient *client.Client) error {
	opts := options{}
	err := bolted.SugaredRead(db, func(tx bolted.SugaredReadTx) error {
		return json.Unmarshal(tx.Get(path.Append("options")), &opts)
	})

	if err != nil {
		return fmt.Errorf("could not load webhook options: %w", err)
	}

	log = log.WithValues("webhook", opts.Name)

	prg, err := goja.Compile("webhook.js", opts.Code, true)
	if err != nil {
		return fmt.Errorf("could not parse webhook code: %w", err)
	}

	lastID := ""

	lastIDPath := path.Append("last_id")

	err = bolted.SugaredRead(db, func(tx bolted.SugaredReadTx) error {

		if tx.Exists(lastIDPath) {
			lastID = string(tx.Get(lastIDPath))
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("could not determine last ID: %w", err)
	}

	rt := goja.New()
	_, err = rt.RunProgram(prg)
	if err != nil {
		log.Error(err, "running code failed")
		return fmt.Errorf("could not run code: %w", err)
	}

	mapEvents, ok := goja.AssertFunction(rt.Get("mapEvents"))
	if !ok {
		return fmt.Errorf("could not find mapEvents function")
	}

	updateStatus := func(status string) {
		err := bolted.SugaredWrite(db, func(tx bolted.SugaredWriteTx) error {
			tx.Put(path.Append("status"), []byte(status))
			return nil
		})
		if err != nil {
			log.Error(err, "could not update status")
		}
	}

	postWebhook := func(ctx context.Context, payload any) error {
		d, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("could not marshal payload: %w", err)
		}

		req, err := http.NewRequest("POST", opts.WebhookURL, bytes.NewReader(d))
		if err != nil {
			return fmt.Errorf("could not create request: %w", err)
		}

		req.Header.Set("content-type", "application/json")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("could not perform request: %w", err)
		}

		defer res.Body.Close()

		if !(res.StatusCode == http.StatusOK || res.StatusCode == http.StatusAccepted) {
			rd, _ := io.ReadAll(res.Body)
			return fmt.Errorf("unexpected status %s: %s", res.Status, string(rd))
		}

		return nil

	}

	updateLastID := func(lastID string) error {
		err := bolted.SugaredWrite(db, func(tx bolted.SugaredWriteTx) error {
			tx.Put(lastIDPath, []byte(lastID))
			return nil
		})
		if err != nil {
			log.Error(err, "could not update last id")
		}
		return nil
	}

	go func() (err error) {
		defer func() {
			if err != nil {
				log.Error(err, "webhook failed")
			} else {
				log.Info("webhook terminated")
			}
		}()

	mainLoop:
		for ctx.Err() == nil {

			events := []any{}

			ids, err := bufferClient.PollForEvents(ctx, lastID, opts.BatchLimit, &events)
			if err != nil {
				log.Error(err, "polling events failed")
				updateStatus(fmt.Errorf("could not poll events: %w", err).Error())
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(time.Second):
					// try again
				}
			}

			eventsWithIDs := make([][]any, len(events))

			for i, ev := range events {
				id := ids[i]
				eventsWithIDs[i] = []any{id, ev}
			}

			jsResult, err := mapEvents(goja.Undefined(), rt.ToValue(eventsWithIDs))
			if err != nil {
				log.Error(err, "mapEvents failed")
				updateStatus(fmt.Errorf("mapEvents failed: %w", err).Error())
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(time.Second):
					continue mainLoop
				}
			}

			var result []any
			err = rt.ExportTo(jsResult, &result)
			if err != nil {
				log.Error(err, "exportValues failed")
				updateStatus(fmt.Errorf("exportValues failed: %w", err).Error())
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(time.Second):
					continue mainLoop
				}
			}

			if len(result) > 0 {

				err = postWebhook(ctx, result)
				if err != nil {
					log.Error(err, "postWebhook failed")
					updateStatus(fmt.Errorf("postWebhook failed: %w", err).Error())
					select {
					case <-ctx.Done():
						return nil
					case <-time.After(time.Second):
						continue mainLoop
					}
				}
			}

			if len(ids) > 0 {
				newLastID := ids[len(ids)-1]
				err = updateLastID(newLastID)
				if err != nil {
					log.Error(err, "updating last id failed")
					updateStatus(fmt.Errorf("updating last id failed: %w", err).Error())
					select {
					case <-ctx.Done():
						return nil
					case <-time.After(time.Second):
						continue mainLoop
					}
				} else {
					lastID = newLastID
				}
			}

		}

		return nil

	}()

	return nil

}
