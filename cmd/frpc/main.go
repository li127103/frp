package frpc

import (
	_ "github.com/li127103/frp/assets/frps"
	"github.com/li127103/frp/cmd/frpc/sub"
	"github.com/li127103/frp/pkg/util/system"
)

func main() {
	system.EnableCompatibilityMode()
	sub.Execute()
}
