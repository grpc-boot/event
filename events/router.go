package events

import (
	"event/components/router"

	"github.com/grpc-boot/base"
)

const (
	EventMessage = 0x0300
)

func LoadRouter() *router.Route {
	r := router.NewRouter()

	r.On(base.EventClose, Close)
	r.On(base.EventConnectSuccess, Connect)
	r.On(EventMessage, Message)

	return r
}
