package vnet

import (
	"context"
	"github.com/fatedier/golib/pool"
	"golang.zx2c4.com/wireguard/tun"
	"io"
)

const (
	offset            = 16
	defaultPacketSize = 1420
)

type TunDevice interface {
	io.ReadWriteCloser
}

func OpenTun(ctx context.Context, addr string) (TunDevice, error) {
	td, err := openTun(ctx, addr)
	if err != nil {
		return nil, err
	}
	mtu, err := td.MTU()
	if err != nil {
		mtu = defaultPacketSize
	}
	bufferSize := max(mtu, defaultPacketSize)
	batchSize := td.BatchSize()
	device := &tunDeviceWrapper{
		dev:         td,
		bufferSize:  bufferSize,
		readBuffers: make([][]byte, batchSize),
		sizeBuffer:  make([]int, batchSize),
	}

	for i := range device.readBuffers {
		device.readBuffers[i] = make([]byte, offset+bufferSize)
	}

	return device, nil
}

type tunDeviceWrapper struct {
	dev           tun.Device
	bufferSize    int
	readBuffers   [][]byte
	packetBuffers [][]byte
	sizeBuffer    []int
}

func (d *tunDeviceWrapper) Read(p []byte) (int, error) {
	if len(d.packetBuffers) > 0 {
		n := copy(p, d.packetBuffers[0])
		d.packetBuffers = d.packetBuffers[1:]
		return n, nil
	}
	n, err := d.dev.Read(d.readBuffers, d.sizeBuffer, offset)

	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, io.EOF
	}

	for i := range n {
		if d.sizeBuffer[i] <= 0 {
			continue
		}
		d.packetBuffers = append(d.packetBuffers, d.readBuffers[i][offset:offset+d.sizeBuffer[i]])
	}
	dataSize := copy(p, d.packetBuffers[0])
	d.packetBuffers = d.packetBuffers[1:]

	return dataSize, nil
}
func (d *tunDeviceWrapper) Write(p []byte) (int, error) {
	buf := pool.GetBuf(offset + d.bufferSize)
	defer pool.PutBuf(buf)

	n := copy(buf[offset:], p)
	_, err := d.dev.Write([][]byte{buf[:offset+n]}, offset)
	return n, err
}

func (d *tunDeviceWrapper) Close() error {
	return d.dev.Close()
}
