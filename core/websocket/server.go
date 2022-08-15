package websocket

import (
	"net"
	"runtime"
	"time"

	"github.com/gobwas/ws"
	"github.com/grpc-boot/base"
	"github.com/grpc-boot/base/core/gopool"
	"github.com/mailru/easygo/netpoll"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

var (
	tick = time.NewTicker(time.Second * 5)
)

type Server interface {
	Serve(ln net.Listener) (err error)
	OnStart(func(s Server) error)
	OnShutDown(func() error)
}

type server struct {
	poller          netpoll.Poller
	gopool          *gopool.Pool
	options         *Options
	acceptDesc      *netpoll.Desc
	done            chan struct{}
	hasDone         atomic.Bool
	startHandler    func(s Server) error
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
		options:     options,
		poller:      netPoller,
		gopool:      options.gopool,
		done:        make(chan struct{}, 1),
		connManager: options.connManager,
	}

	go s.tick()

	return s, nil
}

func (s *server) tick() {
	for range tick.C {
		base.ZapInfo("info",
			zap.Int("Total Conn", s.connManager.ConnTotal()),
			zap.Int("Total Guest", s.connManager.GuestTotal()),
		)
	}
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
		base.ZapError("upgrade failed",
			zap.Error(err),
		)
		return
	}

	base.ZapInfo("upgrade success",
		zap.String("Protocol", hs.Protocol),
	)

	desc, err := netpoll.HandleRead(c)
	if err != nil {
		base.ZapError("handler read failed",
			zap.Error(err),
		)
	}

	err = s.poller.Start(desc, func(ev netpoll.Event) {
		if ev&(netpoll.EventReadHup|netpoll.EventHup) != 0 {
			_ = s.poller.Stop(desc)
			_ = conn.Close()
			s.connManager.ReleaseConn(conn)
			return
		}

		err = s.gopool.Submit(func() {
			if err = conn.Receive(); err != nil {
				_ = s.poller.Stop(desc)
				_ = conn.Close()
				s.connManager.ReleaseConn(conn)
				base.ZapError("receive failed",
					zap.Error(err),
				)
			}
		})
		if err != nil {
			base.ZapError("submit receive failed",
				zap.Error(err),
			)
		}
	})

	if err != nil {
		base.ZapError("start poller for conn failed",
			zap.Error(err),
		)
	}
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
				base.ZapError("accept failed",
					zap.Error(er),
				)
				return
			}

			s.handle(conn)
		})

		if err != nil {
			base.ZapError("submit event loop failed",
				zap.Error(err),
			)
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				delay := 5 * time.Millisecond
				time.Sleep(delay)
			}
		}

		if err = s.poller.Resume(s.acceptDesc); err != nil {
			base.ZapError("resume conn failed",
				zap.Error(err),
			)
		}
	})

	if err != nil {
		return err
	}

	if s.startHandler != nil {
		if err = s.startHandler(s); err != nil {
			return err
		}
	}

	<-s.done

	_ = s.shutdown()
	return
}

func (s *server) OnStart(start func(s Server) error) {
	s.startHandler = start
}

func (s *server) OnShutDown(shutdown func() error) {
	s.shutdownHandler = shutdown
}

func (s *server) Shutdown() {
	s.done <- struct{}{}
}

func (s *server) shutdown() error {
	if !s.hasDone.CAS(false, true) {
		return nil
	}

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
