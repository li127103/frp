package types

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

type Wrap struct {
	B   BandwidthQuantity `json:"b"`
	Int int               `json:"int"`
}

func TestBandwidthQuantity(t *testing.T) {
	require := require.New(t)

	var w Wrap
	err := json.Unmarshal([]byte(`{"b":"1KB","int":5}`), &w)
	require.NoError(err)
	require.EqualValues(1*KB, w.B.Bytes())

	buf, err := json.Marshal(&w)
	require.NoError(err)
	require.Equal(`{"b":"1KB","int":5}`, string(buf))
}

func TestPortsRangeSlice2String(t *testing.T) {
	require := require.New(t)

	ports := []PortsRange{
		{
			Start: 1000,
			End:   2000,
		},
		{
			Single: 3000,
		},
	}
	str := PortsRangeSlice(ports).String()
	require.Equal("1000-2000,3000", str)
}

func TestNewPortsRangeSliceFromString(t *testing.T) {
	require := require.New(t)

	ports, err := NewPortsRangeSliceFromString("1000-2000,3000")
	require.NoError(err)
	require.Equal([]PortsRange{
		{
			Start: 1000,
			End:   2000,
		},
		{
			Single: 3000,
		},
	}, ports)
}
