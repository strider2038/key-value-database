package network

import "context"

type Handler interface {
	Handle(ctx context.Context, request []byte) []byte
}

type HandlerFunc func(ctx context.Context, request []byte) []byte

func (f HandlerFunc) Handle(ctx context.Context, request []byte) []byte {
	return f(ctx, request)
}
