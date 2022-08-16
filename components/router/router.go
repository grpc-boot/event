package router

import (
	"event/core/server"
	"event/core/zapkey"

	"github.com/grpc-boot/base"
)

type Route struct {
}

func NewRouter() *Route {
	return &Route{}
}

func (r *Route) ConnectHandle(conn *server.Conn) error {
	base.ZapDebug("connect create",
		zapkey.Event("connect"),
		zapkey.Address(conn.PeerAddr()),
	)
	return nil
}

func (r *Route) Handle(conn *server.Conn, data []byte) error {
	base.ZapInfo("got new msg",
		zapkey.Data(data),
		zapkey.Event("message"),
	)
	return nil
}

func (r *Route) CloseHandle(conn *server.Conn) error {
	base.ZapDebug("connect close",
		zapkey.Event("close"),
		zapkey.Address(conn.PeerAddr()),
	)
	return nil
}
