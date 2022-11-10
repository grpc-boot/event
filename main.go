package main

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"event/components"
	"event/components/router"
	"event/core/server"
	"event/lib/constant"

	"github.com/Allenxuxu/gev"
	"github.com/Allenxuxu/gev/plugins/websocket/ws"
	"github.com/Allenxuxu/gev/plugins/websocket/ws/util"
	"github.com/grpc-boot/base"
	"github.com/grpc-boot/base/core/zaplogger"
	"go.uber.org/zap"
)

func init() {
	var c = &base.Config{}
	err := base.JsonDecodeFile("./conf/app.json", c)
	if err != nil {
		base.RedFatal("read conf file error:%s", err)
	}

	base.Green("run with config:%+v", *c)

	base.DefaultContainer.SetConfig(c)

	err = base.InitZapWithOption(c.Logger)
	if err != nil {
		base.RedFatal("init logger error:%s", err)
	}

	components.Bootstrap()
}

func main() {
	var (
		err error
	)

	var (
		conf      = base.DefaultContainer.Config()
		accept, _ = base.DefaultContainer.Get(constant.Accept)
	)

	wsUpgrader := &ws.Upgrader{}
	wsUpgrader.OnRequest = func(c *gev.Connection, uri []byte) error {
		urlInfo, err := url.ParseRequestURI(base.Bytes2String(uri))
		if err != nil {
			return ws.ErrHandshakeBadUpgrade
		}

		l, _ := strconv.Atoi(urlInfo.Query().Get("l"))
		key := urlInfo.Query().Get("k")
		level := uint8(l)

		base.Green("level:%d", level)

		protocol, err := accept.(*base.Accept).AcceptHex(level, []byte(key))
		if err != nil {
			return ws.ErrHandshakeBadUpgrade
		}

		if level > base.LevelV1 {
			pkg := &base.Package{
				Id:   base.ConnectSuccess,
				Name: "connect success",
				Param: base.JsonParam{
					"data": nil,
				},
			}

			k := protocol.ResponseKey()
			if len(k) > 0 {
				pkg.Param["data"] = k
			}

			text := pkg.Pack()
			msg, err := util.PackData(ws.MessageText, text)
			if err != nil {
				base.ZapError("pack text msg failed",
					zaplogger.Error(err),
					zaplogger.Value(text),
				)
				return ws.ErrHandshakeBadUpgrade
			}

			if err = c.Send(msg); err != nil {
				base.ZapError("send connect success failed",
					zaplogger.Error(err),
					zaplogger.Value(text),
				)
				return ws.ErrHandshakeBadUpgrade
			}
		}

		c.Set(server.Protocol, protocol)
		return nil
	}

	s := server.NewServer()
	r := router.NewRouter()
	s.WithHandler(r)

	go func() {
		tick := time.NewTicker(time.Second)
		for range tick.C {
			msg := &base.Package{
				Id:   base.Tick,
				Name: "tick",
				Param: base.JsonParam{
					"current": time.Now().String(),
				},
			}

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
