package metrics

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	beacon "github.com/spacemeshos/go-spacemesh/beacon/metrics"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/timesync"
)

type Client struct {
	client v1.API

	clock *timesync.NodeClock
}

func NewClient(url string, clock *timesync.NodeClock) (*Client, error) {
	client, err := api.NewClient(api.Config{
		Address: url,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	api := v1.NewAPI(client)
	return &Client{
		client: api,
		clock:  clock,
	}, nil
}

func (c *Client) FetchBeaconValue(ctx context.Context, epoch int) (string, error) {
	lid := types.EpochID(epoch).FirstLayer()
	ts := c.clock.LayerToTime(lid)

	ts, err := time.Parse(time.RFC3339, "2023-07-11T18:50:00+02:00")
	if err != nil {
		return "", fmt.Errorf("failed to parse time: %w", err)
	}
	result, warnings, err := c.client.Query(ctx, fmt.Sprintf(`group by(beacon) (%s{epoch="517"})`, beacon.MetricNameCalculatedWeight()), ts)
	if err != nil {
		return "", fmt.Errorf("failed to fetch metric: %w", err)
	}
	if len(warnings) > 0 {
		log.Println("Query warnings:", warnings)
	}

	// Check if the result is a vector
	vector, ok := result.(model.Vector)
	if !ok {
		return "", fmt.Errorf("Query result is not a vector")
	}

	if len(vector) != 1 {
		return "", fmt.Errorf("nodes did not find consensus on a single beacon value")
	}

	// TODO(mafa): use result
	log.Println("Fetched metric value:", vector[0].Metric["beacon"])
	return string(vector[0].Metric["beacon"]), nil
}
