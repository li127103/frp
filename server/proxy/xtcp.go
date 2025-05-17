package proxy

import (
	"fmt"
	v1 "github.com/li127103/frp/pkg/config/v1"
	"github.com/li127103/frp/pkg/msg"
	"reflect"
	"sync"
)

func init() {
	RegisterProxyFactory(reflect.TypeOf(&v1.XTCPProxyConfig{}), NewXTCPProxy)
}

type XTCPProxy struct {
	*BaseProxy
	cfg *v1.XTCPProxyConfig

	closeCh   chan struct{}
	closeOnce sync.Once
}

func NewXTCPProxy(baseProxy *BaseProxy) Proxy {
	unwrapped, ok := baseProxy.GetConfigurer().(*v1.XTCPProxyConfig)
	if !ok {
		return nil
	}
	return &XTCPProxy{
		BaseProxy: baseProxy,
		cfg:       unwrapped,
		closeCh:   make(chan struct{}),
	}
}

func (pxy *XTCPProxy) Run() (remoteAddr string, err error) {
	xl := pxy.xl

	if pxy.rc.NatHoleController == nil {
		err = fmt.Errorf("xtcp is not supported in frps")
		return
	}

	allowUsers := pxy.cfg.AllowUsers
	// if allowUsers is empty, only allow same user from proxy
	if len(allowUsers) == 0 {
		allowUsers = []string{pxy.GetUserInfo().User}
	}
	sidCh, err := pxy.rc.NatHoleController.ListenClient(pxy.GetName(), pxy.cfg.Secretkey, allowUsers)
	if err != nil {
		return "", err
	}
	go func() {
		for {
			select {
			case <-pxy.closeCh:
				return
			case sid := <-sidCh:
				workConn, errRet := pxy.GetWorkConnFromPool(nil, nil)
				if errRet != nil {
					continue
				}
				m := &msg.NatHoleSid{
					Sid: sid,
				}
				errRet = msg.WriteMsg(workConn, m)
				if errRet != nil {
					xl.Warnf("write nat hole sid package error, %v", errRet)
				}
				workConn.Close()
			}
		}
	}()
	return
}

func (pxy *XTCPProxy) Close() {
	pxy.closeOnce.Do(func() {
		pxy.BaseProxy.Close()
		pxy.rc.NatHoleController.CloseClient(pxy.GetName())
		close(pxy.closeCh)
	})
}
