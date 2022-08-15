package server

import (
	"github.com/grpc-boot/base"
	"go.uber.org/zap"
	"net/http"

	"github.com/Allenxuxu/gev"
	"github.com/Allenxuxu/gev/plugins/websocket/ws"
	"github.com/Allenxuxu/gev/plugins/websocket/ws/util"
)

type Session struct {
	first  bool
	header http.Header
	conn   *gev.Connection
}

func (s *Session) SendText(text []byte) error {
	msg, err := util.PackData(ws.MessageText, text)
	if err != nil {
		base.ZapError("pack text msg failed",
			zap.Error(err),
			zap.ByteString("Data", text),
		)
		return err
	}

	return s.conn.Send(msg)
}
