package qp

import (
	"os/exec"
	"testing"
	"time"

	"github.com/qp/go/codecs"
	"github.com/qp/go/transports"
	"github.com/stretchr/testify/assert"
)

var tests = []struct {
	name      string
	pre       func() bool
	transport func() transports.Transport
	post      func()
}{
	{
		name: "InProc",
		pre:  func() bool { return true },
		transport: func() transports.Transport {
			return transports.MakeInProc(transports.MakeLog(true))
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
		transport: func() transports.Transport {
			return transports.MakeRedis("127.0.0.1:6379")
		},
		post: func() { exec.Command("redis-cli", "shutdown").Run() },
	},
}

func TestRequestMessenger(t *testing.T) {

	for _, test := range tests {

		if !test.pre() {
			t.Skip("Skipping because prefunc failed")
		}

		rm := MakeRequestMessenger("test", test.name, codecs.MakeJSON(), test.transport())
		rm2 := MakeRequestMessenger("test2", test.name, codecs.MakeJSON(), test.transport())
		if assert.NotNil(t, rm) && assert.NotNil(t, rm2) {
			rm.OnRequest(func(channel string, request *Request) {
				request.Data = "hello from handler"
			}, "test")

			rm.Start()
			rm2.Start()

			rf, err := rm2.Request("data", "test")
			if assert.NoError(t, err) {
				assert.Equal(t, rf.Response().Data.(string), "hello from handler")
				assert.Equal(t, rf.Response().From[0], "test2."+test.name)
			}

			rm.Stop()
			rm2.Stop()
		}

		test.post()
	}

}

func TestRequestMessengerMultipleJumps(t *testing.T) {

	for _, test := range tests {

		if !test.pre() {
			t.Skip("Skipping because prefunc failed")
		}

		rm := MakeRequestMessenger("multitest", test.name, codecs.MakeJSON(), test.transport())
		s1 := MakeRequestMessenger("one", test.name, codecs.MakeJSON(), test.transport())
		s2 := MakeRequestMessenger("two", test.name, codecs.MakeJSON(), test.transport())
		s3 := MakeRequestMessenger("three", test.name, codecs.MakeJSON(), test.transport())

		if assert.NotNil(t, rm) &&
			assert.NotNil(t, s1) &&
			assert.NotNil(t, s2) &&
			assert.NotNil(t, s3) {
			s1.OnRequest(func(channel string, request *Request) {
				request.Data = append(request.Data.([]interface{}), "one")
			}, "one")
			s2.OnRequest(func(channel string, request *Request) {
				request.Data = append(request.Data.([]interface{}), "two")
			}, "two")
			s3.OnRequest(func(channel string, request *Request) {
				request.Data = append(request.Data.([]interface{}), "three")
			}, "three")

			rm.Start()
			s1.Start()
			s2.Start()
			s3.Start()

			rf, err := rm.Request([]string{"origin"}, "one", "two", "three")
			if assert.NoError(t, err) {
				assert.Equal(t, rf.Response().Data.([]interface{})[0].(string), "origin")
				assert.Equal(t, rf.Response().Data.([]interface{})[1].(string), "one")
				assert.Equal(t, rf.Response().Data.([]interface{})[2].(string), "two")
				assert.Equal(t, rf.Response().Data.([]interface{})[3].(string), "three")
				assert.Equal(t, rf.Response().From[0], "multitest."+test.name)
				assert.Equal(t, rf.Response().From[1], "one."+test.name)
				assert.Equal(t, rf.Response().From[2], "two."+test.name)
				assert.Equal(t, rf.Response().From[3], "three."+test.name)
			}

			rm.Stop()
			s1.Stop()
			s2.Stop()
			s3.Stop()

			test.post()

		}
	}
}
