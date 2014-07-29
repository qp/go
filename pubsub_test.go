package qp_test

import (
	"testing"

	"github.com/qp/go"
	"github.com/stretchr/testify/require"
)

func TestPublisher(t *testing.T) {

	tp := &TestPubSubTransport{}
	testData := map[string]interface{}{"test": "data"}
	p := qp.NewPublisher("name", "instanceID", qp.JSON, tp)

	require.NoError(t, p.Publish("channel", testData))

	require.NotNil(t, tp.Published["channel"])
	var event qp.Event
	qp.JSON.Unmarshal(tp.Published["channel"], &event)
	require.NotNil(t, event)
	require.Equal(t, event.From, "name.instanceID")
	require.Equal(t, event.Data.(map[string]interface{})["test"], testData["test"])

}

func TestSubscriber(t *testing.T) {

	tp := &TestPubSubTransport{}

	s := qp.NewSubscriber(qp.JSON, tp)
	var events []*qp.Event
	s.SubscribeFunc("channel", func(e *qp.Event) {
		events = append(events, e)
	})

	event := &qp.Event{From: "place.id", Data: map[string]interface{}{"key": "value"}}
	message := &qp.Message{Source: "somewhere", Data: json(event)}
	require.NotNil(t, tp.Subscribed["channel"])
	tp.Subscribed["channel"].Handle(message)

	require.Equal(t, 1, len(events))
	require.Equal(t, "place.id", events[0].From)
	require.Equal(t, "value", events[0].Data.(map[string]interface{})["key"])

}
