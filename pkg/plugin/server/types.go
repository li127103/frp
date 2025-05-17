package server

import "github.com/li127103/frp/pkg/msg"

type Request struct {
	Version string `json:"version"`
	Op      string `json:"op"`
	Content any    `json:"content"`
}

type Response struct {
	Reject       bool   `json:"reject"`
	RejectReason string `json:"reject_reason"`
	Unchange     bool   `json:"unchange"`
	Content      any    `json:"content"`
}

type LoginContent struct {
	msg.Login

	ClientAddress string `json:"client_address,omitempty"`
}

type UserInfo struct {
	User  string            `json:"user"`
	Metas map[string]string `json:"metas"`
	RunID string            `json:"run_id"`
}

type NewProxyContent struct {
	User UserInfo `json:"user"`
	msg.NewProxy
}

type CloseProxyContent struct {
	User UserInfo `json:"user"`
	msg.CloseProxy
}

type PingContent struct {
	User UserInfo `json:"user"`
	msg.Ping
}

type NewWorkConnContent struct {
	User UserInfo `json:"user"`
	msg.NewWorkConn
}

type NewUserConnContent struct {
	User       UserInfo `json:"user"`
	ProxyName  string   `json:"proxy_name"`
	ProxyType  string   `json:"proxy_type"`
	RemoteAddr string   `json:"remote_addr"`
}
