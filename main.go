package main

import (
	"log"
	"math/rand"
	"net"
	_ "net/http/pprof"
	"time"

	"event/components/container"

	"event/core/chat"

	"github.com/gobwas/ws"
	"github.com/grpc-boot/base"
	"github.com/grpc-boot/base/core/gopool"
	"github.com/mailru/easygo/netpoll"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	var (
		err    error
		poller netpoll.Poller
		pool   *gopool.Pool
	)

	conf := container.DefaultContainer.Config()
	poller, err = netpoll.New(nil)
	if err != nil {
		log.Fatal(err)
	}

	pool, err = base.NewGoPool(int(conf.App.MaxWorkers),
		gopool.WithQueueLength(32),
		gopool.WithSpawnSize(1),
	)

	var (
		chat = chat.NewChat(pool)
		exit = make(chan struct{})
	)

	// handle is a new incoming connection handler.
	// It upgrades TCP connection to WebSocket, registers netpoll listener on
	// it and stores it as a chat user in Chat instance.
	//
	// We will call it below within accept() loop.
	handle := func(conn net.Conn) {
		var (
			hs ws.Handshake
		)
		safeConn := deadliner{conn, time.Millisecond * time.Duration(conf.App.IoTimeoutMs)}

		hs, err = ws.Upgrade(safeConn)
		if err != nil {
			log.Printf("%s: upgrade error: %v", nameConn(conn), err)
			conn.Close()
			return
		}

		log.Printf("%s: established websocket connection: %+v", nameConn(conn), hs)

		// Register incoming user in chat.
		user := chat.Register(safeConn)

		// Create netpoll event descriptor for conn.
		// We want to handle only read events of it.
		desc := netpoll.Must(netpoll.HandleRead(conn))

		// Subscribe to events about conn.
		poller.Start(desc, func(ev netpoll.Event) {
			if ev&(netpoll.EventReadHup|netpoll.EventHup) != 0 {
				// When ReadHup or Hup received, this mean that client has
				// closed at least write end of the connection or connections
				// itself. So we want to stop receive events about such conn
				// and remove it from the chat registry.
				poller.Stop(desc)
				chat.Remove(user)
				return
			}
			// Here we can read some new message from connection.
			// We can not read it right here in callback, because then we will
			// block the poller's inner loop.
			// We do not want to spawn a new goroutine to read single message.
			// But we want to reuse previously spawned goroutine.
			pool.Submit(func() {
				if err := user.Receive(); err != nil {
					// When receive failed, we can only disconnect broken
					// connection and stop to receive events about it.
					poller.Stop(desc)
					chat.Remove(user)
				}
			})
		})
	}

	ln, err := net.Listen("tcp", conf.App.Addr)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("websocket is listening on %s", ln.Addr().String())

	// Create netpoll descriptor for the listener.
	// We use OneShot here to manually resume events stream when we want to.
	acceptDesc := netpoll.Must(netpoll.HandleListener(
		ln, netpoll.EventRead|netpoll.EventOneShot,
	))

	// accept is a channel to signal about next incoming connection Accept()
	// results.
	accept := make(chan error, 1)

	poller.Start(acceptDesc, func(e netpoll.Event) {
		err = pool.SubmitTimeout(time.Millisecond, func() {
			conn, err := ln.Accept()
			if err != nil {
				accept <- err
				return
			}

			accept <- nil
			handle(conn)
		})
		if err == nil {
			err = <-accept
		}
		if err != nil {
			if err != gopool.ErrSubmitTimeout {
				goto cooldown
			}
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				goto cooldown
			}

			log.Fatalf("accept error: %v", err)

		cooldown:
			delay := 5 * time.Millisecond
			log.Printf("accept error: %v; retrying in %s", err, delay)
			time.Sleep(delay)
		}

		poller.Resume(acceptDesc)
	})

	<-exit
}

func nameConn(conn net.Conn) string {
	return conn.LocalAddr().String() + " > " + conn.RemoteAddr().String()
}

type deadliner struct {
	net.Conn
	t time.Duration
}

func (d deadliner) Write(p []byte) (int, error) {
	if err := d.Conn.SetWriteDeadline(time.Now().Add(d.t)); err != nil {
		return 0, err
	}
	return d.Conn.Write(p)
}

func (d deadliner) Read(p []byte) (int, error) {
	if err := d.Conn.SetReadDeadline(time.Now().Add(d.t)); err != nil {
		return 0, err
	}
	return d.Conn.Read(p)
}
