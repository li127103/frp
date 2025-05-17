package visitor

import (
	"fmt"
	libio "github.com/fatedier/golib/io"
	netpkg "github.com/li127103/frp/pkg/util/net"
	"github.com/li127103/frp/pkg/util/util"
	"io"
	"net"
	"slices"
	"sync"
)

type listenerBundle struct {
	l          *netpkg.InternalListener
	sk         string
	allowUsers []string
}

// Manager for visitor listeners.
type Manager struct {
	listeners map[string]*listenerBundle

	mu sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		listeners: make(map[string]*listenerBundle),
	}
}

func (vm *Manager) Listen(name string, sk string, allowUsers []string) (*netpkg.InternalListener, error) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if _, ok := vm.listeners[name]; ok {
		return nil, fmt.Errorf("custom listener for [%s] is repeated", name)
	}

	l := netpkg.NewInternalListener()
	vm.listeners[name] = &listenerBundle{
		l:          l,
		sk:         sk,
		allowUsers: allowUsers,
	}
	return l, nil
}

func (vm *Manager) NewConn(name string, conn net.Conn, timestamp int64, signKey string,
	useEncryption bool, useCompression bool, visitorUser string) (err error) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if l, ok := vm.listeners[name]; ok {
		if util.GetAuthKey(l.sk, timestamp) != signKey {
			err = fmt.Errorf("visitor connection of [%s] auth failed", name)
			return
		}
		if !slices.Contains(l.allowUsers, visitorUser) && !slices.Contains(l.allowUsers, "*") {
			err = fmt.Errorf("visitor connection of [%s] user [%s] not allowed", name, visitorUser)
			return
		}
		var rwc io.ReadWriteCloser = conn
		if useEncryption {
			if rwc, err = libio.WithEncryption(rwc, []byte(l.sk)); err != nil {
				err = fmt.Errorf("create encryption connection failed: %v", err)
				return
			}
		}
		if useCompression {
			rwc = libio.WithCompression(rwc)
		}
		err = l.l.PutConn(netpkg.WrapReadWriteCloserToConn(rwc, conn))
	} else {
		err = fmt.Errorf("custom listener for [%s] doesn't exist", name)
		return
	}
	return
}
func (vm *Manager) CloseListener(name string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	delete(vm.listeners, name)
}
