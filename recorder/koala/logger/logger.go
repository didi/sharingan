package logger

import (
	"os"
	"strings"

	"github.com/v2pro/plz/countlog"
	"github.com/v2pro/plz/countlog/output"
	"github.com/v2pro/plz/countlog/output/async"
	"github.com/v2pro/plz/countlog/output/json"
	"github.com/v2pro/plz/countlog/output/lumberjack"
)

var logFile = "STDOUT"
var logLevel = countlog.LevelInfo
var logFormat = "HumanReadableFormat"

// Init int logger
func Init() {
	if level := os.Getenv("RECORDER_LOG_LEVEL"); level != "" {
		SetLogLevel(level)
	}

	if file := os.Getenv("RECORDER_LOG_FILE"); file != "" {
		SetLogFile(file)
	}

	writer := async.NewAsyncWriter(async.AsyncWriterConfig{
		Writer: &lumberjack.Logger{
			Filename: logFile,
		},
	})
	defer writer.Close()

	countlog.EventWriter = output.NewEventWriter(output.EventWriterConfig{
		Format: &json.JsonFormat{},
		Writer: writer,
	})
}

// SetLogFile set log file, default: STDOUT
func SetLogFile(file string) {
	logFile = file
}

// SetLogLevel set log level，TRACE, DEBUG, INFO, WARN, ERROR, FATAL, default: DEBUG
func SetLogLevel(level string) {
	level = strings.ToUpper(level)

	switch level {
	case "TRACE":
		logLevel = countlog.LevelTrace
	case "DEBUG":
		logLevel = countlog.LevelDebug
	case "INFO":
		logLevel = countlog.LevelInfo
	case "WARN":
		logLevel = countlog.LevelWarn
	case "ERROR":
		logLevel = countlog.LevelError
	case "FATAL":
		logLevel = countlog.LevelFatal
	}
}
