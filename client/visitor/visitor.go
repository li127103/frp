package visitor

import (
	"context"
	v1 "github.com/li127103/frp/pkg/config/v1"
	plugin "github.com/li127103/frp/pkg/plugin/visitor"
	"github.com/li127103/frp/pkg/transport"
	netpkg "github.com/li127103/frp/pkg/util/net"
	"github.com/li127103/frp/pkg/util/xlog"
	"github.com/li127103/frp/pkg/vnet"
	"net"
	"sync"
)

// Helper wraps some functions for visitor to use.
type Helper interface {
	// ConnectServer directly connects to the frp server.
	ConnectServer() (net.Conn, error)
	// TransferConn transfers the connection to another visitor.
	TransferConn(string, net.Conn) error
	// MsgTransporter returns the message transporter that is used to send and receive messages
	// to the frp server through the controller.
	MsgTransporter() transport.MessageTransporter
	// VNetController returns the vnet controller that is used to manage the virtual network.
	VNetController() *vnet.Controller
	// RunID returns the run id of current controller.
	RunID() string
}

// Visitor is used for forward traffics from local port tot remote service.
type Visitor interface {
	Run() error
	AcceptConn(conn net.Conn) error
	Close()
}

func NewVisitor(
	ctx context.Context,
	cfg v1.VisitorConfigurer,
	clientCfg *v1.ClientCommonConfig,
	helper Helper,
) (Visitor, error) {
	xl := xlog.FromContextSafe(ctx).Spawn().AppendPrefix(cfg.GetBaseConfig().Name)
	ctx = xlog.NewContext(ctx, xl)
	var visitor Visitor
	baseVisitor := BaseVisitor{
		clientCfg:  clientCfg,
		helper:     helper,
		ctx:        ctx,
		internalLn: netpkg.NewInternalListener(),
	}
	if cfg.GetBaseConfig().Plugin.Type != "" {
		p, err := plugin.Create(
			cfg.GetBaseConfig().Plugin.Type,
			plugin.PluginContext{
				Name:           cfg.GetBaseConfig().Name,
				Ctx:            ctx,
				VnetController: helper.VNetController(),
				HandleConn: func(conn net.Conn) {
					_ = baseVisitor.AcceptConn(conn)
				},
			},
			cfg.GetBaseConfig().Plugin.VisitorPluginOptions,
		)
		if err != nil {
			return nil, err
		}
		baseVisitor.plugin = p
	}
	switch cfg := cfg.(type) {
	case *v1.STCPVisitorConfig:
		visitor = &STCPVisitor{
			BaseVisitor: &baseVisitor,
			cfg:         cfg,
		}
	case *v1.XTCPVisitorConfig:
		visitor = &XTCPVisitor{
			BaseVisitor:   &baseVisitor,
			cfg:           cfg,
			startTunnelCh: make(chan struct{}),
		}
	case *v1.SUDPVisitorConfig:
		visitor = &SUDPVisitor{
			BaseVisitor:  &baseVisitor,
			cfg:          cfg,
			checkCloseCh: make(chan struct{}),
		}
	}
	return visitor, nil
}

type BaseVisitor struct {
	clientCfg  *v1.ClientCommonConfig
	helper     Helper
	l          net.Listener
	internalLn *netpkg.InternalListener
	plugin     plugin.Plugin

	mu  sync.RWMutex
	ctx context.Context
}

func (v *BaseVisitor) AcceptConn(conn net.Conn) error {
	return v.internalLn.PutConn(conn)
}

func (v *BaseVisitor) Close() {
	if v.l != nil {
		v.l.Close()
	}
	if v.internalLn != nil {
		v.internalLn.Close()
	}
	if v.plugin != nil {
		v.plugin.Close()
	}
}
