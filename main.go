package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/draganm/bolted/embedded"
	"github.com/draganm/event-tap/server"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

func main() {
	logger, _ := zap.Config{
		Encoding:    "json",
		Level:       zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "message",
			LevelKey:     "level",
			EncodeLevel:  zapcore.CapitalLevelEncoder,
			TimeKey:      "time",
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}.Build()

	defer logger.Sync()

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "state-file",
				Value:   "state",
				EnvVars: []string{"STATE_FILE"},
			},
			&cli.StringFlag{
				Name:    "api-addr",
				Value:   ":6677",
				EnvVars: []string{"API_ADDR"},
			},
			&cli.StringFlag{
				Name:    "metrics-addr",
				Value:   ":3000",
				EnvVars: []string{"METRICS_ADDR"},
			},
			&cli.StringFlag{
				Name:    "event-buffer-base-url",
				Value:   "http://localhost:5566",
				EnvVars: []string{"EVENT_BUFFER_BASE_URL"},
			},
		},
		Action: func(c *cli.Context) error {
			log := zapr.NewLogger(logger)
			defer log.Info("server exiting")
			eg, ctx := errgroup.WithContext(context.Background())

			db, err := embedded.Open(c.String("state-file"), 0700, embedded.Options{})
			if err != nil {
				return fmt.Errorf("could not open state: %w", err)
			}

			// signal handler
			eg.Go(func() error {
				sigChan := make(chan os.Signal, 1)
				signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
				select {
				case sig := <-sigChan:
					log.Info("received signal", "signal", sig.String())
					return fmt.Errorf("received signal %s", sig.String())
				case <-ctx.Done():
					return nil
				}
			})

			// api server
			s, err := server.New(log, db, c.String("event-buffer-base-url"))
			if err != nil {
				return fmt.Errorf("could not start server: %w", err)
			}

			eg.Go(runHttp(ctx, log, c.String("api-addr"), "api", s))

			// run metrics server
			metricsRouter := mux.NewRouter()
			metricsRouter.Methods("GET").Path("/metrics").Handler(promhttp.Handler())
			eg.Go(runHttp(ctx, log, c.String("metrics-addr"), "metrics", metricsRouter))

			return eg.Wait()

		},
	}
	app.RunAndExitOnError()

}

func runHttp(ctx context.Context, log logr.Logger, addr, name string, handler http.Handler) func() error {

	return func() error {
		l, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("could not listen for %s requests: %w", name, err)

		}

		s := &http.Server{
			Handler: handler,
		}

		go func() {
			<-ctx.Done()
			shutdownContext, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			log.Info(fmt.Sprintf("graceful shutdown of the %s server", name))
			err := s.Shutdown(shutdownContext)
			if errors.Is(err, context.DeadlineExceeded) {
				log.Info(fmt.Sprintf("%s server did not shut down gracefully, forcing close", name))
				s.Close()
			}
		}()

		log.Info(fmt.Sprintf("%s server started", name), "addr", l.Addr().String())
		return s.Serve(l)
	}
}
