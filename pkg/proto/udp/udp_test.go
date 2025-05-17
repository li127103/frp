package udp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUdpPacket(t *testing.T) {
	assert := assert.New(t)

	buf := []byte("hello world")
	udpMsg := NewUDPPacket(buf, nil, nil)

	newBuf, err := GetContent(udpMsg)
	assert.NoError(err)
	assert.EqualValues(buf, newBuf)
}
