package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func SetupLogger(env string, logDirPath string) (*slog.Logger, *os.File) {
	var log *slog.Logger
	filename := time.Now().Format("2006-01-02") + ".log"
	err := os.MkdirAll(logDirPath, 0777)
	if err != nil {
		fmt.Println("failed to create log directory: " + err.Error())
	}

	fp := filepath.Clean(logDirPath + "/" + filename)
	file, err := os.OpenFile(fp, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		panic("failed to open log file: " + err.Error())
	}
	writer := io.MultiWriter(os.Stdout, file)

	var lvl slog.Leveler
	switch env {
	case envLocal:
		lvl = slog.LevelDebug
	case envDev:
		lvl = slog.LevelInfo
	case envProd:
		lvl = slog.LevelWarn
	}
	log = slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level: lvl,
	}))
	return log, file
}
