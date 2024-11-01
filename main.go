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

	context, cancel := context.WithCancel(context.Background())

	s, err := server.NewServer(
		server.WithConfig(conf),
		server.WithGinEngin(),
		server.WithContext(context),
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

	// 捕捉系统quit信号
	defer func() {
		cancel()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	<-signals
}
