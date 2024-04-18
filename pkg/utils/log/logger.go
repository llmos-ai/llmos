package log

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Logger is the interface we want for our logger, so we can plug different ones easily
type Logger interface {
	Info(string, ...interface{})
	Debug(string, ...interface{})
	Error(string, ...interface{})
	Fatal(string, ...interface{})
	SetLevel(string)
	SetDebugLevel()
	GetLogger() *slog.Logger
	GetLeveler() *slog.LevelVar
	IsDebug() bool
}

type logger struct {
	ctx      context.Context
	Logger   *slog.Logger
	logLevel *slog.LevelVar
}

func NewLogger(ctx context.Context) Logger {
	level := &slog.LevelVar{} // INFO
	opts := &slog.HandlerOptions{
		Level: level,
	}
	if viper.GetBool("debug") {
		level.Set(slog.LevelDebug)
	}

	if viper.GetBool("dev") {
		opts.AddSource = true
	}

	l := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(l)
	return &logger{
		ctx:      ctx,
		Logger:   l,
		logLevel: level,
	}
}

func (l *logger) Info(msg string, args ...interface{}) {
	l.Logger.InfoContext(l.ctx, msg, args...)
}

func (l *logger) Debug(msg string, args ...interface{}) {
	l.Logger.DebugContext(l.ctx, msg, args...)
}

func (l *logger) Error(msg string, args ...interface{}) {
	l.Logger.ErrorContext(l.ctx, msg, args...)
}

func (l *logger) Fatal(msg string, args ...interface{}) {
	l.Logger.ErrorContext(l.ctx, msg, args...)
	os.Exit(1)
}

func (l *logger) SetLevel(logLevel string) {
	level := strings.ToLower(logLevel)
	switch level {
	case "debug":
		l.logLevel.Set(slog.LevelDebug)
	case "info":
		l.logLevel.Set(slog.LevelInfo)
	case "warn":
		l.logLevel.Set(slog.LevelWarn)
	case "error":
		l.logLevel.Set(slog.LevelError)
	default:
		l.logLevel.Set(slog.LevelInfo)
	}
}

func (l *logger) SetDebugLevel() {
	l.logLevel.Set(slog.LevelDebug)
}

func (l *logger) GetLogger() *slog.Logger {
	return l.Logger
}

func (l *logger) GetLeveler() *slog.LevelVar {
	return l.logLevel
}

func (l *logger) IsDebug() bool {
	return l.logLevel.Level() == slog.LevelDebug
}
