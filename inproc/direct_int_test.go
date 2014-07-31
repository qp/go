package inproc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDirectStopping(t *testing.T) {

	d := NewDirect()
	require.NotNil(t, d)

	require.NotNil(t, directInstances[d])

	d.Stop(1 * time.Millisecond)
	select {
	case <-d.StopChan():
		require.Equal(t, len(directInstances), 0)
	case <-time.After(10 * time.Millisecond):
		require.FailNow(t, "transport did not stop")
	}

}
