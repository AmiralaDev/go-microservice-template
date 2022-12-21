package wrapper

import (
	"golang.org/x/net/context"
)

type middleware func(HandlerFunc) HandlerFunc
type HandlerFunc func(ctx context.Context, args ...interface{}) (interface{}, error)

func BuildChain(f HandlerFunc, m ...middleware) HandlerFunc {
	if len(m) == 0 {
		return f
	}
	return m[0](BuildChain(f, m[1:cap(m)]...))
}

func (hf HandlerFunc) ToCronJobFunc(ctx context.Context, args ...interface{}) func() {
	return func() {
		hf(ctx, args)
	}
}