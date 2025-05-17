package proxy

import (
	v1 "github.com/li127103/frp/pkg/config/v1"
	"reflect"
)

func init() {
	RegisterProxyFactory(reflect.TypeOf(&v1.SUDPProxyConfig{}), NewSUDPProxy)
}

type SUDPProxy struct {
	*BaseProxy
	cfg *v1.SUDPProxyConfig
}

func NewSUDPProxy(baseProxy *BaseProxy) Proxy {
	unwrapped, ok := baseProxy.GetConfigurer().(*v1.SUDPProxyConfig)
	if !ok {
		return nil
	}
	return &SUDPProxy{
		BaseProxy: baseProxy,
		cfg:       unwrapped,
	}
}
func (pxy *SUDPProxy) Run() (remoteAddr string, err error) {

	xl := pxy.xl
	allowUsers := pxy.cfg.AllowUsers
	// if allowUsers is empty, only allow same user from proxy
	if len(allowUsers) == 0 {
		allowUsers = []string{pxy.GetUserInfo().User}
	}
	listener, errRet := pxy.rc.VisitorManager.Listen(pxy.GetName(), pxy.cfg.Secretkey, allowUsers)
	if errRet != nil {
		err = errRet
		return
	}
	pxy.listeners = append(pxy.listeners, listener)
	xl.Infof("sudp proxy custom listen success")

	pxy.startCommonTCPListenersHandler()
	return
}

func (pxy *SUDPProxy) Close() {
	pxy.BaseProxy.Close()
	pxy.rc.VisitorManager.CloseListener(pxy.GetName())
}
