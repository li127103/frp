package visitor

import (
	libio "github.com/fatedier/golib/io"
	v1 "github.com/li127103/frp/pkg/config/v1"
	"github.com/li127103/frp/pkg/msg"
	"github.com/li127103/frp/pkg/util/util"
	"github.com/li127103/frp/pkg/util/xlog"
	"io"
	"net"
	"strconv"
	"time"
)

type STCPVisitor struct {
	*BaseVisitor

	cfg *v1.STCPVisitorConfig
}

func (sv *STCPVisitor) Run() (err error) {
	if sv.cfg.BindPort > 0 {
		sv.l, err = net.Listen("tcp", net.JoinHostPort(sv.cfg.BindAddr, strconv.Itoa(sv.cfg.BindPort)))
		if err != nil {
			return
		}
		go sv.worker()
	}

	go sv.internalConnWorker()

	if sv.plugin != nil {
		sv.plugin.Start()
	}
	return
}
func (sv *STCPVisitor) Close() {
	sv.BaseVisitor.Close()
}
func (sv *STCPVisitor) worker() {
	xl := xlog.FromContextSafe(sv.ctx)
	for {
		conn, err := sv.l.Accept()
		if err != nil {
			xl.Warnf("stcp local listener closed")
			return
		}
		go sv.handleConn(conn)
	}
}

func (sv *STCPVisitor) internalConnWorker() {
	xl := xlog.FromContextSafe(sv.ctx)
	for {
		conn, err := sv.internalLn.Accept()
		if err != nil {
			xl.Warnf("stcp internal listener closed")
			return
		}
		go sv.handleConn(conn)
	}
}

func (sv *STCPVisitor) handleConn(userConn net.Conn) {
	xl := xlog.FromContextSafe(sv.ctx)
	defer userConn.Close()

	xl.Debugf("get a new stcp user connection")
	visitorConn, err := sv.helper.ConnectServer()
	if err != nil {
		return
	}
	defer visitorConn.Close()

	now := time.Now().Unix()
	newVisitorConnMsg := &msg.NewVisitorConn{
		RunID:          sv.helper.RunID(),
		ProxyName:      sv.cfg.ServerName,
		SignKey:        util.GetAuthKey(sv.cfg.SecretKey, now),
		Timestamp:      now,
		UseEncryption:  sv.cfg.Transport.UseEncryption,
		UseCompression: sv.cfg.Transport.UseCompression,
	}
	err = msg.WriteMsg(visitorConn, newVisitorConnMsg)
	if err != nil {
		xl.Warnf("send newVisitorConnMsg to server error: %v", err)
		return
	}
	var newVisitorConnRespMsg msg.NewVisitorConnResp
	_ = visitorConn.SetReadDeadline(time.Now().Add(10 * time.Second))
	err = msg.ReadMsgInto(visitorConn, &newVisitorConnRespMsg)
	if err != nil {
		xl.Warnf("get newVisitorConnRespMsg error: %v", err)
		return
	}
	_ = visitorConn.SetReadDeadline(time.Time{})

	if newVisitorConnRespMsg.Error != "" {
		xl.Warnf("start new visitor connection error: %s", newVisitorConnRespMsg.Error)
		return
	}

	var remote io.ReadWriteCloser
	remote = visitorConn
	if sv.cfg.Transport.UseEncryption {
		remote, err = libio.WithEncryption(remote, []byte(sv.cfg.SecretKey))
		if err != nil {
			xl.Errorf("create encryption stream error: %v", err)
			return
		}
	}

	if sv.cfg.Transport.UseCompression {
		var recycleFn func()
		remote, recycleFn = libio.WithCompressionFromPool(remote)
		defer recycleFn()
	}

	libio.Join(userConn, remote)
}
