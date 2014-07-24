package qp_test

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/qp/go"
	"github.com/qp/go/inproc"
	"github.com/qp/go/redis"
	"github.com/stretchr/testify/assert"
)

var eventTests = []struct {
	name      string
	pre       func() bool
	transport func() qp.EventTransport
	post      func()
}{
	{
		name: "InProc",
		pre:  func() bool { return true },
		transport: func() qp.EventTransport {
			return inproc.NewPubSubTransport(nil)
		},
		post: func() {},
	},
	{
		name: "Redis",
		pre: func() bool {
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
		},
		transport: func() qp.EventTransport {
			return redis.NewPubSubTransport("127.0.0.1:6379")
		},
		post: func() { exec.Command("redis-cli", "shutdown").Run() },
	},
}

func TestEventMessenger(t *testing.T) {

	for _, test := range eventTests {

		if !test.pre() {
			fmt.Println("Skipping", test.name, "due to pre-func fail.")
			continue
		}

		em := qp.NewPubSub("test", "one", qp.JSON, test.transport())

		if assert.NotNil(t, em) {
			hit := make(chan struct{})
			data := map[string]interface{}{"name": "Tyler"}
			eh := func(channel string, event *qp.Event) {
				assert.Equal(t, "publisher.one", event.From)
				assert.Equal(t, data, event.Data)
				hit <- struct{}{}
			}
			ehw := func(channel string, event *qp.Event) {
				assert.Equal(t, "publisher.one", event.From)
				assert.Equal(t, data, event.Data)
				hit <- struct{}{}
			}

			em.Subscribe(eh, "test.event.one")
			em.SubscribeChildren(ehw, "test.event")

			em.Start()

			pub := qp.NewPubSub("publisher", "one", qp.JSON, test.transport())
			pub.Start()
			pub.Send(data, "test.event.one")
			pub.Stop()

			count := 0
		loop:
			for {
				select {
				case <-hit:
					count++
					if count == 2 {
						break loop
					}
				case <-time.After(500 * time.Millisecond):
					assert.Fail(t, "Timed out while waiting for events")
					break loop
				}
			}
			assert.Equal(t, 2, count)

			em.Stop()

		}

		test.post()
	}

}
