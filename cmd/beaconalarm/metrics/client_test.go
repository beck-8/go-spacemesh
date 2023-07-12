package metrics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Client(t *testing.T) {
	c, err := NewClient("http://localhost:9090")
	require.NoError(t, err)

	value, err := c.FetchBeaconValue(context.Background(), 11)
	require.NoError(t, err)

	require.Equal(t, "ca702962", value)
}
