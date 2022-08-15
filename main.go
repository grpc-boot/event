package main

import (
	"math/rand"
	"net"
	_ "net/http/pprof"
	"time"

	"event/components/config"
	"event/components/container"
	"event/components/logger"
	"event/core/websocket"

	"github.com/grpc-boot/base"
	"github.com/grpc-boot/base/core/gopool"
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
		err    error
		pool   *gopool.Pool
		server websocket.Server
	)

	conf := container.DefaultContainer.Config()

	pool, err = base.NewGoPool(int(conf.App.MaxWorkers),
		gopool.WithQueueLength(32),
		gopool.WithSpawnSize(1),
	)

	server, err = websocket.NewServer(
		websocket.WithGopool(pool),
	)
	if err != nil {
		base.ZapFatal("new websocket server failed",
			zap.Error(err),
		)
	}

	ln, err := net.Listen("tcp", conf.App.Addr)
	if err != nil {
		base.ZapFatal("listen addr failed",
			zap.Error(err),
		)
	}

	server.OnStart(func(s websocket.Server) error {
		base.ZapInfo("server has started",
			zap.String("Addr", ln.Addr().String()),
		)
		return nil
	})

	if err = server.Serve(ln); err != nil {
		base.ZapFatal("serve failed",
			zap.Error(err),
		)
	}
}
