package websocket

import (
	"fmt"
	"net"
	"runtime"
	"time"

	"event/core/chat"

	"github.com/gobwas/ws"
	"github.com/grpc-boot/base/core/gopool"
	"github.com/mailru/easygo/netpoll"
)

type Server interface {
	Serve(ln net.Listener) (err error)
}

type server struct {
	poller          netpoll.Poller
	gopool          *gopool.Pool
	options         *Options
	acceptDesc      *netpoll.Desc
	done            chan struct{}
	shutdownHandler func() error
	connManager     ConnManager
}

func NewServer(opts ...Option) (Server, error) {
	netPoller, err := netpoll.New(nil)
	if err != nil {
		return nil, err
	}

	options := loadOptions(opts...)
	if options.gopool == nil {
		var (
			gp     *gopool.Pool
			cpuNum = runtime.NumCPU()
		)

		gp, err = gopool.NewPool(cpuNum*10,
			gopool.WithSpawnSize(cpuNum),
			gopool.WithQueueLength(8),
		)

		if err != nil {
			return nil, err
		}
		options.gopool = gp
	}

	s := &server{
		options:         options,
		poller:          netPoller,
		gopool:          options.gopool,
		done:            make(chan struct{}, 1),
		connManager:     options.connManager,
		shutdownHandler: options.shutdownHandler,
	}

	return s, nil
}

func (s *server) handle(c net.Conn) {
	var (
		err error
		hs  ws.Handshake
	)

	conn := s.connManager.AcquireConn(c)

	hs, err = ws.Upgrade(conn)
	if err != nil {
		_ = conn.Close()
		s.connManager.ReleaseConn(conn)
		return
	}

	fmt.Println(hs)

	// Register incoming user in chat.
	user := chat.Register(safeConn)

	desc, err := netpoll.HandleRead(c)

	s.poller.Start(desc, func(ev netpoll.Event) {
		if ev&(netpoll.EventReadHup|netpoll.EventHup) != 0 {
			// When ReadHup or Hup received, this mean that client has
			// closed at least write end of the connection or connections
			// itself. So we want to stop receive events about such conn
			// and remove it from the chat registry.
			s.poller.Stop(desc)
			chat.Remove(user)
			return
		}

		s.gopool.Submit(func() {
			if err := user.Receive(); err != nil {
				// When receive failed, we can only disconnect broken
				// connection and stop to receive events about it.
				poller.Stop(desc)
				chat.Remove(user)
			}
		})
	})
}

func (s *server) Serve(ln net.Listener) (err error) {
	s.acceptDesc, err = netpoll.HandleListener(ln, netpoll.EventRead|netpoll.EventOneShot)
	if err != nil {
		return
	}

	err = s.poller.Start(s.acceptDesc, func(e netpoll.Event) {
		err = s.gopool.Submit(func() {
			conn, er := ln.Accept()
			if er != nil {
				return
			}

			s.handle(conn)
		})

		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				delay := 5 * time.Millisecond
				time.Sleep(delay)
			}
		}

		err = s.poller.Resume(s.acceptDesc)
	})

	if err != nil {
		return err
	}

	<-s.done

	return
}

func (s *server) Shutdown() {
	s.done <- struct{}{}
}

func (s *server) shutdown() error {
	err := s.poller.Stop(s.acceptDesc)

	for {
		if s.gopool.PendingTaskTotal() == 0 {
			break
		}
		time.Sleep(time.Millisecond * 10)
	}

	if s.shutdownHandler != nil {
		return s.shutdownHandler()
	}

	return err
}
