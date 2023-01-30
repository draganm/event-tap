package server_test

import (
	"context"

	"github.com/draganm/event-buffer/client"
	tapClient "github.com/draganm/event-tap/client"
)

type StateKeyType string

const stateKey = StateKeyType("")

type State struct {
	bufferClient  *client.Client
	tapClient     *tapClient.Client
	webhookClient *client.Client
	webhookURL    string
}

func getState(ctx context.Context) *State {
	return ctx.Value(stateKey).(*State)
}
