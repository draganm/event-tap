package server_test

import (
	"context"

	"github.com/draganm/event-buffer/client"
	tapClient "github.com/draganm/event-tap/client"
	"github.com/draganm/event-tap/data"
)

type StateKeyType string

const stateKey = StateKeyType("")

type State struct {
	bufferClient  *client.Client
	tapClient     *tapClient.Client
	webhookClient *client.Client
	webhookURL    string
	listResult    []data.TapListEntry
}

func getState(ctx context.Context) *State {
	return ctx.Value(stateKey).(*State)
}
