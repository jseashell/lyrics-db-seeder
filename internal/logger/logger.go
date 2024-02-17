// Copyright 2024 John Schellinger.
// Use of this file is governed by the MIT license that can
// be found in the LICENSE.txt file in the project root.

// Package `logger` contains configuration for structured logging.
package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

func New() *slog.Logger {
	envLevel := os.Getenv("LOG_LEVEL")
	var slogLevel slog.Level

	switch envLevel {
	case "DEBUG":
		slogLevel = slog.LevelDebug
		break
	case "WARN":
		slogLevel = slog.LevelWarn
		break
	case "ERROR":
		slogLevel = slog.LevelError
	default: // INFO
		slogLevel = slog.LevelInfo
	}

	os.Mkdir("logs", os.ModePerm)
	file, _ := os.Create(fmt.Sprintf("logs/%s.log", time.Now().UTC().Format(time.RFC3339)))

	mw := io.MultiWriter(os.Stdout, file)
	logger := slog.New(slog.NewJSONHandler(mw, &slog.HandlerOptions{
		Level: slogLevel,
	}))

	return logger
}

func Close() {}
