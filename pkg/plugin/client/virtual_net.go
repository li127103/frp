//go:build !frps

package client

import (
	v1 "github.com/li127103/frp/pkg/config/v1"
	"io"
	"sync"
)

func init() {

}

type VirtualNetPlugin struct {
	pluginCtx PluginContext
	opts      *v1.VirtualNetPluginOptions
	mu        sync.Mutex
	conns     map[io.ReadWriteCloser]struct{}
}

func (p *VirtualNetPlugin) RemoveConn(conn io.ReadWriteCloser) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// Check if the map exists, as Close might have set it to nil concurrently
	if p.conns != nil {
		delete(p.conns, conn)
	}
}

func (p *VirtualNetPlugin) Name() string {
	return v1.PluginVirtualNet
}

func (p *VirtualNetPlugin) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Close any remaining connections
	for conn := range p.conns {
		_ = conn.Close()
	}
	p.conns = nil
	return nil
}
