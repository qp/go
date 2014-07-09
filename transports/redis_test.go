package transports

import (
	"encoding/json"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ensure Redis conforms to Transport interface
var _ Transport = (*Redis)(nil)

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
	a := assert.New(t)
	channel := "test"
	mc := make(chan *BinaryMessage)

	data, _ := json.Marshal(map[string]interface{}{"name": "Tyler"})

	r.ListenFor(channel)
	r.OnMessage(func(bm *BinaryMessage) {
		mc <- bm
	})
	r.Start()
	r.Send(channel, data)

	select {
	case bm := <-mc:
		a.Equal(channel, bm.Channel)
		a.Equal(data, bm.Data)
	case <-time.After(100 * time.Millisecond):
		a.Fail("No message received!")
	}

	r.Stop()

	// make sure we can start again after stopping
	r.ListenFor(channel)
	r.OnMessage(func(bm *BinaryMessage) {
		mc <- bm
	})

	r.Start()
	r.Send(channel, data)

	select {
	case bm := <-mc:
		a.Equal(channel, bm.Channel)
		a.Equal(data, bm.Data)
	case <-time.After(100 * time.Millisecond):
		a.Fail("No message received!")
	}

	r.Stop()

	stopRedis()

}
