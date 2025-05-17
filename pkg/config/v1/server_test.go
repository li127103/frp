package v1

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestServerConfigComplete(t *testing.T) {
	require := require.New(t)
	c := &ServerConfig{}
	c.Complete()

	require.EqualValues("token", c.Auth.Method)
	require.Equal(true, lo.FromPtr(c.Transport.TCPMux))
	require.Equal(true, lo.FromPtr(c.DetailedErrorsToClient))
}
