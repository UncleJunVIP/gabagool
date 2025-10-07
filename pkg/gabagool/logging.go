package gabagool

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	logger      *slog.Logger
	levelVar    *slog.LevelVar
	logFile     *os.File
	loggerOnce  sync.Once
	logFilename = "app.log"
)

func setLogFilename(filename string) {
	logFilename = filename
}

func GetLoggerInstance() *slog.Logger {
	loggerOnce.Do(func() {
		levelVar = &slog.LevelVar{}
		levelVar.Set(slog.LevelDebug) // default level

		if err := os.MkdirAll("logs", 0755); err != nil {
			panic("Failed to create logs directory: " + err.Error())
		}

		var err error
		logFile, err = os.OpenFile(filepath.Join("logs", logFilename), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic("Failed to open log file: " + err.Error())
		}

		multiWriter := io.MultiWriter(os.Stdout, logFile)

		handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
			Level:     levelVar,
			AddSource: false,
		})
		logger = slog.New(handler)
	})
	return logger
}

func SetLogLevel(level slog.Level) {
	GetLoggerInstance()
	levelVar.Set(level)
}

func SetRawLogLevel(rawLevel string) {
	var level slog.Level

	switch strings.ToLower(rawLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	GetLoggerInstance()
	levelVar.Set(level)
}

func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}
