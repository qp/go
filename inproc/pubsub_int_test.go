package inproc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPubSubStopping(t *testing.T) {

	ps := NewPubSub()
	require.NotNil(t, ps)

	require.NotNil(t, instances[ps])

	ps.Stop(1 * time.Millisecond)
	select {
	case <-ps.StopChan():
		require.Equal(t, len(instances), 0)
	case <-time.After(10 * time.Millisecond):
		require.FailNow(t, "transport did not stop")
	}

}
