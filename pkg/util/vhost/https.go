package vhost

import (
	"crypto/tls"
	libnet "github.com/fatedier/golib/net"
	"io"
	"net"
	"time"
)

type HTTPSMuxer struct {
	*Muxer
}

func NewHTTPSMuxer(listener net.Listener, timeout time.Duration) (*HTTPSMuxer, error) {
	mux, err := NewMuxer(listener, GetHTTPSHostname, timeout)
	mux.SetFailHookFunc(vhostFailed)
	if err != nil {
		return nil, err
	}
	return &HTTPSMuxer{mux}, err
}

func GetHTTPSHostname(c net.Conn) (_ net.Conn, _ map[string]string, err error) {
	reqInfoMap := make(map[string]string, 0)
	sc, rd := libnet.NewSharedConn(c)
	clientHello, err := readClientHello(rd)
	if err != nil {
		return nil, reqInfoMap, err
	}
	reqInfoMap["Host"] = clientHello.ServerName
	reqInfoMap["Scheme"] = "https"
	return sc, reqInfoMap, nil
}

func readClientHello(reader io.Reader) (*tls.ClientHelloInfo, error) {
	var hello *tls.ClientHelloInfo

	// Note that Handshake always fails because the readOnlyConn is not a real connection.
	// As long as the Client Hello is successfully read, the failure should only happen after GetConfigForClient is called,
	// so we only care about the error if hello was never set.
	err := tls.Server(readOnlyConn{reader: reader}, &tls.Config{
		GetConfigForClient: func(argHello *tls.ClientHelloInfo) (*tls.Config, error) {
			hello = &tls.ClientHelloInfo{}
			*hello = *argHello
			return nil, nil
		},
	}).Handshake()

	if hello == nil {
		return nil, err
	}
	return hello, nil
}

func vhostFailed(c net.Conn) {
	// Alert with alertUnrecognizedName
	_ = tls.Server(c, &tls.Config{}).Handshake()
	c.Close()
}

type readOnlyConn struct {
	reader io.Reader
}

func (conn readOnlyConn) Read(p []byte) (int, error)         { return conn.reader.Read(p) }
func (conn readOnlyConn) Write(_ []byte) (int, error)        { return 0, io.ErrClosedPipe }
func (conn readOnlyConn) Close() error                       { return nil }
func (conn readOnlyConn) LocalAddr() net.Addr                { return nil }
func (conn readOnlyConn) RemoteAddr() net.Addr               { return nil }
func (conn readOnlyConn) SetDeadline(_ time.Time) error      { return nil }
func (conn readOnlyConn) SetReadDeadline(_ time.Time) error  { return nil }
func (conn readOnlyConn) SetWriteDeadline(_ time.Time) error { return nil }
