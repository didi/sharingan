package logger

import (
	"os"
	"strings"

	"github.com/v2pro/plz/countlog"
)

var logFile = "STDOUT"
var logLevel = countlog.LevelInfo
var logFormat = "HumanReadableFormat"

func init() {
	if level := os.Getenv("RECORDER_LOG_LEVEL"); level != "" {
		SetLogLevel(level)
	}

	if file := os.Getenv("RECORDER_LOG_FILE"); file != "" {
		SetLogFile(file)
	}
}

// Setup 设置countlog
func Setup() {
	logWriter := countlog.NewAsyncLogWriter(logLevel, countlog.NewFileLogOutput(logFile))
	logWriter.LogFormatter = &countlog.HumanReadableFormat{
		ContextPropertyNames: []string{"threadID", "outboundSrc"},
		StringLengthCap:      512,
	}

	logWriter.Start()
	countlog.LogWriters = append(countlog.LogWriters, logWriter)
}

// SetLogFile 设置日志文件, default: STDOUT
func SetLogFile(file string) {
	logFile = file
}

// SetLogLevel 设置日志级别，TRACE, DEBUG, INFO, WARN, ERROR, FATAL, default: DEBUG
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
