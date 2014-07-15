package event

import (
	"os/exec"
	"testing"
	"time"

	"github.com/qp/go/transports"
	"github.com/stretchr/testify/assert"
)

// ensure Redis conforms to Transport interface
var _ transports.EventTransport = (*Redis)(nil)

// initRedis does the following:
// 1. Checks if Redis is installed
// 2. Checks if Redis is running
// 3. Starts it if it is not
func initRedis() bool {
	err := exec.Command("which", "redis-cli").Run()
	if err != nil {
		return false
	}
	err = exec.Command("redis-cli", "ping").Run()
	if err != nil {
		// Redis is not running. Run it.
		err = exec.Command("redis-server", "--daemonize", "yes").Run()
		if err != nil {
			return false
		}
		time.Sleep(200 * time.Millisecond)
	}
	return true
}

func stopRedis() {
	exec.Command("redis-cli", "shutdown").Run()
}

func TestRedis(t *testing.T) {

	if !initRedis() {
		t.Skip("Cannot start redis - skipping.")
	}

	r := MakeRedis("127.0.0.1:6379")
	data := []byte(`testing`)
	hit := make(chan struct{})

	r.ListenFor("test.event")
	r.ListenForChildren("test.event")
	r.OnMessage(func(bm *transports.BinaryMessage) {
		assert.Equal(t, "test.event", bm.Channel)
		assert.Equal(t, data, bm.Data)
		hit <- struct{}{}
	})

	r.Start()

	time.Sleep(10 * time.Millisecond)

	publisher := MakeRedis("127.0.0.1:6379")
	publisher.Start()
	publisher.Publish("test.event", data)
	publisher.Stop()

	count := 0
loop:
	for {
		select {
		case <-hit:
			count++
			if count == 2 {
				break loop
			}
		case <-time.After(100 * time.Millisecond):
			assert.Fail(t, "Timed out while waiting for message")
			break loop
		}
	}

	assert.Equal(t, 2, count)

	r.Stop()

	stopRedis()

}
