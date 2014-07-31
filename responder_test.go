package qp_test

import (
	"testing"

	"github.com/qp/go"
	"github.com/stretchr/testify/require"
)

func TestHandlerFunc(t *testing.T) {

	var _ qp.RequestHandler = qp.RequestHandlerFunc(func(r *qp.Request) {})

}

func TestResponder(t *testing.T) {

	var testData = map[string]interface{}{"key": "value"}
	tp := &TestDirectTransport{}
	r1 := qp.NewResponder("function-one", "instance", qp.JSON, tp)
	r2 := qp.NewResponder("function-two", "instance", qp.JSON, tp)
	r3 := qp.NewResponder("function-three", "instance", qp.JSON, tp)

	require.NotNil(t, r1)
	require.NotNil(t, r2)
	require.NotNil(t, r3)

	var requests []*qp.Request
	require.NoError(t, r1.HandleFunc("one", func(r *qp.Request) {
		requests = append(requests, r)
		r.Data.(map[string]interface{})["one"] = true
	}))
	require.NoError(t, r2.HandleFunc("two", func(r *qp.Request) {
		requests = append(requests, r)
		r.Data.(map[string]interface{})["two"] = true
	}))
	require.NoError(t, r3.HandleFunc("three", func(r *qp.Request) {
		requests = append(requests, r)
		r.Data.(map[string]interface{})["three"] = true
	}))

	require.NotNil(t, tp.OnMessages["one"])
	require.NotNil(t, tp.OnMessages["two"])
	require.NotNil(t, tp.OnMessages["three"])

	// send fake response
	testRequest := &qp.Request{
		ID:   qp.RequestID(1),
		Data: testData,
		To:   []string{"two", "three"},
	}

	tp.OnMessages["one"].Handle(&qp.Message{Data: json(testRequest)})
	require.NotNil(t, tp.Sends["two"])
	require.Equal(t, len(requests), 1)

	tp.OnMessages["two"].Handle(&qp.Message{Data: tp.Sends["two"]})
	require.NotNil(t, tp.Sends["three"])
	require.Equal(t, len(requests), 2)

	tp.OnMessages["three"].Handle(&qp.Message{Data: tp.Sends["three"]})
	require.Equal(t, len(requests), 3)
	require.NotNil(t, tp.Sends["function-one.instance"])
	var finalRequest qp.Request

	require.NoError(t, qp.JSON.Unmarshal(tp.Sends["function-one.instance"], &finalRequest))

	require.Equal(t, qp.RequestID(1), finalRequest.ID)
	require.Equal(t, len(finalRequest.To), 0)
	require.Equal(t, len(finalRequest.From), 3)
	require.Equal(t, finalRequest.From[0], "function-one.instance")
	require.Equal(t, finalRequest.From[1], "function-two.instance")
	require.Equal(t, finalRequest.From[2], "function-three.instance")

	require.True(t, finalRequest.Data.(map[string]interface{})["one"].(bool))
	require.True(t, finalRequest.Data.(map[string]interface{})["two"].(bool))
	require.True(t, finalRequest.Data.(map[string]interface{})["three"].(bool))

}
