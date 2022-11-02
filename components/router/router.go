package router

import (
	"event/core/server"
	"event/core/zapkey"

	"github.com/grpc-boot/base"
	"github.com/grpc-boot/base/core/zaplogger"
)

type Route struct {
}

func NewRouter() *Route {
	return &Route{}
}

func (r *Route) ConnectHandle(conn *server.Conn) error {
	base.Debug("connect create",
		zaplogger.Event("connect"),
		zapkey.Address(conn.PeerAddr()),
	)
	return nil
}

func (r *Route) Handle(conn *server.Conn, data []byte) error {
	base.Info("got new msg",
		zaplogger.Data(data),
		zaplogger.Event("message"),
	)
	return nil
}

func (r *Route) CloseHandle(conn *server.Conn) error {
	base.Debug("connect close",
		zaplogger.Event("close"),
		zapkey.Address(conn.PeerAddr()),
	)
	return nil
}
