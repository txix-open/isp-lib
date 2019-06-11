package http

import "github.com/valyala/fasthttp"

type Option func(ss *HttpService)

type ErrorHandler func(ctx *Ctx, err error) interface{}

type UnimplMethodErrorHandler func(ctx *Ctx, actionKey string) interface{}

type Middleware func(ctx *Ctx) (*Ctx, error)

type Interceptor func(ctx *Ctx, proceed func() (interface{}, error)) (interface{}, error)

type Validator func(ctx *Ctx, mappedRequestBody interface{}) error

func WithErrorHandler(em ErrorHandler) Option {
	return func(ss *HttpService) {
		ss.errorHandler = em
	}
}

func WithMiddlewares(mws ...Middleware) Option {
	return func(ss *HttpService) {
		ss.mws = append(ss.mws, mws...)
	}
}

func WithUnimplErrorHandler(em UnimplMethodErrorHandler) Option {
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

func WithInterceptor(interceptor Interceptor) Option {
	return func(ss *HttpService) {
		ss.interceptor = interceptor
	}
}

func WithValidator(validator Validator) Option {
	return func(ss *HttpService) {
		ss.validator = validator
	}
}
