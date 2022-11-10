package server

import (
	"errors"
	"net/http"

	"github.com/Allenxuxu/gev"
	"github.com/Allenxuxu/gev/plugins/websocket/ws"
	"github.com/Allenxuxu/gev/plugins/websocket/ws/util"
	"github.com/grpc-boot/base"
	"github.com/grpc-boot/base/core/zaplogger"
)

var (
	ErrProtocolNotExists = errors.New("protocol not exists")
)

type Conn struct {
	first  bool
	header http.Header
	*gev.Connection
}

func newConn(conn *gev.Connection) (id uint64, c *Conn) {
	id = setId(conn)

	c = &Conn{
		first:      true,
		Connection: conn,
	}

	return
}

func (c *Conn) GetId() (id uint64, exists bool) {
	return GetId(c.Connection)
}

func (c *Conn) SendPackage(pkg *base.Package) error {
	proto, exists := c.Get(Protocol)
	if !exists {
		return ErrProtocolNotExists
	}

	data := proto.(base.Protocol).Pack(pkg)
	return c.SendBinary(data)
}

func (c *Conn) SendText(text []byte) error {
	msg, err := util.PackData(ws.MessageText, text)
	if err != nil {
		base.ZapError("pack text msg failed",
			zaplogger.Error(err),
			zaplogger.Value(text),
		)
		return err
	}

	return c.Send(msg)
}

func (c *Conn) SendBinary(data []byte) error {
	msg, err := util.PackData(ws.MessageBinary, data)
	if err != nil {
		base.ZapError("pack binary msg failed",
			zaplogger.Error(err),
			zaplogger.Value(data),
		)
		return err
	}
	return c.Send(msg)
}

func (c *Conn) SendClose(reason string) error {
	msg, err := util.PackCloseData(reason)
	if err != nil {
		base.ZapError("pack close msg failed",
			zaplogger.Error(err),
			zaplogger.Value(reason),
		)
		return err
	}
	return c.Send(msg)
}
