package proxy

import (
	v1 "github.com/li127103/frp/pkg/config/v1"
	"github.com/li127103/frp/pkg/util/util"
	"github.com/li127103/frp/pkg/util/vhost"
	"reflect"
	"strings"
)

func init() {
	RegisterProxyFactory(reflect.TypeOf(&v1.HTTPSProxyConfig{}), NewHTTPSProxy)
}

type HTTPSProxy struct {
	*BaseProxy
	cfg *v1.HTTPSProxyConfig
}

func NewHTTPSProxy(baseProxy *BaseProxy) Proxy {
	unwrapped, ok := baseProxy.GetConfigurer().(*v1.HTTPSProxyConfig)
	if !ok {
		return nil
	}
	return &HTTPSProxy{
		BaseProxy: baseProxy,
		cfg:       unwrapped,
	}
}

func (pxy *HTTPSProxy) Run() (remoteAddr string, err error) {
	xl := pxy.xl
	routeConfig := &vhost.RouteConfig{}

	defer func() {
		if err != nil {
			pxy.Close()
		}
	}()
	addrs := make([]string, 0)

	for _, domain := range pxy.cfg.CustomDomains {
		if domain == "" {
			continue
		}
		routeConfig.Domain = domain
		l, errRet := pxy.rc.VhostHTTPSMuxer.Listen(pxy.ctx, routeConfig)
		if errRet != nil {
			err = errRet
			return
		}
		xl.Infof("https proxy listen for host [%s]", routeConfig.Domain)
		pxy.listeners = append(pxy.listeners, l)
		addrs = append(addrs, util.CanonicalAddr(routeConfig.Domain, pxy.serverCfg.VhostHTTPSPort))
	}

	if pxy.cfg.SubDomain != "" {
		routeConfig.Domain = pxy.cfg.SubDomain + "." + pxy.serverCfg.SubDomainHost
		l, errRet := pxy.rc.VhostHTTPSMuxer.Listen(pxy.ctx, routeConfig)
		if errRet != nil {
			err = errRet
			return
		}
		xl.Infof("https proxy listen for host [%s]", routeConfig.Domain)
		pxy.listeners = append(pxy.listeners, l)
		addrs = append(addrs, util.CanonicalAddr(routeConfig.Domain, pxy.serverCfg.VhostHTTPSPort))
	}
	pxy.startCommonTCPListenersHandler()
	remoteAddr = strings.Join(addrs, ",")
	return
}

func (pxy *HTTPSProxy) Close() {
	pxy.BaseProxy.Close()
}
