package slogger

import "log/slog"

// HoneybadgerLogger is a type that represents a slog with the capability to log messages using the Honeybadger log format.
type HoneybadgerLogger struct {
	Logger *slog.Logger
}

// Printf logs a formatted message using the HoneybadgerLogger's underlying slogger.
func (l HoneybadgerLogger) Printf(format string, v ...interface{}) {
	l.Logger.Error(format, v...)
}
