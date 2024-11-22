package server

import (
	"chainmscan/config"
	"chainmscan/db"
	"chainmscan/logger"
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Server struct {
	ctx               context.Context
	logBus            *logger.LoggerBus
	ginEngine         *gin.Engine
	config            *config.Config
	gormDb            *gorm.DB
	ctxCancel         context.CancelFunc
	workerPool        *WorkerPool
	chainList         map[string]chan<- string
	chainListMapMutex sync.Mutex
}
type Option func(s *Server)

func WithGinEngin() Option {
	return func(s *Server) {
		g := gin.New()
		s.ginEngine = g
	}
}

func WithConfig(cfg *config.Config) Option {
	return func(s *Server) {
		s.config = cfg
	}
}

func WithLog(logBus *logger.LoggerBus) Option {
	return func(s *Server) {
		s.logBus = logBus
	}
}

func WithContext(ctx context.Context) Option {
	return func(s *Server) {
		s.ctx, s.ctxCancel = context.WithCancel(ctx)
	}
}

func NewServer(opts ...Option) (*Server, error) {
	server := new(Server)
	for _, opt := range opts {
		opt(server)
	}

	server.chainList = make(map[string]chan<- string)
	server.chainListMapMutex = sync.Mutex{}

	wpLog, err := server.GetZapLogger("WorkerPool")
	if err != nil {
		return nil, err
	}

	server.workerPool, err = NewWorkerPool(server.ctx, wpLog)
	if err != nil {
		return nil, err
	}

	return server, nil
}

func (s *Server) Start() error {
	zlog, err := s.logBus.GetZapLogger("mysql")
	if err != nil {
		return err
	}

	mysqlDb, err := db.MysqlInit(s.config.MysqlConfig, s.config.GormConfig,
		db.TableSlice, zlog)
	if err != nil {
		return err
	}

	s.gormDb = mysqlDb

	// 启动协程管理池
	s.workerPool.Start()

	// 启动gin
	err = s.workerPool.Submit(s.ginRun)
	if err != nil {
		return err
	}

	err = s.SubscriberStart()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) GetZapLogger(name ...string) (*zap.SugaredLogger, error) {
	return s.logBus.GetZapLogger(name...)
}

func (s *Server) GinEngine() *gin.Engine {
	return s.ginEngine
}

func (s *Server) Db() *gorm.DB {
	return s.gormDb
}

func (s *Server) SeverPort() string {
	return s.config.ServerPort
}

func (s *Server) UploadFilePath() string {
	return s.config.UploadFilePath
}

func (s *Server) Stop() error {
	s.workerPool.Stop()
	s.ctxCancel()
	return nil
}

func (s *Server) SysLog() *zap.SugaredLogger {
	logger, _ := s.logBus.GetZapLogger("Server")
	return logger
}

func (s *Server) ginRun(ctx context.Context) error {
	httpServer := &http.Server{
		Addr:    ":" + s.SeverPort(),
		Handler: s.ginEngine,
	}

	// 启动http server
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.SysLog().Errorf("HttpServer listen err: %s\n", err.Error())
			return
		}
	}()

	<-ctx.Done()

	err := httpServer.Shutdown(context.Background())
	if err != nil {
		s.SysLog().Errorf("http server shutdown err: %s\n", err.Error())
		return err
	}

	s.SysLog().Info("http server has been closed ...")
	return nil
}
