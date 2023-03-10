package server_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/draganm/event-buffer/client"
	tapClient "github.com/draganm/event-tap/client"
	"github.com/draganm/event-tap/data"
	"github.com/draganm/event-tap/server/testrig"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

func init() {
	logger, _ := zap.NewDevelopment()
	if true {
		opts.DefaultContext = logr.NewContext(context.Background(), zapr.NewLogger(logger))
	}
}

var opts = godog.Options{
	Output:        os.Stdout,
	StopOnFailure: true,
	Strict:        true,
	Format:        "progress",
	Paths:         []string{"features"},
	NoColors:      true,
	Concurrency:   runtime.NumCPU(),
}

func init() {
	godog.BindCommandLineFlags("godog.", &opts)
}

func TestMain(m *testing.M) {
	pflag.Parse()
	opts.Paths = pflag.Args()

	status := godog.TestSuite{
		Name:                "godogs",
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}.Run()

	os.Exit(status)
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	var cancel context.CancelFunc

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		ctx, cancel = context.WithTimeout(ctx, 2*time.Second)

		return ctx, nil

	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		cancel()
		return ctx, nil
	})

	state := &State{}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		log := logr.FromContextOrDiscard(ctx)

		// buffer

		bufferServerURL, err := testrig.StartBufferServer(ctx, log)
		if err != nil {
			return ctx, fmt.Errorf("could not start server: %w", err)
		}

		bufferClient, err := client.New(bufferServerURL)
		if err != nil {
			return ctx, fmt.Errorf("could not create buffer client: %w", err)
		}

		state.bufferClient = bufferClient

		// tap

		tapServerURL, err := testrig.StartServer(ctx, log, bufferServerURL)
		if err != nil {
			return ctx, fmt.Errorf("could not create tap client: %w", err)
		}

		tapClient, err := tapClient.New(tapServerURL)
		if err != nil {
			return ctx, fmt.Errorf("could not create tap client: %w", err)
		}

		state.tapClient = tapClient

		// webhook

		webhookServerURL, err := testrig.StartBufferServer(ctx, log)
		if err != nil {
			return ctx, fmt.Errorf("could not start webhook: %w", err)
		}

		webhookClient, err := client.New(webhookServerURL)
		if err != nil {
			return ctx, fmt.Errorf("could not create buffer client for webhook: %w", err)
		}

		state.webhookClient = webhookClient
		state.webhookURL, err = url.JoinPath(webhookServerURL, "events")
		if err != nil {
			return ctx, fmt.Errorf("could not create webhookURL path")
		}

		ctx = context.WithValue(ctx, stateKey, state)

		return ctx, nil
	})

	ctx.Step(`^I create a new map of events$`, iCreateANewMapOfEvents)
	ctx.Step(`^one event in the buffer$`, oneEventInTheBuffer)
	ctx.Step(`^the receiver should receive that event as webhook$`, theReceiverShouldReceiveThatEventAsWebhook)
	ctx.Step(`^I list the taps$`, iListTheTaps)
	ctx.Step(`^the result should be empty$`, theResultShouldBeEmpty)
	ctx.Step(`^there are no taps$`, thereAreNoTaps)
	ctx.Step(`^there is one tap$`, thereIsOneTap)
	ctx.Step(`^the result should have one tap$`, theResultShouldHaveOneTap)
	ctx.Step(`^I delete the tap$`, iDeleteTheTap)
	ctx.Step(`^the list of taps should not contain the deleted tap$`, theListOfTapsShouldNotContainTheDeletedTap)

}

func iCreateANewMapOfEvents(ctx context.Context) error {
	s := getState(ctx)
	_, err := s.tapClient.CreateTap(ctx, tapClient.CreateTapOptions{
		Name:       "tap1",
		Code:       `function mapEvents(evts){return evts.map(([id, evt]) => evt)}`,
		WebhookURL: s.webhookURL,
		BatchLimit: 20,
	})

	if err != nil {
		return fmt.Errorf("could not create tap: %w", err)
	}

	return nil
}

func oneEventInTheBuffer(ctx context.Context) error {
	s := getState(ctx)
	err := s.bufferClient.SendEvents(ctx, []any{"evt1"})
	if err != nil {
		return fmt.Errorf("could not send event: %w", err)
	}
	return nil
}

func theReceiverShouldReceiveThatEventAsWebhook(ctx context.Context) error {
	s := getState(ctx)
	evts := []any{}
	_, err := s.webhookClient.PollForEvents(ctx, "", 1, &evts)
	if err != nil {
		return fmt.Errorf("failed polling for webhook events: %w", err)
	}
	diff := cmp.Diff(evts, []any{"evt1"})
	if diff != "" {
		return fmt.Errorf("diff:\n%s", diff)
	}
	return nil
}

func thereAreNoTaps() error {
	// already the case on the beginning of a test
	return nil
}

func iListTheTaps(ctx context.Context) (err error) {
	s := getState(ctx)
	s.listResult, err = s.tapClient.List(ctx)
	if err != nil {
		return fmt.Errorf("could not list taps: %w", err)
	}
	return nil
}

func theResultShouldBeEmpty(ctx context.Context) error {
	s := getState(ctx)
	diff := cmp.Diff(s.listResult, []data.TapListEntry{})
	if diff != "" {
		return fmt.Errorf("diff:\n%s", diff)
	}
	return nil
}

func thereIsOneTap(ctx context.Context) error {
	s := getState(ctx)
	id, err := s.tapClient.CreateTap(ctx, tapClient.CreateTapOptions{
		Name:       "tap1",
		Code:       `function mapEvents(evts){return evts.map(([id, evt]) => evt)}`,
		WebhookURL: s.webhookURL,
		BatchLimit: 20,
	})

	if err != nil {
		return fmt.Errorf("could not create tap: %w", err)
	}

	s.createdTapID = id

	return nil
}

func theResultShouldHaveOneTap(ctx context.Context) error {
	s := getState(ctx)
	ln := len(s.listResult)
	if ln != 1 {
		return fmt.Errorf("expected one tap, but got %d", ln)
	}
	return nil
}

func iDeleteTheTap(ctx context.Context) error {
	s := getState(ctx)
	return s.tapClient.Delete(ctx, s.createdTapID)
}

func theListOfTapsShouldNotContainTheDeletedTap(ctx context.Context) error {
	s := getState(ctx)
	listResult, err := s.tapClient.List(ctx)
	if err != nil {
		return fmt.Errorf("could not list taps: %w", err)
	}

	if len(listResult) != 0 {
		return fmt.Errorf("tap was not deleted")
	}

	return nil
}
