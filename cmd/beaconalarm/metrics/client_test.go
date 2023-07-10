package metrics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	beacon "github.com/spacemeshos/go-spacemesh/beacon/metrics"
)

func Test_Client(t *testing.T) {
	// Create an HTTP client with OAuth2 authentication using a custom RoundTripper
	c, err := NewClient("http://localhost:9090")
	require.NoError(t, err)

	value, err := c.FetchMetricValue(context.Background(), beacon.MetricNameCalculatedWeight(), 11)
	require.NoError(t, err)

	require.Equal(t, 0.0, value)
}
