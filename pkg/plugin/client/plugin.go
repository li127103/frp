package client

import (
	"context"
	"fmt"
	"github.com/fatedier/golib/errors"
	v1 "github.com/li127103/frp/pkg/config/v1"
	"github.com/li127103/frp/pkg/vnet"
	pp "github.com/pires/go-proxyproto"
	"io"
	"net"
	"sync"
)

type PluginContext struct {
	Name           string
	VnetController *vnet.Controller
}

// Creators is used for create plugins to handle connections.
var creators = make(map[string]CreatorFn)

type CreatorFn func(pluginCtx PluginContext, options v1.ClientPluginOptions) (Plugin, error)

func Register(name string, fn CreatorFn) {
	if _, exist := creators[name]; exist {
		panic(fmt.Sprintf("plugin [%s] is already registered", name))
	}
	creators[name] = fn
}
func Create(pluginName string, pluginCtx PluginContext, options v1.ClientPluginOptions) (p Plugin, err error) {
	if fn, ok := creators[pluginName]; ok {
		p, err = fn(pluginCtx, options)
	} else {
		err = fmt.Errorf("plugin [%s] is not registered", pluginName)
	}
	return
}

type ConnectionInfo struct {
	Conn           io.ReadWriteCloser
	UnderlyingConn net.Conn

	ProxyProtocolHeader *pp.Header
	SrcAddr             net.Addr
	DstAddr             net.Addr
}
type Plugin interface {
	Name() string

	Handle(ctx context.Context, connInfo *ConnectionInfo)
	Close() error
}

type Listener struct {
	conns  chan net.Conn
	closed bool
	mu     sync.Mutex
}

func NewProxyListener() *Listener {
	return &Listener{
		conns: make(chan net.Conn, 64),
	}
}

func (l *Listener) Accept() (net.Conn, error) {
	conn, ok := <-l.conns
	if !ok {
		return nil, fmt.Errorf("listener closed")
	}
	return conn, nil
}

func (l *Listener) PutConn(conn net.Conn) error {
	err := errors.PanicToError(func() {
		l.conns <- conn
	})
	return err
}

func (l *Listener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.closed {
		close(l.conns)
		l.closed = true
	}
	return nil
}

func (l *Listener) Addr() net.Addr {
	return (*net.TCPAddr)(nil)
}
