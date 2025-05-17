package visitor

import (
	"context"
	"errors"
	"fmt"
	v1 "github.com/li127103/frp/pkg/config/v1"
	netutil "github.com/li127103/frp/pkg/util/net"
	"github.com/li127103/frp/pkg/util/xlog"
	"net"
	"sync"
	"time"
)

func init() {
	Register(v1.VisitorPluginVirtualNet, NewVirtualNetPlugin)
}

type VirtualNetPlugin struct {
	pluginCtx PluginContext

	routes []net.IPNet

	mu             sync.Mutex
	controllerConn net.Conn
	closeSignal    chan struct{}

	ctx    context.Context
	cancel context.CancelFunc
}

func NewVirtualNetPlugin(pluginCtx PluginContext, options v1.VisitorPluginOptions) (Plugin, error) {
	opts := options.(*v1.VirtualNetVisitorPluginOptions)
	p := &VirtualNetPlugin{
		pluginCtx: pluginCtx,
		routes:    make([]net.IPNet, 0),
	}

	p.ctx, p.cancel = context.WithCancel(pluginCtx.Ctx)
	if opts.DestinationIP == "" {
		return nil, errors.New("destinationIP is required")
	}
	// Parse DestinationIP and create a host route.
	ip := net.ParseIP(opts.DestinationIP)
	if ip == nil {
		return nil, fmt.Errorf("invalid destination IP address [%s]", opts.DestinationIP)
	}

	var mask net.IPMask
	if ip.To4() != nil {
		mask = net.CIDRMask(32, 32) // /32 for IPv4
	} else {
		mask = net.CIDRMask(128, 128) // /128 for IPv6
	}
	p.routes = append(p.routes, net.IPNet{IP: ip, Mask: mask})

	return p, nil
}

func (p *VirtualNetPlugin) Name() string {
	return v1.VisitorPluginVirtualNet
}

func (p *VirtualNetPlugin) Start() {
	xl := xlog.FromContextSafe(p.pluginCtx.Ctx)
	if p.pluginCtx.VnetController == nil {
		return
	}

	routeStr := "unknown"
	if len(p.routes) > 0 {
		routeStr = p.routes[0].String()
	}
	xl.Infof("starting VirtualNetPlugin for visitor [%s], attempting to register routes for %s", p.pluginCtx.Name, routeStr)
	go p.run()
}
func (p *VirtualNetPlugin) run() {
	xl := xlog.FromContextSafe(p.pluginCtx.Ctx)
	reconnectDelay := 10 * time.Second

	for {
		currentCloseSignal := make(chan struct{})
		p.mu.Lock()
		p.closeSignal = currentCloseSignal
		p.mu.Unlock()

		select {
		case <-p.ctx.Done():
			xl.Infof("VirtualNetPlugin run loop for visitor [%s] stopping (context cancelled before pipe creation).", p.pluginCtx.Name)
			p.cleanupControllerConn(xl)
			return
		default:
		}

		controllerConn, pluginConn := net.Pipe()

		p.mu.Lock()
		p.controllerConn = controllerConn
		p.mu.Unlock()

		pluginNotifyConn := netutil.WrapCloseNotifyConn(pluginConn, func() {
			close(currentCloseSignal) // Signal the run loop on close.
		})
		xl.Infof("attempting to register client route for visitor [%s]", p.pluginCtx.Name)
		p.pluginCtx.VnetController.RegisterClientRoute(p.ctx, p.pluginCtx.Name, p.routes, pluginConn)
		xl.Infof("successfully registered client route for visitor [%s]. Starting connection handler with CloseNotifyConn.", p.pluginCtx.Name)

		// Pass the CloseNotifyConn to HandleConn.
		// HandleConn is responsible for calling Close() on pluginNotifyConn.
		p.pluginCtx.HandleConn(pluginNotifyConn)
		// Wait for context cancellation or connection close.
		select {
		case <-p.ctx.Done():
			xl.Infof("VirtualNetPlugin run loop stopping for visitor [%s] (context cancelled while waiting).", p.pluginCtx.Name)
			p.cleanupControllerConn(xl)
			return
		case <-currentCloseSignal:
			xl.Infof("detected connection closed via CloseNotifyConn for visitor [%s].", p.pluginCtx.Name)
			// HandleConn closed the plugin side. Close the controller side.
			p.cleanupControllerConn(xl)
			xl.Infof("waiting %v before attempting reconnection for visitor [%s]...", reconnectDelay, p.pluginCtx.Name)
			select {
			case <-time.After(reconnectDelay):
			case <-p.ctx.Done():
				xl.Infof("VirtualNetPlugin reconnection delay interrupted for visitor [%s]", p.pluginCtx.Name)
				return
			}
		}
		xl.Infof("re-establishing virtual connection for visitor [%s]...", p.pluginCtx.Name)
	}
}

// cleanupControllerConn closes the current controllerConn (if it exists) under lock.
func (p *VirtualNetPlugin) cleanupControllerConn(xl *xlog.Logger) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.controllerConn != nil {
		xl.Debugf("cleaning up controllerConn for visitor [%s]", p.pluginCtx.Name)
		p.controllerConn.Close()
		p.controllerConn = nil
	}

	p.closeSignal = nil
}

// Close initiates the plugin shutdown.
func (p *VirtualNetPlugin) Close() error {
	xl := xlog.FromContextSafe(p.pluginCtx.Ctx)
	xl.Infof("closing VirtualNetPlugin for visitor [%s]", p.pluginCtx.Name)
	// Signal the run loop goroutine to stop.
	p.cancel()

	// Unregister the route from the controller.
	if p.pluginCtx.VnetController != nil {
		p.pluginCtx.VnetController.UnregisterClientRoute(p.pluginCtx.Name)
		xl.Infof("unregistered client route for visitor [%s]", p.pluginCtx.Name)
	}
	// Explicitly close the controller side of the pipe.
	// This ensures the pipe is broken even if the run loop is stuck or HandleConn hasn't closed its end.
	p.cleanupControllerConn(xl)
	xl.Infof("finished cleaning up connections during close for visitor [%s]", p.pluginCtx.Name)

	return nil
}
