//go:build !frps

package client

import (
	"context"
	gosocks5 "github.com/armon/go-socks5"
	v1 "github.com/li127103/frp/pkg/config/v1"
	netpkg "github.com/li127103/frp/pkg/util/net"
	"io"
	"log"
)

func init() {
	Register(v1.PluginSocks5, NewSocks5Plugin)
}

type Socks5Plugin struct {
	Server *gosocks5.Server
}

func NewSocks5Plugin(_ PluginContext, options v1.ClientPluginOptions) (p Plugin, err error) {
	opts := options.(*v1.Socks5PluginOptions)

	cfg := &gosocks5.Config{
		Logger: log.New(io.Discard, "", log.LstdFlags),
	}
	if opts.Username != "" && opts.Password != "" {
		cfg.Credentials = gosocks5.StaticCredentials(map[string]string{opts.Username: opts.Password})
	}
	sp := &Socks5Plugin{}
	sp.Server, err = gosocks5.New(cfg)
	p = sp
	return
}

func (sp *Socks5Plugin) Handle(_ context.Context, connInfo *ConnectionInfo) {
	defer connInfo.Conn.Close()
	wrapConn := netpkg.WrapReadWriteCloserToConn(connInfo.Conn, connInfo.UnderlyingConn)
	_ = sp.Server.ServeConn(wrapConn)
}

func (sp *Socks5Plugin) Name() string {
	return v1.PluginSocks5
}

func (sp *Socks5Plugin) Close() error {
	return nil
}
