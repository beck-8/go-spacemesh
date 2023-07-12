package metrics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	beacon "github.com/spacemeshos/go-spacemesh/beacon/metrics"
)

func Test_Client(t *testing.T) {
	c, err := NewClient("http://localhost:9090")
	require.NoError(t, err)

	value, err := c.FetchBeaconValue(context.Background(), beacon.MetricNameCalculatedWeight())
	require.NoError(t, err)

	require.Equal(t, "ca702962", value)
}
