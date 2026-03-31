package logging

import (
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

type Logger struct {
	inner *slog.Logger
}

type Config struct {
	Level slog.Level

	Format Format

	LogToStdout bool

	FilePath   string
	MaxSizeMB  int
	MaxAgeDays int
	MaxBackups int
	LocalTime  bool
	Compress   bool
}

func New(cfg Config) Logger {
	var writer io.Writer

	cfg = cfg.withDefaults()

	switch {
	case cfg.LogToStdout || cfg.FilePath == "":
		writer = os.Stdout
	default:
		writer = &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSizeMB,
			MaxAge:     cfg.MaxAgeDays,
			MaxBackups: cfg.MaxBackups,
			LocalTime:  cfg.LocalTime,
			Compress:   cfg.Compress,
		}
	}

	opts := &slog.HandlerOptions{
		Level: cfg.Level,
	}

	var handler slog.Handler
	switch cfg.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(writer, opts)
	default:
		handler = slog.NewTextHandler(writer, opts)
	}

	return Logger{
		inner: slog.New(handler),
	}
}

func (c Config) withDefaults() Config {
	if c.Level == 0 {
		c.Level = slog.LevelInfo
	}

	if c.Format == "" {
		c.Format = FormatText
	}

	if !c.LogToStdout && c.FilePath == "" {
		c.LogToStdout = true
	}

	if c.MaxSizeMB <= 0 {
		c.MaxSizeMB = 10
	}

	if c.MaxAgeDays <= 0 {
		c.MaxAgeDays = 30
	}

	if c.MaxBackups <= 0 {
		c.MaxBackups = 3
	}

	return c
}

func (l *Logger) Info(msg string, args ...any) {
	baseArgs := []any{
		"component", "fail2cloudflare",
	}

	l.inner.Info(msg, append(baseArgs, args...)...)
}

func (l *Logger) Error(msg string, err error, args ...any) {
	baseArgs := []any{
		"component", "fail2cloudflare",
	}

	if err != nil {
		baseArgs = append(baseArgs, "error", err)
	}

	l.inner.Error(msg, append(baseArgs, args...)...)
}
