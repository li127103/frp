//go:build !frps

package client

import (
	"context"
	libio "github.com/fatedier/golib/io"
	v1 "github.com/li127103/frp/pkg/config/v1"
	"github.com/li127103/frp/pkg/util/xlog"
	"net"
)

func init() {
	Register(v1.PluginUnixDomainSocket, NewUnixDomainSocketPlugin)
}

type UnixDomainSocketPlugin struct {
	UnixAddr *net.UnixAddr
}

func NewUnixDomainSocketPlugin(_ PluginContext, options v1.ClientPluginOptions) (p Plugin, err error) {
	opts := options.(*v1.UnixDomainSocketPluginOptions)
	unixAddr, errRet := net.ResolveUnixAddr("unix", opts.UnixPath)
	if errRet != nil {
		err = errRet
		return
	}

	p = &UnixDomainSocketPlugin{
		UnixAddr: unixAddr,
	}
	return
}

func (uds *UnixDomainSocketPlugin) Handle(ctx context.Context, connInfo *ConnectionInfo) {
	xl := xlog.FromContextSafe(ctx)
	localConn, err := net.DialUnix("unix", nil, uds.UnixAddr)
	if err != nil {
		xl.Warnf("dial to uds %s error: %v", uds.UnixAddr, err)
		return
	}
	if connInfo.ProxyProtocolHeader != nil {
		if _, err := connInfo.ProxyProtocolHeader.WriteTo(localConn); err != nil {
			return
		}
	}
	libio.Join(localConn, connInfo.Conn)
}

func (uds *UnixDomainSocketPlugin) Name() string {
	return v1.PluginUnixDomainSocket
}

func (uds *UnixDomainSocketPlugin) Close() error {
	return nil
}
