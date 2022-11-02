package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"event/components/router"
	"event/core/server"
	"github.com/Allenxuxu/gev"
	"github.com/Allenxuxu/gev/plugins/websocket/ws"
	"github.com/grpc-boot/base"
	"github.com/grpc-boot/base/core/zaplogger"
	"go.uber.org/zap"
)

func init() {
	var c = &base.Config{}
	err := base.YamlDecodeFile("./conf/app.yml", c)
	if err != nil {
		base.RedFatal("read conf file error:%s", err)
	}

	base.Green("run with config:%+v", *c)

	base.DefaultContainer.SetConfig(c)

	rand.Seed(time.Now().UnixNano())
	err = base.InitZapWithOption(c.Logger)
	if err != nil {
		base.RedFatal("init logger error:%s", err)
	}
}

func main() {
	var (
		err error
	)

	conf := base.DefaultContainer.Config()

	wsUpgrader := &ws.Upgrader{}
	wsUpgrader.OnHeader = func(c *gev.Connection, key, value []byte) error {
		base.ZapInfo("header",
			zap.ByteString("Key", key),
			zaplogger.Value(value),
		)
		return nil
	}

	wsUpgrader.OnRequest = func(c *gev.Connection, uri []byte) error {
		base.ZapInfo("request",
			zaplogger.Uri(base.Bytes2String(uri)),
			zaplogger.Event("OnRequest"),
		)
		return nil
	}

	s := server.NewServer()
	r := router.NewRouter()
	s.WithHandler(r)

	go func() {
		tick := time.NewTicker(time.Second)
		for range tick.C {
			msg := []byte(time.Now().String())
			s.Broadcast(msg)
		}
	}()

	go handlerSignal(s, conf)

	err = s.Serve(wsUpgrader,
		gev.Network("tcp"),
		gev.Address(conf.Addr),
		gev.NumLoops(conf.Params.Int("numLoops")),
		gev.ReusePort(false),
		gev.IdleTime(time.Second*time.Duration(conf.Params.Int64("maxIdleSeconds"))),
	)
	if err != nil {
		base.Fatal("new server failed",
			zaplogger.Error(err),
		)
	}
}

func handlerSignal(s *server.Server, conf *base.Config) {
	defer func() {
		if er := recover(); er != nil {
			base.ZapError("recover msg",
				zaplogger.Error(er.(error)),
				zaplogger.Event("recover"),
			)
		}

		base.ZapInfo("shutdown")

		if err := s.Shutdown(time.Second * 10); err != nil {
			base.ZapError("shutdown failed",
				zaplogger.Event("shutdown"),
				zaplogger.Error(err),
			)
		}
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGINT, syscall.SIGTERM)
	for {
		sig := <-signalCh
		base.ZapInfo("signal",
			zap.String("Signal", sig.String()),
		)

		switch sig {
		case syscall.SIGUSR1:
			if conf.PprofAddr == "" {
				continue
			}

			go func() {
				if err := base.StartPprof(conf.PprofAddr, nil); err != nil {
					base.ZapError("start pprof failed",
						zaplogger.Event("pprof start"),
						zaplogger.Error(err),
					)
				}
			}()
			continue
		case syscall.SIGUSR2:
			if conf.PprofAddr == "" {
				continue
			}

			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				if err := base.StopPprof(ctx); err != nil {
					base.ZapError("stop pprof failed",
						zaplogger.Event("pprof stop"),
						zaplogger.Error(err),
					)
				}
			}()
			continue
		default:
			return
		}
	}
}
