package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/spacemeshos/go-spacemesh/beacon"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/log"
	"github.com/spacemeshos/go-spacemesh/log/logtest"
	"github.com/spacemeshos/go-spacemesh/timesync"
)

func Test_Client(t *testing.T) {
	genesis, err := time.Parse(time.RFC3339, "2023-05-31T20:00:00.498Z")
	require.NoError(t, err)

	types.SetLayersPerEpoch(576)

	clock, err := timesync.NewClock(
		timesync.WithLayerDuration(5*time.Minute),
		timesync.WithTickInterval(1*time.Second),
		timesync.WithGenesisTime(genesis),
		timesync.WithLogger(log.NewDefault("clock")),
	)
	require.NoError(t, err)

	cfg := beacon.Config{
		ProposalDuration:         4 * time.Minute,
		FirstVotingRoundDuration: 30 * time.Minute,
		VotingRoundDuration:      4 * time.Minute,
		WeakCoinRoundDuration:    4 * time.Minute,
		RoundsNumber:             200,
	}

	c, err := NewClient("http://localhost:9090", cfg, logtest.New(t, zapcore.InfoLevel), clock)
	require.NoError(t, err)

	value, err := c.FetchBeaconValue(context.Background(), 20)
	require.NoError(t, err)

	require.Equal(t, "1b34ba7d", value)
}
