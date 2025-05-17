//go:build !frps

package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/fatedier/golib/pool"
	v1 "github.com/li127103/frp/pkg/config/v1"
	"github.com/li127103/frp/pkg/transport"
	httppkg "github.com/li127103/frp/pkg/util/http"
	"github.com/li127103/frp/pkg/util/log"
	netpkg "github.com/li127103/frp/pkg/util/net"
	"github.com/samber/lo"
	stdlog "log"
	"net/http"
	"net/http/httputil"
	"time"
)

func init() {
	Register(v1.PluginHTTPS2HTTPS, NewHTTPS2HTTPSPlugin)
}

type HTTPS2HTTPSPlugin struct {
	opts *v1.HTTPS2HTTPSPluginOptions

	l *Listener
	s *http.Server
}

func NewHTTPS2HTTPSPlugin(_ PluginContext, options v1.ClientPluginOptions) (Plugin, error) {
	opts := options.(*v1.HTTPS2HTTPSPluginOptions)

	listener := NewProxyListener()
	p := &HTTPS2HTTPSPlugin{
		opts: opts,
		l:    listener,
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	rp := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.Out.Header["X-Forwarded-For"] = r.In.Header["X-Forwarded-For"]
			r.SetXForwarded()
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

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.TLS != nil {
			tlsServerName, _ := httppkg.CanonicalHost(r.TLS.ServerName)
			host, _ := httppkg.CanonicalHost(r.Host)
			if tlsServerName != "" && tlsServerName != host {
				w.WriteHeader(http.StatusMisdirectedRequest)
				return
			}
		}
		rp.ServeHTTP(w, r)
	})

	tlsConfig, err := transport.NewServerTLSConfig(p.opts.CrtPath, p.opts.KeyPath, "")
	if err != nil {
		return nil, fmt.Errorf("gen TLS config error: %v", err)
	}
	p.s = &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 60 * time.Second,
		TLSConfig:         tlsConfig,
	}
	if !lo.FromPtr(opts.EnableHTTP2) {
		p.s.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))
	}

	go func() {
		_ = p.s.ServeTLS(listener, "", "")
	}()
	return p, nil
}

func (p *HTTPS2HTTPSPlugin) Handle(_ context.Context, connInfo *ConnectionInfo) {
	wrapConn := netpkg.WrapReadWriteCloserToConn(connInfo.Conn, connInfo.UnderlyingConn)
	if connInfo.SrcAddr != nil {
		wrapConn.SetRemoteAddr(connInfo.SrcAddr)
	}
	_ = p.l.PutConn(wrapConn)
}

func (p *HTTPS2HTTPSPlugin) Name() string {
	return v1.PluginHTTPS2HTTPS
}

func (p *HTTPS2HTTPSPlugin) Close() error {
	return p.s.Close()
}
