package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/spacemeshos/go-spacemesh/cmd"
	"github.com/spacemeshos/go-spacemesh/cmd/beaconalarm/metrics"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/config"
	"github.com/spacemeshos/go-spacemesh/log"
	"github.com/spacemeshos/go-spacemesh/node"
	"github.com/spacemeshos/go-spacemesh/timesync"
)

var serverURL string

func main() {
	if err := getCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "alarm",
		Short: "start alarm",
		Run: func(c *cobra.Command, args []string) {
			conf, err := loadConfig(c)
			if err != nil {
				log.With().Fatal("failed to initialize config", log.Err(err))
			}

			if conf.LOGGING.Encoder == config.JSONLogEncoder {
				log.JSONLog(true)
			}

			// Create a context for graceful shutdown
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
			defer cancel()

			// Start the clock
			clock, err := setupClock()
			if err != nil {
				log.With().Fatal("failed to setup clock", log.Err(err))
			}

			// Start the server
			server, err := metrics.NewServer(":8080")
			if err != nil {
				log.With().Fatal("failed to create server", log.Err(err))
			}
			if err := server.Start(); err != nil {
				log.With().Fatal("failed to start server", log.Err(err))
			}

			// Start the client
			client, err := metrics.NewClient(serverURL, clock)
			if err != nil {
				shutdownServer(server)
				log.With().Fatal("failed to create client", log.Err(err))
			}

			var eg errgroup.Group
			eg.Go(func() error {
				for {
					select {
					case <-ctx.Done():
						shutdownServer(server)
						return nil
					case <-time.After(5 * time.Second):
						// Fetch the metric value from the Prometheus server
						_, err := client.FetchBeaconValue(ctx, 11)
						if err != nil {
							log.With().Error("failed to beacon value", log.Err(err))
						}
					}
				}
			})

			if err := eg.Wait(); err != nil {
				log.With().Fatal("failed to wait for goroutines", log.Err(err))
			}
		},
	}

	cmd.AddCommands(c)
	c.PersistentFlags().StringVar(&serverURL, "prometheus", "http://localhost:9090", "Prometheus server URL")
	return c
}

func loadConfig(c *cobra.Command) (*config.Config, error) {
	conf, err := node.LoadConfigFromFile()
	if err != nil {
		return nil, err
	}
	if err := cmd.EnsureCLIFlags(c, conf); err != nil {
		return nil, fmt.Errorf("mapping cli flags to config: %w", err)
	}
	return conf, nil
}

func shutdownServer(server *metrics.Server) {
	log.Info("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Stop(shutdownCtx); err != nil {
		log.With().Error("failed to stop server", log.Err(err))
	}
}

func setupClock() (*timesync.NodeClock, error) {
	appConfig, err := node.LoadConfigFromFile()
	if err != nil {
		return nil, fmt.Errorf("cannot load config file: %w", err)
	}
	types.SetLayersPerEpoch(appConfig.LayersPerEpoch)
	gTime, err := time.Parse(time.RFC3339, appConfig.Genesis.GenesisTime)
	if err != nil {
		return nil, fmt.Errorf("cannot parse genesis time %s: %w", appConfig.Genesis.GenesisTime, err)
	}
	clock, err := timesync.NewClock(
		timesync.WithLayerDuration(appConfig.LayerDuration),
		timesync.WithTickInterval(1*time.Second),
		timesync.WithGenesisTime(gTime),
		timesync.WithLogger(log.NewDefault("clock")),
	)
	if err != nil {
		return nil, err
	}
	return clock, nil
}
