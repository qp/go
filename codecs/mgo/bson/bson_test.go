package bson_test

import (
	"testing"

	"github.com/qp/go/codecs/mgo/bson"
	"github.com/stretchr/testify/require"
	mgobson "labix.org/v2/mgo/bson"
)

func TestMarshal(t *testing.T) {

	data := map[string]interface{}{"hello": "world"}

	b, err := bson.Codec.Marshal(data)
	require.NoError(t, err)
	require.Equal(t, string(b), "\x16\x00\x00\x00\x02hello\x00\x06\x00\x00\x00world\x00\x00")

}

func TestUnmarshal(t *testing.T) {

	var data interface{}
	bson.Codec.Unmarshal([]byte("\x16\x00\x00\x00\x02hello\x00\x06\x00\x00\x00world\x00\x00"), &data)

	require.NotNil(t, data)
	require.Equal(t, "world", data.(mgobson.M)["hello"])

}
