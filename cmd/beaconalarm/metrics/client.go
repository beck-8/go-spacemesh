package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	"github.com/spacemeshos/go-spacemesh/beacon"
	beaconMetrics "github.com/spacemeshos/go-spacemesh/beacon/metrics"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/log"
	"github.com/spacemeshos/go-spacemesh/timesync"
)

type Client struct {
	offset time.Duration

	client v1.API
	logger log.Logger
	clock  *timesync.NodeClock
}

func NewClient(url string, cfg beacon.Config, logger log.Logger, clock *timesync.NodeClock) (*Client, error) {
	client, err := api.NewClient(api.Config{
		Address: url,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	offset := cfg.ProposalDuration + cfg.FirstVotingRoundDuration + time.Duration(cfg.RoundsNumber-1)*(cfg.VotingRoundDuration+cfg.WeakCoinRoundDuration)
	logger.With().Info("using alarm offset",
		log.Duration("offset", offset),
		log.Duration("proposalDuration", cfg.ProposalDuration),
		log.Duration("firstVotingRoundDuration", cfg.FirstVotingRoundDuration),
		log.Duration("votingRoundDuration", cfg.VotingRoundDuration),
		log.Duration("weakCoinRoundDuration", cfg.WeakCoinRoundDuration),
		log.FieldNamed("roundsNumber", cfg.RoundsNumber),
	)

	return &Client{
		offset: offset,
		client: v1.NewAPI(client),
		logger: logger,
		clock:  clock,
	}, nil
}

func (c *Client) FetchBeaconValue(ctx context.Context, epoch int) (string, error) {
	lid := types.EpochID(epoch).FirstLayer()
	ts := c.clock.LayerToTime(lid)
	ts = ts.Add(c.offset)
	c.logger.With().Info("fetching beacon value", log.Int("epoch", epoch), log.Time("ts", ts))

	ts, err := time.Parse(time.RFC3339, "2023-07-11T23:10:00+00:00")
	if err != nil {
		return "", fmt.Errorf("failed to parse time: %w", err)
	}
	result, warnings, err := c.client.Query(ctx, fmt.Sprintf(`group by(beacon) (%s{kubernetes_namespace="testnet-05",epoch="%d"})`, beaconMetrics.MetricNameCalculatedWeight(), epoch+1), ts)
	if err != nil {
		return "", fmt.Errorf("failed to fetch metric: %w", err)
	}
	if len(warnings) > 0 {
		c.logger.With().Warning("query warnings:", log.Strings("warnings", warnings))
	}

	// Check if the result is a vector
	vector, ok := result.(model.Vector)
	if !ok {
		return "", fmt.Errorf("Query result is not a vector")
	}

	if len(vector) != 1 {
		return "", fmt.Errorf("nodes did not find consensus on a single beacon value")
	}

	beaconValue := string(vector[0].Metric["beacon"])
	log.With().Info("fetched beacon value", log.String("beacon", beaconValue))
	return beaconValue, nil
}
