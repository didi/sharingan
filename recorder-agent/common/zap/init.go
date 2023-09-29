package zap

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/didi/sharingan/recorder-agent/common/conf"
	"github.com/didi/sharingan/recorder-agent/common/path"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger zapLog
var Logger *zap.Logger

func init() {
	var err error

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
	ws := zapcore.AddSync(rlogs)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.LevelKey = ""
	encoderConfig.TimeKey = ""

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		ws,
		zap.InfoLevel,
	)

	Logger = zap.New(core)
}

func Format(ctx context.Context, level string, format string, args ...interface{}) string {
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

	var ctxString string
	if t, ok := ctx.Value(tracerKey).(Tracer); ok {
		ctxString = t.Format()
	}

	return fmt.Sprintf("[%s][%s][%s:%d] %s%s", level, ts, file, line, ctxString, fmt.Sprintf(format, args...))
}
