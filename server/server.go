package server

import (
	"chainmscan/blockchain"
	"chainmscan/config"
	"chainmscan/db"
	"chainmscan/logger"
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Server struct {
	ctx        context.Context
	logBus     *logger.LoggerBus
	ginEngine  *gin.Engine
	config     *config.Config
	gormDb     *gorm.DB
	subscriber *blockchain.Subscriber
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
		s.ctx = ctx
	}
}

func NewServer(opts ...Option) (*Server, error) {
	server := new(Server)
	for _, opt := range opts {
		opt(server)
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

	go s.GinEngine().Run(":" + s.SeverPort())

	slog, err := s.logBus.GetZapLogger("subscriber")
	if err != nil {
		return err
	}

	sub, err := blockchain.NewSubscriber(s.ctx, slog)
	if err != nil {
		return err
	}

	s.subscriber = sub
	err = s.subscriber.Start(s.gormDb)
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
