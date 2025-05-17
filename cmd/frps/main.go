package main

import (
	_ "github.com/li127103/frp/assets/frps"
	_ "github.com/li127103/frp/pkg/metrics"
	"github.com/li127103/frp/pkg/util/system"
)

func main() {
	system.EnableCompatibilityMode()
	Execute()
}
