package logger

import (
	"context"
	"encoding/hex"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	ReqLogger = "reqLogger"
)

// Config 日志配置
type Config struct {
	Level      string         `json:"level"`
	File       string         `json:"file"`
	AccessFile string         `json:"access_file"`
	Rotation   RotationConfig `json:"rotation"` // 日志轮转配置
}

// RotationConfig 日志轮转配置
type RotationConfig struct {
	// MaxSize 单个日志文件最大大小（MB），默认 100MB
	MaxSize int `json:"max_size"`
	// MaxBackups 保留的旧日志文件最大数量，0表示保留所有，默认 10
	MaxBackups int `json:"max_backups"`
	// MaxAge 保留的旧日志文件最大天数，0表示不根据时间删除，默认 30
	MaxAge int `json:"max_age"`
}

func New(conf Config) (*zap.Logger, error) {
	if conf.Level == "" {
		conf.Level = "info"
	}
	
	ecfg := zap.NewProductionEncoderConfig()
	ecfg.EncodeTime = zapcore.ISO8601TimeEncoder
	ecfg.EncodeLevel = zapcore.CapitalLevelEncoder
	encoder := zapcore.NewConsoleEncoder(ecfg)
	level := zap.NewAtomicLevelAt(logLevel(conf.Level))

	// 创建 WriteSyncer
	var ws zapcore.WriteSyncer
	if conf.File == "" || conf.File == "stdout" {
		ws = zapcore.AddSync(os.Stdout)
	} else {
		// 设置默认的日志轮转配置
		setDefaultRotation(&conf.Rotation)
		lumberLogger := &lumberjack.Logger{
			Filename:   conf.File,
			MaxSize:    conf.Rotation.MaxSize,
			MaxBackups: conf.Rotation.MaxBackups,
			MaxAge:     conf.Rotation.MaxAge,
			LocalTime:  true,
		}
		ws = zapcore.AddSync(lumberLogger)
	}
	// 构建 logger
	core := zapcore.NewCore(encoder, ws, level)
	logger := zap.New(core, zap.AddStacktrace(zap.PanicLevel), zap.AddCaller())

	zap.ReplaceGlobals(logger)
	return logger, nil
}

// setDefaultRotation 设置默认的日志轮转配置
func setDefaultRotation(rotation *RotationConfig) {
	if rotation.MaxSize == 0 {
		rotation.MaxSize = 100 // 默认 100MB
	}
	if rotation.MaxBackups == 0 {
		rotation.MaxBackups = 10 // 默认保留 10 个备份
	}
	if rotation.MaxAge == 0 {
		rotation.MaxAge = 30 // 默认保留 30 天
	}
}

func logLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

type loggerKey struct{}

var LoggerKey = loggerKey{}

type Logger struct {
	ReqID string
	*zap.SugaredLogger
}

func NewLogger(reqID string) *Logger {
	log := zap.S().Named(reqID)
	return &Logger{
		ReqID:         reqID,
		SugaredLogger: log,
	}
}

func NewContext(reqID string) context.Context {
	return context.WithValue(context.Background(), LoggerKey, NewLogger(reqID))
}

func FromGinContext(c *gin.Context) *Logger {
	// ReqLogger 在 gin 上下文中一定存在
	return c.MustGet(ReqLogger).(*Logger)
}

func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(LoggerKey).(*Logger); ok {
		return logger
	}
	uid := uuid.New()
	reqID := hex.EncodeToString(uid[:])
	return &Logger{
		ReqID:         reqID,
		SugaredLogger: zap.S().Named(reqID),
	}
}
