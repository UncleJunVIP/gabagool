package ui

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func InitLogger(logFilePath string) error {
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	handler := slog.NewTextHandler(logFile, opts)
	Logger = slog.New(handler)

	return nil
}
