package events

import (
	"event/core/server"

	"github.com/grpc-boot/base"
)

func Connect(conn *server.Conn, pkg *base.Package) error {
	base.Green("connect success")

	return nil
}
