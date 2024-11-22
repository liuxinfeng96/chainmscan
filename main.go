package main

import (
	"chainmscan/api"
	"chainmscan/config"
	"chainmscan/logger"
	"chainmscan/server"
	"context"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	conf, err := config.InitConfig("")
	if err != nil {
		panic(err)
	}

	logBus := logger.NewLoggerBus(conf.LogConfig)

	s, err := server.NewServer(
		server.WithConfig(conf),
		server.WithGinEngin(),
		server.WithContext(context.Background()),
		server.WithLog(logBus),
	)
	if err != nil {
		panic(err)
	}

	err = api.LoadHttpHandlers(s)
	if err != nil {
		panic(err)
	}

	err = s.Start()
	if err != nil {
		panic(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	<-signals

	err = s.Stop()
	if err != nil {
		s.SysLog().Error("server stop err: %s", err.Error())
	}

	s.SysLog().Info("the service exits normally")
}
