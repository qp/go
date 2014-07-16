package qp

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/qp/go/codecs"
	"github.com/qp/go/exchange"
	"github.com/qp/go/transports"
	"github.com/qp/go/transports/event"
	"github.com/stretchr/testify/assert"
)

var eventTests = []struct {
	name      string
	pre       func() bool
	transport func() transports.EventTransport
	post      func()
}{
	{
		name: "InProc",
		pre:  func() bool { return true },
		transport: func() transports.EventTransport {
			return event.MakeInProc(nil)
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
		transport: func() transports.EventTransport {
			return event.MakeRedis("127.0.0.1:6379")
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

		em := MakeEventMessenger("test", "one", codecs.MakeJSON(), test.transport())

		if assert.NotNil(t, em) {
			hit := make(chan struct{})
			data := map[string]interface{}{"name": "Tyler"}
			eh := func(channel string, event *exchange.Event) {
				assert.Equal(t, "publisher.one", event.From)
				assert.Equal(t, data, event.Data)
				hit <- struct{}{}
			}
			ehw := func(channel string, event *exchange.Event) {
				assert.Equal(t, "publisher.one", event.From)
				assert.Equal(t, data, event.Data)
				hit <- struct{}{}
			}

			em.Subscribe(eh, "test.event")
			em.SubscribeChildren(ehw, "test.event")

			em.Start()

			pub := MakeEventMessenger("publisher", "one", codecs.MakeJSON(), test.transport())
			pub.Start()
			pub.Publish(data, "test.event")
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
				case <-time.After(200 * time.Millisecond):
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
