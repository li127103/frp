//go:build !frps

package client

import (
	"context"
	"github.com/fatedier/golib/pool"
	v1 "github.com/li127103/frp/pkg/config/v1"
	"github.com/li127103/frp/pkg/util/log"
	netpkg "github.com/li127103/frp/pkg/util/net"
	stdlog "log"
	"net/http"
	"net/http/httputil"
)

func init() {
	Register(v1.PluginHTTP2HTTP, NewHTTP2HTTPPlugin)
}

type HTTP2HTTPPlugin struct {
	opts *v1.HTTP2HTTPPluginOptions

	l *Listener
	s *http.Server
}

func NewHTTP2HTTPPlugin(_ PluginContext, options v1.ClientPluginOptions) (Plugin, error) {
	opts := options.(*v1.HTTP2HTTPPluginOptions)

	listener := NewProxyListener()

	p := &HTTP2HTTPPlugin{
		opts: opts,
		l:    listener,
	}

	rp := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			req := r.Out
			req.URL.Scheme = "http"
			req.URL.Host = p.opts.LocalAddr
			if p.opts.HostHeaderRewrite != "" {
				req.Host = p.opts.HostHeaderRewrite
			}
			for k, v := range p.opts.RequestHeaders.Set {
				req.Header.Set(k, v)
			}
		},
		BufferPool: pool.NewBuffer(32 * 1024),
		ErrorLog:   stdlog.New(log.NewWriteLogger(log.WarnLevel, 2), "", 0),
	}

	p.s = &http.Server{
		Handler:           rp,
		ReadHeaderTimeout: 0,
	}

	go func() {
		_ = p.s.Serve(listener)
	}()

	return p, nil
}

func (p *HTTP2HTTPPlugin) Handle(_ context.Context, connInfo *ConnectionInfo) {
	wrapConn := netpkg.WrapReadWriteCloserToConn(connInfo.Conn, connInfo.UnderlyingConn)
	_ = p.l.PutConn(wrapConn)
}

func (p *HTTP2HTTPPlugin) Name() string {
	return v1.PluginHTTP2HTTP
}

func (p *HTTP2HTTPPlugin) Close() error {
	return p.s.Close()
}
