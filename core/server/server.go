package server

import (
	"math/rand"
	"runtime"

	"event/core/conngroup"

	"github.com/Allenxuxu/gev"
	"github.com/Allenxuxu/gev/plugins/websocket"
	"github.com/Allenxuxu/gev/plugins/websocket/ws"
	"github.com/Allenxuxu/gev/plugins/websocket/ws/util"
	"github.com/grpc-boot/base"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type Server struct {
	nextConntionId atomic.Int64
	connections    *conngroup.ConnGroup
	broadcastCh    chan []byte
	gev            *gev.Server
}

func NewServer() *Server {
	server := &Server{
		connections: conngroup.NewConnGroup(),
		broadcastCh: make(chan []byte, 1024),
	}

	go server.broadcast()

	return server
}

func (s *Server) broadcast() {
	for {
		msg, ok := <-s.broadcastCh
		if !ok {
			break
		}

		data, err := util.PackData(ws.MessageText, msg)
		if err != nil {
			continue
		}

		s.connections.RangeValues(func(values []interface{}) {
			defer func() {
				_ = recover()
			}()

			for _, sess := range values {
				if session, ok := sess.(*Session); ok && session.conn != nil {
					_ = session.conn.Send(data)
				}
			}
		})
	}
}

func (s *Server) Broadcast(msg []byte) {
	s.broadcastCh <- msg
}

func (s *Server) OnMessage(c *gev.Connection, data []byte) (messageType ws.MessageType, out []byte) {
	base.ZapInfo("new msg",
		zap.ByteString("Msg", data),
	)
	id, exists := getId(c)
	if !exists {
		return
	}

	sess, ok := s.connections.Get(id)
	if !ok {
		return
	}

	session, ok := sess.(*Session)
	if !ok {
		return
	}

	if session.first {
		session.first = false
	}

	messageType = ws.MessageBinary
	switch rand.Int() % 4 {
	case 0:
		out = data
	case 1:
		msg, err := util.PackData(ws.MessageText, data)
		if err != nil {
			panic(err)
		}

		if err = c.Send(msg); err != nil {
			msg, err = util.PackCloseData(err.Error())
			if err != nil {
				panic(err)
			}
			if e := c.Send(msg); e != nil {
				panic(e)
			}
		}
	/*case 2:
	msg, err := util.PackCloseData("close")
	if err != nil {
		panic(err)
	}
	if e := c.Send(msg); e != nil {
		panic(e)
	}*/
	case 3:
		go func() {
			msg, err := util.PackData(ws.MessageText, []byte("async write data"))
			if err != nil {
				panic(err)
			}
			if e := c.Send(msg); e != nil {
				panic(e)
			}
		}()
	}
	return
}

func (s *Server) OnConnect(c *gev.Connection) {
	id := s.nextConntionId.Inc()
	setId(c, id)

	session := &Session{
		first: true,
		conn:  c,
	}

	s.connections.Set(id, session)
}

func (s *Server) OnClose(c *gev.Connection) {
	id, exists := getId(c)
	if !exists {
		base.ZapError("not found")
		return
	}
	s.connections.Delete(id)
}

func (s *Server) Shutdown() {
	s.gev.Stop()
}

func (s *Server) Serve(upgrader *ws.Upgrader, opts ...gev.Option) error {
	defaultOpts := []gev.Option{
		gev.Network("tcp"),
		gev.NumLoops(runtime.NumCPU()),
		gev.CustomProtocol(websocket.New(upgrader)),
	}

	opts = append(defaultOpts, opts...)

	ser, err := gev.NewServer(websocket.NewHandlerWrap(upgrader, s), opts...)
	if err != nil {
		return err
	}

	s.gev = ser

	s.gev.Start()

	return nil
}
