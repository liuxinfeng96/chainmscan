package logger

import (
	rotatelogs "chainmscan/logger/file-rotatelogs"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang/groupcache/lru"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 日志级别，配置文件定义的常量
const (
	DEBUG = "DEBUG"
	INFO  = "INFO"
	WARN  = "WARN"
	ERROR = "ERROR"
)

const (
	DefaultLogPath = "./log/sys.log"

	// DefaultLogMaxAge 默认最大保留天数
	DefaultLogMaxAge = 30

	DefaultLogLevel = INFO

	// DefaultLogRotationTime 日志滚动时间（小时）
	DefaultLogRotationTime = 24

	// DefaultLogRotationSize 日志滚动大小（MB）暂时不用
	DefaultLogRotationSize = 100
)

// LogConfig 日志记录的配置
type LogConfig struct {
	LogPath string `mapstructure:"log_path"`

	LogLevel string `mapstructure:"log_level"`

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	MaxAge int `mapstructure:"max_age"`

	RotationTime int `mapstructure:"rotation_time"`

	RotationSize int64 `mapstructure:"rotation_size"`

	// jsonFormat: log file use json format
	JsonFormat bool `mapstructure:"json_format"`

	// showLine: show filename and line number
	ShowLine bool `mapstructure:"show_line"`

	// logInConsole: show logs in console at the same time
	LogInConsole bool `mapstructure:"log_in_console"`

	// if true, only show log, won't print log level、caller func and line
	IsBrief bool `mapstructure:"is_brief"`

	// StackTraceLevel record a stack trace for all messages at or above a given level.
	// Empty string or invalid level will not open stack trace.
	StackTraceLevel string `mapstructure:"stack_trace_level"`
}

type LoggerBus struct {
	logConfig *LogConfig
	logMutex  sync.Mutex
	logCache  *lru.Cache
}

var Logger *LoggerBus

func NewLoggerBus(config *LogConfig) *LoggerBus {
	var lb LoggerBus

	lb.logConfig = config
	lb.logMutex = sync.Mutex{}
	lb.logCache = lru.New(1024)

	return &lb
}

func SetLogConfig(config *LogConfig) {
	Logger = NewLoggerBus(config)
}

// GetZapLogger 创建/获取模块日志对象
func GetZapLogger(modelName ...string) (*zap.SugaredLogger, error) {
	Logger.logMutex.Lock()
	defer Logger.logMutex.Unlock()
	var name string
	for _, v := range modelName {
		name += fmt.Sprintf("[@%s]", v)
	}
	if len(name) == 0 {
		name = "[@default]"
	}

	zlog, ok := Logger.logCache.Get(name)
	if !ok {
		log, err := initLogger(Logger.logConfig, name)
		if err != nil {
			return nil, err
		}
		Logger.logCache.Add(name, log)
		return log, nil
	}

	log := zlog.(*zap.SugaredLogger)
	return log, nil
}

// GetZapLogger 创建/获取模块日志对象
func (l *LoggerBus) GetZapLogger(modelName ...string) (*zap.SugaredLogger, error) {
	l.logMutex.Lock()
	defer l.logMutex.Unlock()
	var name string
	for _, v := range modelName {
		name += fmt.Sprintf("[@%s]", v)
	}
	if len(name) == 0 {
		name = "[@default]"
	}

	zlog, ok := l.logCache.Get(name)
	if !ok {
		log, err := initLogger(l.logConfig, name)
		if err != nil {
			return nil, err
		}

		l.logCache.Add(name, log)
		return log, nil
	}

	log := zlog.(*zap.SugaredLogger)
	return log, nil
}

func checkLogConfig(logConf *LogConfig) {
	if len(logConf.LogLevel) == 0 {
		logConf.LogLevel = DefaultLogLevel
	}
	if len(logConf.LogPath) == 0 {
		logConf.LogPath = DefaultLogPath
	}
	if logConf.MaxAge == 0 {
		logConf.MaxAge = DefaultLogMaxAge
	}
	if logConf.RotationTime == 0 {
		logConf.RotationTime = DefaultLogRotationTime
	}
	if logConf.RotationSize == 0 {
		logConf.RotationTime = DefaultLogRotationSize
	}
}

func getZapLevel(lvl string) (*zap.AtomicLevel, error) {
	var zapLevel zapcore.Level
	switch strings.ToUpper(lvl) {
	case ERROR:
		zapLevel = zap.ErrorLevel
	case WARN:
		zapLevel = zap.WarnLevel
	case INFO:
		zapLevel = zap.InfoLevel
	case DEBUG:
		zapLevel = zap.DebugLevel
	default:
		return nil, errors.New("invalid log level")
	}
	aLevel := zap.NewAtomicLevel()
	aLevel.SetLevel(zapLevel)

	return &aLevel, nil
}

func initLogger(logConfig *LogConfig, name string) (*zap.SugaredLogger, error) {

	checkLogConfig(logConfig)

	hook, err := getHook(logConfig.LogPath, logConfig.MaxAge, logConfig.RotationTime)
	if err != nil {
		return nil, err
	}

	level, err := getZapLevel(logConfig.LogLevel)
	if err != nil {
		level, _ = getZapLevel(DefaultLogLevel)
	}

	var syncer zapcore.WriteSyncer
	syncers := []zapcore.WriteSyncer{zapcore.AddSync(hook)}
	if logConfig.LogInConsole {
		syncers = append(syncers, zapcore.AddSync(os.Stdout))
	}

	syncer = zapcore.NewMultiWriteSyncer(syncers...)

	logger := newLogger(logConfig, level, syncer).Named(name)
	sugaredLogger := logger.Sugar()

	return sugaredLogger, nil
}

func newLogger(logConfig *LogConfig, level *zap.AtomicLevel, writeSyncer zapcore.WriteSyncer) *zap.Logger {

	var encoderConfig zapcore.EncoderConfig
	if logConfig.IsBrief {
		encoderConfig = zapcore.EncoderConfig{
			TimeKey:    "time",
			MessageKey: "msg",
			EncodeTime: CustomTimeEncoder,
			LineEnding: zapcore.DefaultLineEnding,
		}
	} else {
		encoderConfig = zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "line",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    CustomLevelEncoder,
			EncodeTime:     CustomTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeName:     zapcore.FullNameEncoder,
		}
	}

	var encoder zapcore.Encoder
	if logConfig.JsonFormat {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	core := zapcore.NewCore(
		encoder,
		writeSyncer,
		level,
	)

	l := zap.New(core)

	if logConfig.ShowLine {
		l = l.WithOptions(zap.AddCaller())
	}

	if lvl, err := getZapLevel(logConfig.StackTraceLevel); err == nil {
		l = l.WithOptions(zap.AddStacktrace(lvl))
	}

	// l = l.WithOptions(zap.AddCallerSkip(2))
	return l
}

func getHook(filename string, maxAge, rotationTime int) (io.Writer, error) {

	hook, err := rotatelogs.New(
		filename+".%Y%m%d%H",
		rotatelogs.WithRotationTime(time.Hour*time.Duration(rotationTime)),
		//filename+".%Y%m%d%H%M",
		// rotatelogs.WithRotationSize(rotationSize*ROTATION_SIZE_MB),
		rotatelogs.WithLinkName(filename),
		rotatelogs.WithMaxAge(time.Hour*24*time.Duration(maxAge)),
	)

	if err != nil {
		return nil, err
	}

	return hook, nil
}

// CustomLevelEncoder 自定义日志级别的输出格式
// @param level
// @param enc
func CustomLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + level.CapitalString() + "]")
}

// CustomTimeEncoder 自定义时间转字符串的编码方法
// @param t
// @param enc
func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}
