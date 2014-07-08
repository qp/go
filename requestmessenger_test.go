package qp

import (
	"testing"

	"github.com/qp/go/codecs"
	"github.com/qp/go/transports"
	"github.com/stretchr/testify/assert"
)

func TestRequestMessengerInProc(t *testing.T) {

	rm := MakeRequestMessenger("test", "inproc", codecs.MakeJSON(), transports.MakeInProc())
	rm2 := MakeRequestMessenger("test2", "inproc2", codecs.MakeJSON(), transports.MakeInProc())
	if assert.NotNil(t, rm) && assert.NotNil(t, rm2) {
		rm.OnRequest(func(channel string, request *Request) {
			request.Data = "hello from handler"
		}, "test")

		rm.Start()

		rf, err := rm2.Request("data", "test")
		if assert.NoError(t, err) {
			assert.Equal(t, rf.Response().Data.(string), "hello from handler")
			assert.Equal(t, rf.Response().From[0], "test2.inproc2")
		}

		rm.Stop()
	}
}

func TestRequestMessengerInProcMultipleJumps(t *testing.T) {

	rm := MakeRequestMessenger("multitest", "inproc", codecs.MakeJSON(), transports.MakeInProc())
	s1 := MakeRequestMessenger("one", "inproc", codecs.MakeJSON(), transports.MakeInProc())
	s2 := MakeRequestMessenger("two", "inproc", codecs.MakeJSON(), transports.MakeInProc())
	s3 := MakeRequestMessenger("three", "inproc", codecs.MakeJSON(), transports.MakeInProc())

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
			assert.Equal(t, rf.Response().From[0], "multitest.inproc")
			assert.Equal(t, rf.Response().From[1], "one.inproc")
			assert.Equal(t, rf.Response().From[2], "two.inproc")
			assert.Equal(t, rf.Response().From[3], "three.inproc")
		}

		rm.Stop()
		s1.Stop()
		s2.Stop()
		s3.Stop()

	}
}
