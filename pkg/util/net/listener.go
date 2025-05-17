package net

import (
	"fmt"
	"github.com/fatedier/golib/errors"
	"net"
	"sync"
)

// InternalListener is a listener that can be used to accept connections from
// other goroutines.
type InternalListener struct {
	acceptCh chan net.Conn
	closed   bool
	mu       sync.Mutex
}

func NewInternalListener() *InternalListener {
	return &InternalListener{
		acceptCh: make(chan net.Conn, 128),
	}
}

func (l *InternalListener) Accept() (net.Conn, error) {
	conn, ok := <-l.acceptCh
	if !ok {
		return nil, fmt.Errorf("listener closed")
	}
	return conn, nil
}

func (l *InternalListener) PutConn(conn net.Conn) error {
	err := errors.PanicToError(func() {
		select {
		case l.acceptCh <- conn:
		default:
			conn.Close()
		}
	})
	if err != nil {
		return fmt.Errorf("put conn error: listener is closed")
	}
	return nil
}

func (l *InternalListener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.closed {
		close(l.acceptCh)
		l.closed = true
	}
	return nil
}

func (l *InternalListener) Addr() net.Addr {
	return &InternalAddr{}
}

type InternalAddr struct{}

func (ia *InternalAddr) Network() string {
	return "internal"
}

func (ia *InternalAddr) String() string {
	return "internal"
}
