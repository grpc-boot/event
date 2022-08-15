package main

import (
	"math/rand"
	_ "net/http/pprof"
	"runtime"
	"time"

	"event/components/config"
	"event/components/container"
	"event/components/logger"
	"event/core/server"

	"github.com/Allenxuxu/gev"
	"github.com/Allenxuxu/gev/plugins/websocket"
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

	container.DefaultContainer.SetConfig(c)

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
			zap.ByteString("Value", value),
		)
		return nil
	}

	wsUpgrader.OnRequest = func(c *gev.Connection, uri []byte) error {
		base.ZapInfo("request",
			zap.ByteString("Uri", uri),
			zap.String("Event", "OnRequest"),
		)
		return nil
	}

	handler := server.NewServer()
	go func() {
		tick := time.NewTicker(time.Second)
		for range tick.C {
			msg := []byte(time.Now().String())
			handler.Broadcast(msg)
		}
	}()

	s, err := gev.NewServer(websocket.NewHandlerWrap(wsUpgrader, handler),
		gev.CustomProtocol(websocket.New(wsUpgrader)),
		gev.Network("tcp"),
		gev.Address(conf.App.Addr),
		gev.NumLoops(runtime.NumCPU()))
	if err != nil {
		base.ZapFatal("new server failed",
			zap.Error(err),
		)
	}

	s.Start()
}
