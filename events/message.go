package events

import (
	"time"

	"event/core/server"

	"github.com/grpc-boot/base"
)

func Message(conn *server.Conn, pkg *base.Package) error {
	time.Sleep(time.Second)

	return conn.Emit(pkg)
}
