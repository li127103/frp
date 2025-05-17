package auth

import (
	"fmt"
	v1 "github.com/li127103/frp/pkg/config/v1"
	"github.com/li127103/frp/pkg/msg"
)

type Setter interface {
	SetLogin(*msg.Login) error
	SetPing(*msg.Ping) error
	SetNewWorkConn(*msg.NewWorkConn) error
}

func NewAuthSetter(cfg v1.AuthClientConfig) (authProvider Setter) {
	switch cfg.Method {
	case v1.AuthMethodToken:
		authProvider = NewTokenAuth(cfg.AdditionalScopes, cfg.Token)
	case v1.AuthMethodOIDC:
		authProvider = NewOidcAuthSetter(cfg.AdditionalScopes, cfg.OIDC)
	default:
		panic(fmt.Sprintf("wrong method: '%s'", cfg.Method))
	}
	return authProvider
}

type Verifier interface {
	VerifyLogin(*msg.Login) error
	VerifyPing(*msg.Ping) error
	VerifyNewWorkConn(*msg.NewWorkConn) error
}

func NewAuthVerifier(cfg v1.AuthServerConfig) (authVerifier Verifier) {
	switch cfg.Method {
	case v1.AuthMethodToken:
		authVerifier = NewTokenAuth(cfg.AdditionalScopes, cfg.Token)
	case v1.AuthMethodOIDC:
		tokenVerifier := NewTokenVerifier(cfg.OIDC)
		authVerifier = NewOidcAuthVerifier(cfg.AdditionalScopes, tokenVerifier)
	}
	return authVerifier
}
