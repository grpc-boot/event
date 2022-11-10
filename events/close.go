package events

import (
	"event/core/server"

	"github.com/grpc-boot/base"
)

func Close(conn *server.Conn, pkg *base.Package) error {
	base.Green("connection close")

	return nil
}
