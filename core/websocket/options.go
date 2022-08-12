package websocket

import (
	"fmt"

	"github.com/grpc-boot/base/core/gopool"
)

var (
	defaultOptions = func() *Options {
		return &Options{
			panicHandler: func(err interface{}) {
				fmt.Printf("panic error:%+v", err)
			},
			connManager: NewConnManager(),
		}
	}
)

type Options struct {
	gopool          *gopool.Pool
	panicHandler    func(err interface{})
	shutdownHandler func() error
	connManager     ConnManager
}

type Option func(opts *Options)

func loadOptions(options ...Option) *Options {
	opts := defaultOptions()
	for _, option := range options {
		option(opts)
	}
	return opts
}

// WithPanicHandler set panicHandler
func WithPanicHandler(panicHandler func(err interface{})) Option {
	return func(opts *Options) {
		opts.panicHandler = panicHandler
	}
}

// WithShutdownHandler set shutdown handler
func WithShutdownHandler(shutdownHandler func() error) Option {
	return func(opts *Options) {
		opts.shutdownHandler = shutdownHandler
	}
}

// WithGopool set go pool
func WithGopool(gp *gopool.Pool) Option {
	return func(opts *Options) {
		opts.gopool = gp
	}
}

// WithConnPool set connection pool
func WithConnPool(pool ConnManager) Option {
	return func(opts *Options) {
		opts.connManager = pool
	}
}
