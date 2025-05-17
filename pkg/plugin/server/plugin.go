package server

import "context"

const (
	APIVersion = "0.1.0"

	OpLogin       = "Login"
	OpNewProxy    = "NewProxy"
	OpCloseProxy  = "CloseProxy"
	OpPing        = "Ping"
	OpNewWorkConn = "NewWorkConn"
	OpNewUserConn = "NewUserConn"
)

type Plugin interface {
	Name() string
	IsSupport(op string) bool
	Handle(ctx context.Context, op string, content any) (res *Response, retContent any, err error)
}
