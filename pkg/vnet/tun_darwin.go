package vnet

import (
	"golang.zx2c4.com/wireguard/tun"
	"net"
	"os/exec"
)

const (
	defaultTunName = "utun"
	defaultMTU     = 1420
)

func openTun(_ context.Context, addr string) (tun.Device, error) {
	dev, err := tun.CreateTUN(defaultTunName, defaultMTU)
	if err != nil {
		return nil, err
	}
	name, err := dev.Name()
	if err != nil {
		return nil, err
	}

	ip, ipNet, err := net.ParseCIDR(addr)
	if err != nil {
		return nil, err
	}

	// Calculate a peer IP for the point-to-point tunnel
	peerIP := generatePeerIP(ip)

	// Configure the interface with proper point-to-point addressing
	if err = exec.Command("ifconfig", name, "inet", ip.String(), peerIP.String(), "mtu", fmt.Sprint(defaultMTU), "up").Run(); err != nil {
		return nil, err
	}

	// Add default route for the tunnel subnet
	routes := []net.IPNet{*ipNet}
	if err = addRoutes(name, routes); err != nil {
		return nil, err
	}
	return dev, nil
}

func generatePeerIP(ip net.IP) net.IP {
	// Make a copy to avoid modifying the original
	peerIP := make(net.IP, len(ip))
	copy(peerIP, ip)

	// Increment the last octet
	peerIP[len(peerIP)-1]++

	return peerIP
}

// addRoutes configures system routes for the TUN interface
func addRoutes(ifName string, routes []net.IPNet) error {
	for _, route := range routes {
		routeStr := route.String()
		if err := exec.Command("route", "add", "-net", routeStr, "-interface", ifName).Run(); err != nil {
			return err
		}
	}
	return nil
}
