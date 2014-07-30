package qp_test

import (
	"testing"

	"github.com/stretchr/pat/stop"

	"github.com/qp/go"
	"github.com/qp/go/inproc"
	"github.com/stretchr/testify/require"
)

func TestServiceHandler(t *testing.T) {
	d := inproc.NewDirect()
	qp.Service("name", "instance", qp.JSON, d,
		qp.RequestHandlerFunc(func(r *qp.Request) {
			r.Data = "hit"
		}))

	defer func() {
		d.Stop(stop.NoWait)
		<-d.StopChan()
	}()

	d.Start()

	requester := qp.NewRequester("requester", "one", qp.JSON, d)
	f, err := requester.Issue([]string{"name"}, "test")
	require.NoError(t, err)
	require.Equal(t, "hit", f.Response().Data)
	require.Equal(t, "requester.one", f.Response().From[0])
	require.Equal(t, "name.instance", f.Response().From[1])
}

func TestServiceMultiple(t *testing.T) {
	d := inproc.NewDirect()
	qp.Service("name", "instance", qp.JSON, d,
		qp.RequestHandlerFunc(func(r *qp.Request) {
			r.Data = append(r.Data.([]interface{}), "first")
		}))
	qp.Service("name2", "instance", qp.JSON, d,
		qp.RequestHandlerFunc(func(r *qp.Request) {
			r.Data = append(r.Data.([]interface{}), "second")
		}))
	qp.Service("name3", "instance", qp.JSON, d,
		qp.RequestHandlerFunc(func(r *qp.Request) {
			r.Data = append(r.Data.([]interface{}), "third")
		}))

	defer func() {
		d.Stop(stop.NoWait)
		<-d.StopChan()
	}()

	d.Start()

	requester := qp.NewRequester("requester", "one", qp.JSON, d)
	f, err := requester.Issue([]string{"name", "name2", "name3"}, []string{"origin"})
	require.NoError(t, err)
	require.Equal(t, "origin", f.Response().Data.([]interface{})[0])
	require.Equal(t, "first", f.Response().Data.([]interface{})[1])
	require.Equal(t, "second", f.Response().Data.([]interface{})[2])
	require.Equal(t, "third", f.Response().Data.([]interface{})[3])
	require.Equal(t, "requester.one", f.Response().From[0])
	require.Equal(t, "name.instance", f.Response().From[1])
	require.Equal(t, "name2.instance", f.Response().From[2])
	require.Equal(t, "name3.instance", f.Response().From[3])
}
