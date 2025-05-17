//go:build !frps

package client

import (
	"context"
	"crypto/tls"
	"github.com/fatedier/golib/pool"
	v1 "github.com/li127103/frp/pkg/config/v1"
	"github.com/li127103/frp/pkg/util/log"
	netpkg "github.com/li127103/frp/pkg/util/net"
	stdlog "log"
	"net/http"
	"net/http/httputil"
)

func init() {
	Register(v1.PluginHTTP2HTTPS, NewHTTP2HTTPSPlugin)
}

type HTTP2HTTPSPlugin struct {
	opts *v1.HTTP2HTTPSPluginOptions

	l *Listener
	s *http.Server
}

func NewHTTP2HTTPSPlugin(_ PluginContext, options v1.ClientPluginOptions) (Plugin, error) {
	opts := options.(*v1.HTTP2HTTPSPluginOptions)
	listener := NewProxyListener()
	p := &HTTP2HTTPSPlugin{
		opts: opts,
		l:    listener,
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	rp := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.Out.Header["X-Forwarded-For"] = r.In.Header["X-Forwarded-For"]
			r.Out.Header["X-Forwarded-Host"] = r.In.Header["X-Forwarded-Host"]
			r.Out.Header["X-Forwarded-Proto"] = r.In.Header["X-Forwarded-Proto"]
			req := r.Out
			req.URL.Scheme = "https"
			req.URL.Host = p.opts.LocalAddr
			if p.opts.HostHeaderRewrite != "" {
				req.Host = p.opts.HostHeaderRewrite
			}
			for k, v := range p.opts.RequestHeaders.Set {
				req.Header.Set(k, v)
			}
		},
		Transport:  tr,
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

func (p *HTTP2HTTPSPlugin) Handle(_ context.Context, connInfo *ConnectionInfo) {
	wrapConn := netpkg.WrapReadWriteCloserToConn(connInfo.Conn, connInfo.UnderlyingConn)
	_ = p.l.PutConn(wrapConn)
}

func (p *HTTP2HTTPSPlugin) Name() string {
	return v1.PluginHTTP2HTTPS
}

func (p *HTTP2HTTPSPlugin) Close() error {
	return p.s.Close()
}
