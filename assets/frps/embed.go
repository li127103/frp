package frps

import (
	"embed"
	"github.com/li127103/frp/assets"
)

//go:embed static/*
var content embed.FS

func init() {
	assets.Register(content)
}
