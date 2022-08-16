package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"event/components/config"
	"event/components/container"
	"event/components/logger"
	"event/components/router"
	"event/core/helper"
	"event/core/server"
	"event/core/zapkey"

	"github.com/Allenxuxu/gev"
	"github.com/Allenxuxu/gev/plugins/websocket/ws"
	"github.com/grpc-boot/base"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	var c = &config.Config{}
	err := base.YamlDecodeFile("./conf/app.yml", c)
	if err != nil {
		base.RedFatal("read conf file error:%s", err)
	}

	container.DefaultContainer.SetConfig(c.Format())

	rand.Seed(time.Now().UnixNano())

	err = logger.InitLoggerWithPath(zapcore.Level(c.Logger.Level), c.Logger.DebugPath, c.Logger.InfoPath, c.Logger.ErrorPath, nil)
	if err != nil {
		base.RedFatal("init logger error:%s", err)
	}
}

func main() {
	var (
		err error
	)

	conf := container.DefaultContainer.Config()

	wsUpgrader := &ws.Upgrader{}
	wsUpgrader.OnHeader = func(c *gev.Connection, key, value []byte) error {
		base.ZapInfo("header",
			zap.ByteString("Key", key),
			zapkey.Value(value),
		)
		return nil
	}

	wsUpgrader.OnRequest = func(c *gev.Connection, uri []byte) error {
		base.ZapInfo("request",
			zapkey.Uri(helper.Bytes2String(uri)),
			zapkey.Event("OnRequest"),
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

	go func() {
		err = s.Serve(wsUpgrader,
			gev.Network("tcp"),
			gev.Address(conf.App.Addr),
			gev.NumLoops(conf.App.NumLoops),
			gev.IdleTime(time.Second*time.Duration(conf.App.MaxIdleSeconds)),
		)

		if err != nil {
			base.ZapFatal("new server failed",
				zapkey.Error(err),
			)
		}
	}()

	defer func() {
		if err = s.Shutdown(time.Second * 10); err != nil {
			base.ZapError("shutdown failed",
				zapkey.Event("shutdown"),
				zapkey.Error(err),
			)
		}
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGPROF)
	for {
		sig := <-signalCh
		switch sig {
		case syscall.SIGUSR1:
			base.Green("got user1")
			if err = s.StartPprof(conf.App.PprofAddr, nil); err != nil {
				base.ZapError("start pprof failed",
					zapkey.Event("pprof start"),
					zapkey.Error(err),
				)
			}
			continue
		case syscall.SIGUSR2:
			base.Green("got user2")
			func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				if err = s.StopPprof(ctx); err != nil {
					base.ZapError("stop pprof failed",
						zapkey.Event("pprof stop"),
						zapkey.Error(err),
					)
				}
			}()
			continue
		default:
		}
		break
	}
}
