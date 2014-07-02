package transports

import (
	"encoding/json"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ensure Redis conforms to Transport interface
var conformsRedis = Transport(&Redis{})

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

	redis := NewRedis("127.0.0.1:6379")
	a := assert.New(t)
	topic := "test"
	mc := make(chan *BinaryMessage)

	data, _ := json.Marshal(map[string]interface{}{"name": "Tyler"})

	err := redis.ListenFor(topic, func(bm *BinaryMessage) {
		mc <- bm
	})
	redis.Start()
	err = redis.Send(topic, data)
	if a.NoError(err) {
		select {
		case bm := <-mc:
			a.Equal(topic, bm.Topic)
			a.Equal(data, bm.Data)
		case <-time.After(100 * time.Millisecond):
			a.Fail("No message received!")
		}
	}
	redis.Stop()

	stopRedis()
}
