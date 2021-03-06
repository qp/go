package inproc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPubSubStopping(t *testing.T) {

	ps := NewPubSub()
	ps.Start()
	require.NotNil(t, ps)

	require.NotNil(t, pubSubInstances[ps])

	ps.Stop(1 * time.Millisecond)
	select {
	case <-ps.StopChan():
		require.Equal(t, len(pubSubInstances), 0)
	case <-time.After(10 * time.Millisecond):
		require.FailNow(t, "transport did not stop")
	}

}
