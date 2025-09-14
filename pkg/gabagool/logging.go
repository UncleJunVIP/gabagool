package gabagool

import (
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/atomic"
)

var logFile *os.File
var logger atomic.Pointer[slog.Logger]
var currentLevel atomic.Pointer[slog.LevelVar]

var LoggerInitialized atomic.Bool

var onceLogger sync.Once

func LogStandardFatal(msg string, err error) {
	log.SetOutput(os.Stderr)
	log.Fatalf("%s: %v", msg, err)
}

func GetLoggerInstance() *slog.Logger {
	onceLogger.Do(func() {
		logger.Store(createLogger())
	})
	return logger.Load()
}

func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

func createLogger() *slog.Logger {
	LoggerInitialized.Store(false)

	exePath, err := os.Executable()
	if err != nil {
		LogStandardFatal("Couldn't determine executable path: %v", err)
	}

	logFileName := filepath.Base(exePath) + ".log"

	cwd, err := os.Getwd()
	if err != nil {
		LogStandardFatal("Failed to get current working directory", err)
	}

	logFile, err = os.OpenFile(filepath.Join(cwd, logFileName), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		LogStandardFatal("Unable to open log file!", err)
	}

	// Create a level variable for dynamic level changes
	levelVar := &slog.LevelVar{}
	levelVar.Set(slog.LevelInfo) // Default level
	currentLevel.Store(levelVar)

	// Create a multi-writer to write to both console and file
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Create a clean JSON handler
	handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level:     levelVar,
		AddSource: false,
	})

	LoggerInitialized.Store(true)

	return slog.New(handler)
}

func SetLogLevel(rawLevel string) {
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

	// Update the level variable to change the logging level dynamically
	if levelVar := currentLevel.Load(); levelVar != nil {
		levelVar.Set(level)
	}
}
