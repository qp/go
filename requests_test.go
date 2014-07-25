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

var requestTests = []struct {
	name      string
	pre       func() bool
	transport func() qp.RequestTransport
	post      func()
}{
	{
		name: "InProc",
		pre:  func() bool { return true },
		transport: func() qp.RequestTransport {
			return inproc.NewReqTransport(nil)
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
		transport: func() qp.RequestTransport {
			return redis.NewReqTransport("127.0.0.1:6379")
		},
		post: func() { exec.Command("redis-cli", "shutdown").Run() },
	},
}

func TestRequestMessenger(t *testing.T) {

	for _, test := range requestTests {

		if !test.pre() {
			fmt.Println("Skipping", test.name, "due to pre-func fail.")
			continue
		}

		rm := qp.NewRequester("test", test.name, qp.JSON, test.transport())
		rm2 := qp.NewRequester("test2", test.name, qp.JSON, test.transport())
		if assert.NotNil(t, rm) && assert.NotNil(t, rm2) {
			rm.OnRequest(func(channel string, Request *qp.Request) {
				Request.Data = "hello from handler"
			}, []string{"test"})

			rm.Start()
			rm2.Start()

			rf, err := rm2.Request("data", []string{"test"})
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

	for _, test := range requestTests {

		if !test.pre() {
			fmt.Println("Skipping", test.name, "due to pre-func fail.")
			continue
		}

		rm := qp.NewRequester("multitest", test.name, qp.JSON, test.transport())
		s1 := qp.NewRequester("one", test.name, qp.JSON, test.transport())
		s2 := qp.NewRequester("two", test.name, qp.JSON, test.transport())
		s3 := qp.NewRequester("three", test.name, qp.JSON, test.transport())

		if assert.NotNil(t, rm) &&
			assert.NotNil(t, s1) &&
			assert.NotNil(t, s2) &&
			assert.NotNil(t, s3) {
			s1.OnRequest(func(channel string, Request *qp.Request) {
				Request.Data = append(Request.Data.([]interface{}), "one")
			}, []string{"one"})
			s2.OnRequest(func(channel string, Request *qp.Request) {
				Request.Data = append(Request.Data.([]interface{}), "two")
			}, []string{"two"})
			s3.OnRequest(func(channel string, Request *qp.Request) {
				Request.Data = append(Request.Data.([]interface{}), "three")
			}, []string{"three"})

			rm.Start()
			s1.Start()
			s2.Start()
			s3.Start()

			rf, err := rm.Request([]string{"origin"}, []string{"one", "two", "three"})
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
