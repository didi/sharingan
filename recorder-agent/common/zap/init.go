package zap

import (
	"log"
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
