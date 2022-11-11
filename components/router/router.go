package router

import (
	"event/core/server"
	"event/core/zapkey"
	"github.com/grpc-boot/base"
	"github.com/grpc-boot/base/core/zaplogger"
)

type EventHandler func(conn *server.Conn, pkg *base.Package) error

type Route struct {
	handlers map[uint16][]EventHandler
}

func NewRouter() *Route {
	return &Route{
		handlers: make(map[uint16][]EventHandler),
	}
}

func (r *Route) On(eventId uint16, handlers ...EventHandler) {
	if _, exists := r.handlers[eventId]; !exists {
		r.handlers[eventId] = handlers
		return
	}

	r.handlers[eventId] = append(r.handlers[eventId], handlers...)
}

func (r *Route) trigger(conn *server.Conn, pkg *base.Package) error {
	if pkg == nil {
		return nil
	}

	var err error

	if handlers, exists := r.handlers[pkg.Id]; exists {
		for index, _ := range handlers {
			if er := handlers[index](conn, pkg); er != nil {
				err = er
			}
		}
	}

	return err
}

func (r *Route) ConnectHandle(conn *server.Conn) error {
	base.Debug("connect create",
		zaplogger.Event("connect"),
		zapkey.Address(conn.PeerAddr()),
	)

	return r.trigger(conn, &base.Package{
		Id:    base.EventConnectSuccess,
		Name:  "connect success",
		Param: nil,
	})
}

func (r *Route) Handle(conn *server.Conn, data []byte) error {
	base.Debug("got new msg",
		zaplogger.Data(data),
		zaplogger.Event("message"),
	)

	if base.Bytes2String(data) == "ping" {
		return conn.SendText([]byte("pong"))
	}

	pkg, err := conn.Unpack(data)
	if err != nil {
		base.Error("unpack msg failed",
			zaplogger.Error(err),
			zaplogger.Value(data),
		)

		return err
	}

	err = r.trigger(conn, pkg)
	if err != nil {
		base.Error("handler error",
			zaplogger.Error(err),
			zaplogger.Value(pkg),
		)
	}

	return err
}

func (r *Route) CloseHandle(conn *server.Conn) error {
	base.Debug("connect close",
		zaplogger.Event("close"),
		zapkey.Address(conn.PeerAddr()),
	)

	return r.trigger(conn, &base.Package{
		Id:    base.EventClose,
		Name:  "close",
		Param: nil,
	})
}
