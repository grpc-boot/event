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
	gopool       *gopool.Pool
	panicHandler func(err interface{})
	connManager  ConnManager
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

// WithGopool set go pool
func WithGopool(gp *gopool.Pool) Option {
	return func(opts *Options) {
		opts.gopool = gp
	}
}

// WithConnManager set connection manager
func WithConnManager(pool ConnManager) Option {
	return func(opts *Options) {
		opts.connManager = pool
	}
}
