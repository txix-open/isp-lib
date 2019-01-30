package http

import "github.com/valyala/fasthttp"

type Option func(ss *HttpService)

type ErrorMapper func(ctx *Ctx, err error) interface{}

type UnimplMethodErrorMapper func(ctx *Ctx, actionKey string) interface{}

type Middleware func(ctx *Ctx) (*Ctx, error)

func WithErrorMapping(em ErrorMapper) Option {
	return func(ss *HttpService) {
		ss.errorMapper = em
	}
}

func WithMiddlewares(mws ...Middleware) Option {
	return func(ss *HttpService) {
		ss.mws = append(ss.mws, mws...)
	}
}

func WithUnimplErrorMapper(em UnimplMethodErrorMapper) Option {
	return func(ss *HttpService) {
		ss.unimplErrorMapper = em
	}
}

func WithFastHttpEnhancer(enhancer func(s *fasthttp.Server)) Option {
	return func(ss *HttpService) {
		enhancer(ss.server)
	}
}

func WithPostProcessors(pp ...func(c *Ctx)) Option {
	return func(ss *HttpService) {
		ss.pp = append(ss.pp, pp...)
	}
}
