//go:build !darwin && !linux

package vnet

import (
	"context"
	"fmt"
	"golang.zx2c4.com/wireguard/tun"
	"runtime"
)

func openTun(_ context.Context, _ string) (tun.Device, error) {
	return nil, fmt.Errorf("virtual net is not supported on this platform (%s/%s)", runtime.GOOS, runtime.GOARCH)
}
