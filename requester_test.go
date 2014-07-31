package qp_test

import (
	"testing"
	"time"

	"github.com/qp/go"
	"github.com/stretchr/testify/require"
)

func TestRequester(t *testing.T) {

	var testData = map[string]interface{}{"key": "value"}
	var testChannels = []string{"one", "two", "three"}

	tp := &TestDirectTransport{}
	r := qp.NewRequester("name", "instance", qp.JSON, tp)
	require.NotNil(t, r)

	future, err := r.Issue(testChannels, testData)
	require.NoError(t, err)
	require.NotNil(t, future)

	bytes := tp.Sends["one"]
	if len(bytes) == 0 {
		require.FailNow(t, "Send was not called on Transport")
	}
	var req qp.Request
	qp.JSON.Unmarshal(bytes, &req)
	require.NotEmpty(t, req.ID)
	require.Equal(t, len(req.From), 1)
	require.Equal(t, req.From[0], "name.instance")
	require.Equal(t, len(req.To), 2)
	require.Equal(t, req.To[0], "two")
	require.Equal(t, req.To[1], "three")
	require.Equal(t, req.Data, testData)

	// send fake response
	testResponse := &qp.Response{
		ID: req.ID,
	}
	responseMsg := &qp.Message{Source: "", Data: json(testResponse)}
	tp.OnMessages["name.instance"].Handle(responseMsg)

	response := future.Response(1 * time.Second)
	require.Equal(t, testResponse, response)

}

func TestRequesterResponseTimeout(t *testing.T) {

	var testData = map[string]interface{}{"key": "value"}
	var testChannels = []string{"one", "two", "three"}

	tp := &TestDirectTransport{}
	r := qp.NewRequester("name", "instance", qp.JSON, tp)
	require.NotNil(t, r)

	future, err := r.Issue(testChannels, testData)
	require.NoError(t, err)
	require.NotNil(t, future)

	bytes := tp.Sends["one"]
	if len(bytes) == 0 {
		require.FailNow(t, "Send was not called on Transport")
	}
	var req qp.Request
	qp.JSON.Unmarshal(bytes, &req)
	require.NotEmpty(t, req.ID)
	require.Equal(t, len(req.From), 1)
	require.Equal(t, req.From[0], "name.instance")
	require.Equal(t, len(req.To), 2)
	require.Equal(t, req.To[0], "two")
	require.Equal(t, req.To[1], "three")
	require.Equal(t, req.Data, testData)

	// do not send response - force timeout
	response := future.Response(1 * time.Millisecond)

	require.Nil(t, response)
}
