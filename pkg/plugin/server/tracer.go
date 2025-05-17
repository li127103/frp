package server

import "context"

type key int

const (
	reqidKey key = 0
)

func NewReqidContext(ctx context.Context, reqqid string) context.Context {
	return context.WithValue(ctx, reqidKey, reqqid)
}
func GetReqidFromContext(ctx context.Context) string {
	ret, _ := ctx.Value(reqidKey).(string)
	return ret
}
