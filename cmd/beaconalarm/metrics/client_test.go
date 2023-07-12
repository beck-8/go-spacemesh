package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/log"
	"github.com/spacemeshos/go-spacemesh/timesync"
)

func TestMain(m *testing.M) {
	types.SetLayersPerEpoch(10)

	m.Run()
}

func Test_Client(t *testing.T) {
	clock, err := timesync.NewClock(
		timesync.WithLayerDuration(1*time.Minute),
		timesync.WithTickInterval(1*time.Second),
		timesync.WithGenesisTime(time.Now()),
		timesync.WithLogger(log.NewDefault("clock")),
	)
	require.NoError(t, err)

	c, err := NewClient("http://localhost:9090", clock)
	require.NoError(t, err)

	value, err := c.FetchBeaconValue(context.Background(), 11)
	require.NoError(t, err)

	require.Equal(t, "1b34ba7d", value)
}
