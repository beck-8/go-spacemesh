package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"

	beacon "github.com/spacemeshos/go-spacemesh/beacon/metrics"
	"github.com/spacemeshos/go-spacemesh/cmd/beaconalarm/metrics"
	spacemesh "github.com/spacemeshos/go-spacemesh/metrics"
)

const subsystem = "beacons"

func main() {
	// Parse command-line flags
	serverURL := flag.String("prometheus", "http://localhost:9090", "Prometheus server URL")
	flag.Parse()

	// Create a context for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Start the server
	alarmMetric := prometheus.BuildFQName(spacemesh.Namespace, subsystem, "beacon_alarm")
	server, err := metrics.NewServer(":8080", alarmMetric)
	if err != nil {
		log.Fatal(err)
	}
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}

	// Start the client
	client, err := metrics.NewClient(*serverURL)
	if err != nil {
		shutdownServer(server)
		log.Fatal(err)
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
				observedMetric := beacon.MetricNameCalculatedWeight()
				_, err := client.FetchBeaconValue(ctx, observedMetric)
				if err != nil {
					log.Println("Failed to fetch metric value", err)
				}
			}
		}
	})

	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}
}

func shutdownServer(server *metrics.Server) {
	log.Println("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Stop(shutdownCtx); err != nil {
		log.Fatal(err)
	}
}
