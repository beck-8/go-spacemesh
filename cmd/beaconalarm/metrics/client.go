package metrics

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type Client struct {
	client v1.API
}

func NewClient(url string) (*Client, error) {
	client, err := api.NewClient(api.Config{
		Address: url,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	api := v1.NewAPI(client)
	return &Client{
		client: api,
	}, nil
}

func (c *Client) FetchMetricValue(ctx context.Context, metric string, epoch int) (float64, error) {
	result, warnings, err := c.client.QueryRange(ctx, fmt.Sprintf(`%s{epoch="%d"}`, metric, epoch), v1.Range{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
		Step:  1 * time.Minute,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to fetch metric: %w", err)
	}
	if len(warnings) > 0 {
		log.Println("Query warnings:", warnings)
	}

	// TODO(mafa): use result
	log.Println("Fetched metric value:", result)
	return 0, nil
}
