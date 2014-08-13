package qp_test

import (
	"testing"
	"time"

	"github.com/stretchr/pat/stop"

	"github.com/qp/go"
	"github.com/qp/go/inproc"
	"github.com/stretchr/testify/require"
)

func TestServiceHandler(t *testing.T) {
	d := inproc.NewDirect()
	qp.Service("name", "instance", qp.JSON, d,
		qp.TransactionHandlerFunc(func(r *qp.Transaction) *qp.Transaction {
			r.Data = "hit"
			return r
		}),
	)

	defer func() {
		d.Stop(stop.NoWait)
		<-d.StopChan()
	}()

	d.Start()

	requester, err := qp.NewRequester("requester", "one", qp.JSON, d)
	require.NoError(t, err)
	f, err := requester.Issue([]string{"name"}, "test")
	require.NoError(t, err)
	res, _ := f.Response(1 * time.Second)
	require.Equal(t, "hit", res.Data)
	require.Equal(t, "requester.one", res.From[0])
	require.Equal(t, "name.instance", res.From[1])
}

func TestServiceMultiple(t *testing.T) {
	d := inproc.NewDirect()
	qp.Service("name", "instance", qp.JSON, d,
		qp.TransactionHandlerFunc(func(r *qp.Transaction) *qp.Transaction {
			r.Data = append(r.Data.([]interface{}), "first")
			return r
		}))
	qp.Service("name2", "instance", qp.JSON, d,
		qp.TransactionHandlerFunc(func(r *qp.Transaction) *qp.Transaction {
			r.Data = append(r.Data.([]interface{}), "second")
			return r
		}))
	qp.Service("name3", "instance", qp.JSON, d,
		qp.TransactionHandlerFunc(func(r *qp.Transaction) *qp.Transaction {
			r.Data = append(r.Data.([]interface{}), "third")
			return r
		}))

	defer func() {
		d.Stop(stop.NoWait)
		<-d.StopChan()
	}()

	d.Start()

	requester, err := qp.NewRequester("requester", "one", qp.JSON, d)
	require.NoError(t, err)
	f, err := requester.Issue([]string{"name", "name2", "name3"}, []string{"origin"})
	require.NoError(t, err)
	r, err := f.Response(1 * time.Second)

	require.NoError(t, err)
	require.Equal(t, "origin", r.Data.([]interface{})[0])
	require.Equal(t, "first", r.Data.([]interface{})[1])
	require.Equal(t, "second", r.Data.([]interface{})[2])
	require.Equal(t, "third", r.Data.([]interface{})[3])
	require.Equal(t, "requester.one", r.From[0])
	require.Equal(t, "name.instance", r.From[1])
	require.Equal(t, "name2.instance", r.From[2])
	require.Equal(t, "name3.instance", r.From[3])

}
