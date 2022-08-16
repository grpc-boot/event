package server

import (
	"context"
	"errors"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	"event/core/conngroup"
	"event/core/zapkey"

	"github.com/Allenxuxu/gev"
	"github.com/Allenxuxu/gev/plugins/websocket"
	"github.com/Allenxuxu/gev/plugins/websocket/ws"
	"github.com/Allenxuxu/gev/plugins/websocket/ws/util"
	"github.com/grpc-boot/base"
	"go.uber.org/atomic"
)

type Handler interface {
	ConnectHandle(conn *Conn) error
	Handle(conn *Conn, data []byte) error
	CloseHandle(conn *Conn) error
}

type Server struct {
	connections     *conngroup.ConnGroup
	server          *gev.Server
	pprofStatus     atomic.Bool
	pprofServer     *http.Server
	broadcastCh     chan []byte
	shutdownHandler func(s *Server) error
	handler         Handler
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
				if er := recover(); er != nil {
					base.ZapError("broadcast failed",
						zapkey.Error(er.(error)),
						zapkey.Event("broadcast"),
					)
				}
			}()

			for _, conn := range values {
				if c, ok := conn.(*Conn); ok && c.Connection != nil {
					_ = c.Send(data)
				}
			}
		})
	}
}

func (s *Server) Broadcast(msg []byte) {
	s.broadcastCh <- msg
}

func (s *Server) OnMessage(c *gev.Connection, data []byte) (messageType ws.MessageType, out []byte) {
	id, exists := GetId(c)
	if !exists {
		if closeData, err := util.PackCloseData("not found id"); err == nil {
			_ = c.Send(closeData)
		}
		return
	}

	conn, ok := s.connections.Get(id)
	if !ok {
		if closeData, err := util.PackCloseData("not found conn"); err == nil {
			_ = c.Send(closeData)
		}
		return
	}

	cn, ok := conn.(*Conn)
	if !ok {
		return
	}

	if cn.first {
		cn.first = false
	}

	if err := s.handler.Handle(cn, data); err != nil {
		base.ZapError("handler message failed",
			zapkey.Error(err),
			zapkey.Event("message"),
		)
	}
	return
}

func (s *Server) OnConnect(c *gev.Connection) {
	id, conn := newConn(c)

	if err := s.handler.ConnectHandle(conn); err != nil {
		_ = conn.SendClose("connect failed")
		return
	}

	s.connections.Set(id, conn)
}

func (s *Server) OnClose(c *gev.Connection) {
	id, exists := GetId(c)
	if !exists {
		base.ZapError("conn not found id",
			zapkey.Event("close"),
		)
		return
	}

	if conn, ok := s.connections.Get(id); ok {
		if cn, yes := conn.(*Conn); yes {
			_ = s.handler.CloseHandle(cn)
		}
	}

	s.connections.Delete(id)
}

func (s *Server) WithHandler(handler Handler) {
	s.handler = handler
}

func (s *Server) WithShutdown(handler func(s *Server) error) {
	s.shutdownHandler = handler
}

func (s *Server) Shutdown(timeout time.Duration) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan struct{}, 1)
	go func() {
		s.server.Stop()
		if s.shutdownHandler != nil {
			err = s.shutdownHandler(s)
		}
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return errors.New("server: shutdown timeout")
	case <-done:
	}
	return
}

func (s *Server) TotalConns() int64 {
	return s.connections.Length()
}

func (s *Server) StartPprof(addr string, handler http.Handler) error {
	if !s.pprofStatus.CAS(false, true) {
		return nil
	}

	s.pprofServer = &http.Server{
		Handler: handler,
		Addr:    addr,
	}

	return s.pprofServer.ListenAndServe()
}

func (s *Server) StopPprof(ctx context.Context) error {
	if !s.pprofStatus.CAS(true, false) {
		return nil
	}
	return s.pprofServer.Shutdown(ctx)
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

	s.server = ser

	s.server.Start()
	return nil
}
