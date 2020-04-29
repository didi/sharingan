package tlog

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/didi/sharingan/replayer-agent/common/handlers/conf"
	"github.com/didi/sharingan/replayer-agent/common/handlers/path"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// DebugTag ...
const DebugTag = " fdebug"

// DLTagUndefined ...
const DLTagUndefined = " _undef"

// ILog log interface
type ILog interface {
	Debugf(ctx context.Context, tag string, format string, args ...interface{})
	Infof(ctx context.Context, tag string, format string, args ...interface{})
	Warnf(ctx context.Context, tag string, format string, args ...interface{})
	Errorf(ctx context.Context, tag string, format string, args ...interface{})
	Fatalf(ctx context.Context, tag string, format string, args ...interface{})
}

// Handler zapLog
var Handler ILog

// Init Init
func Init() {
	var err error

	// 读取配置
	filename := conf.Handler.GetString("log.filename") // "log/record.log.%Y%m%d%H"
	linkname := conf.Handler.GetString("log.linkname") // , "log/record.log"
	maxHourAge := time.Duration(conf.Handler.GetInt("log.maxHourAge"))
	maxHourRotate := time.Duration(conf.Handler.GetInt("log.maxHourRotate"))

	// 文件拆分规则
	rlogs, err := rotatelogs.New(
		path.Root+"/"+filename,
		rotatelogs.WithLinkName(path.Root+"/"+linkname),
		rotatelogs.WithMaxAge(maxHourAge*time.Hour),
		rotatelogs.WithRotationTime(maxHourRotate*time.Hour),
	)

	if err != nil {
		log.Fatal("Init zap log error: ", err)
	}

	// 文件内容格式和之前保持一致
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.LevelKey = ""
	encoderConfig.TimeKey = ""

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(rlogs),
		getLogLevel(conf.Handler.GetString("log.level")),
	)

	Handler = NewTLog(zap.New(core))
}

func getLogLevel(level string) zapcore.Level {
	logLevel := zap.DebugLevel

	level = strings.ToUpper(level)
	switch level {
	case "DEBUG":
		logLevel = zap.DebugLevel
	case "INFO":
		logLevel = zap.InfoLevel
	case "WARN":
		logLevel = zap.WarnLevel
	case "ERROR":
		logLevel = zap.ErrorLevel
	case "FATAL":
		logLevel = zap.FatalLevel
	}

	return logLevel
}

// TLog ...
type TLog struct {
	log *zap.Logger
}

// NewTLog ...
func NewTLog(log *zap.Logger) *TLog {
	return &TLog{log: log}
}

// Debugf ...
func (tl *TLog) Debugf(ctx context.Context, tag string, format string, args ...interface{}) {
	tl.log.Debug(tl.format(ctx, "DEBUG", format, args...))
}

// Infof ...
func (tl *TLog) Infof(ctx context.Context, tag string, format string, args ...interface{}) {
	tl.log.Info(tl.format(ctx, "INFO", format, args...))
}

// Warnf ...
func (tl *TLog) Warnf(ctx context.Context, tag string, format string, args ...interface{}) {
	tl.log.Warn(tl.format(ctx, "WARN", format, args...))
}

// Errorf ...
func (tl *TLog) Errorf(ctx context.Context, tag string, format string, args ...interface{}) {
	tl.log.Error(tl.format(ctx, "ERROR", format, args...))
}

// Fatalf ...
func (tl *TLog) Fatalf(ctx context.Context, tag string, format string, args ...interface{}) {
	tl.log.Error(tl.format(ctx, "FATAL", format, args...))
}

func (tl *TLog) format(ctx context.Context, level string, format string, args ...interface{}) string {
	// time
	ts := time.Now().Format("2006-01-02 15:04:05.000000")

	// file, line
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "file"
		line = -1
	}

	// igonre dir
	file = strings.TrimPrefix(file, path.Root+"/")

	return fmt.Sprintf("[%s][%s][%s:%d] %s", level, ts, file, line, fmt.Sprintf(format, args...))
}
