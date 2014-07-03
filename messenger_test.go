package qp

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/qp/go/codecs"
	"github.com/qp/go/messages"
	"github.com/qp/go/transports"
	"github.com/stretchr/testify/assert"
)

func TestMessenger(t *testing.T) {
	a := assert.New(t)

	m1 := NewMessenger("first", &codecs.JSON{}, transports.NewInProc())

	m2 := NewMessenger("second", &codecs.JSON{}, transports.NewInProc())
	m2.OnRequest = func(message *messages.Message) interface{} {
		if a.NotNil(message) {
			message.Data.(map[string]interface{})["second"] = true
			return nil
		}
		return errors.New("message was empty")
	}

	m3 := NewMessenger("third", &codecs.JSON{}, transports.NewInProc())
	m3.OnRequest = func(message *messages.Message) interface{} {
		if a.NotNil(message) {
			message.Data.(map[string]interface{})["third"] = true
			return nil
		}
		return errors.New("message was empty")
	}

	r, err := m1.Request(map[string]interface{}{"first": true}, "second", "third")

	if a.NotNil(r) && a.NoError(err) {
		msg := r.Message()
		if a.False(msg.HasError()) {
			data := msg.Data.(map[string]interface{})
			a.True(data["first"].(bool))
			a.True(data["second"].(bool))
			a.True(data["third"].(bool))
			a.True(strings.Contains(msg.From[0], "first/"), msg.From[0])
			a.True(strings.Contains(msg.From[1], "second/"), msg.From[1])
			a.True(strings.Contains(msg.From[2], "third/"), msg.From[2])
		}
	}

	m1.Stop()
	m2.Stop()
	m3.Stop()
}

func TestMessengerError(t *testing.T) {
	a := assert.New(t)

	m1 := NewMessenger("first", &codecs.JSON{}, transports.NewInProc())

	m2 := NewMessenger("second", &codecs.JSON{}, transports.NewInProc())
	m2.OnRequest = func(message *messages.Message) interface{} {
		if a.NotNil(message) {
			message.Data.(map[string]interface{})["second"] = true
			return nil
		}
		return errors.New("message was empty")
	}

	m3 := NewMessenger("third", &codecs.JSON{}, transports.NewInProc())
	m3.OnRequest = func(message *messages.Message) interface{} {
		// an error can be any object
		return map[string]interface{}{"code": 123, "message": "deliberate failure"}
	}

	r, err := m1.Request(map[string]interface{}{"first": true}, "second", "third")

	if a.NotNil(r) && a.NoError(err) {
		msg := r.Message()
		if a.True(msg.HasError()) {
			err := msg.Err.(map[string]interface{})
			a.Equal(err["code"], 123)
			a.Equal(err["message"], "deliberate failure")
			data := msg.Data.(map[string]interface{})
			a.True(data["first"].(bool))
			a.True(data["second"].(bool))
			a.True(strings.Contains(msg.From[0], "first/"), msg.From[0])
			a.True(strings.Contains(msg.From[1], "second/"), msg.From[1])
			a.True(strings.Contains(msg.From[2], "third/"), msg.From[2])
		}
	}

	m1.Stop()
	m2.Stop()
	m3.Stop()
}

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

func TestMessengerRedis(t *testing.T) {
	initRedis()
	a := assert.New(t)

	m1 := NewMessenger("first", &codecs.JSON{}, transports.NewInProc())

	m2 := NewMessenger("second", &codecs.JSON{}, transports.NewInProc())
	m2.OnRequest = func(message *messages.Message) interface{} {
		if a.NotNil(message) {
			message.Data.(map[string]interface{})["second"] = true
			return nil
		}
		return errors.New("message was empty")
	}

	m3 := NewMessenger("third", &codecs.JSON{}, transports.NewInProc())
	m3.OnRequest = func(message *messages.Message) interface{} {
		if a.NotNil(message) {
			message.Data.(map[string]interface{})["third"] = true
			return nil
		}
		return errors.New("message was empty")
	}

	r, err := m1.Request(map[string]interface{}{"first": true}, "second", "third")

	if a.NotNil(r) && a.NoError(err) {
		msg := r.Message()
		if a.False(msg.HasError()) {
			data := msg.Data.(map[string]interface{})
			a.True(data["first"].(bool))
			a.True(data["second"].(bool))
			a.True(data["third"].(bool))
			a.True(strings.Contains(msg.From[0], "first/"), msg.From[0])
			a.True(strings.Contains(msg.From[1], "second/"), msg.From[1])
			a.True(strings.Contains(msg.From[2], "third/"), msg.From[2])
		}
	}

	m1.Stop()
	m2.Stop()
	m3.Stop()
	stopRedis()
}

func TestMessengerErrorRedis(t *testing.T) {
	initRedis()
	a := assert.New(t)

	m1 := NewMessenger("first", &codecs.JSON{}, transports.NewInProc())

	m2 := NewMessenger("second", &codecs.JSON{}, transports.NewInProc())
	m2.OnRequest = func(message *messages.Message) interface{} {
		if a.NotNil(message) {
			message.Data.(map[string]interface{})["second"] = true
			return nil
		}
		return errors.New("message was empty")
	}

	m3 := NewMessenger("third", &codecs.JSON{}, transports.NewInProc())
	m3.OnRequest = func(message *messages.Message) interface{} {
		// an error can be any object
		return map[string]interface{}{"code": 123, "message": "deliberate failure"}
	}

	r, err := m1.Request(map[string]interface{}{"first": true}, "second", "third")

	if a.NotNil(r) && a.NoError(err) {
		msg := r.Message()
		if a.True(msg.HasError()) {
			err := msg.Err.(map[string]interface{})
			a.Equal(err["code"], 123)
			a.Equal(err["message"], "deliberate failure")
			data := msg.Data.(map[string]interface{})
			a.True(data["first"].(bool))
			a.True(data["second"].(bool))
			a.True(strings.Contains(msg.From[0], "first/"), msg.From[0])
			a.True(strings.Contains(msg.From[1], "second/"), msg.From[1])
			a.True(strings.Contains(msg.From[2], "third/"), msg.From[2])
		}
	}

	m1.Stop()
	m2.Stop()
	m3.Stop()
	stopRedis()
}
