package log

import (
	"io"
	"log/slog"
	"time"

	"github.com/lmittmann/tint"
)

// LevelFromString returns the slog.Level corresponding to the given level string.
// If the level string is not recognized, slog.LevelInfo is returned.
func LevelFromString(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Options represents the configuration options for a logger.
//
// Format defines the format of log output. Valid values are "pretty", "text", and "json".
//
// IncludeTime indicates whether to include timestamps in log output.
//
// Level specifies the log level. Valid values are "debug", "info", "warn", and "error".
//
// Verbose indicates whether to include source information in log output.
//
// Writer is the io.Writer to which log output will be written. This lets us redirect the output to a file, for example.
type Options struct {
	Format      string
	IncludeTime bool
	Level       string
	Verbose     bool
	Writer      io.Writer
}

// NewLogger creates a new slog.Logger based on the provided options.
//
// Format can be one of "pretty", "text", or "json". Defaults to "json".
// IncludeTime indicates whether to include timestamps in log output. Defaults to false.
// Level can be one of "debug", "info", "warn", or "error". Defaults to "info".
// Verbose indicates whether to include source information in log output. Defaults to false.
func NewLogger(opts Options) *slog.Logger {
	var replaceAttr func(groups []string, a slog.Attr) slog.Attr

	if opts.IncludeTime {
		replaceAttr = nil
	} else {
		replaceAttr = removeTimeAttr
	}

	if opts.Format == "pretty" {
		return slog.New(tint.NewHandler(opts.Writer,
			&tint.Options{
				AddSource:   opts.Verbose,
				Level:       LevelFromString(opts.Level),
				ReplaceAttr: replaceAttr,
				TimeFormat:  time.Kitchen,
			}))
	}

	if opts.Format == "json" {
		return slog.New(slog.NewJSONHandler(opts.Writer,
			&slog.HandlerOptions{
				AddSource:   opts.Verbose,
				Level:       LevelFromString(opts.Level),
				ReplaceAttr: replaceAttr,
			}))
	}

	return slog.New(slog.NewTextHandler(opts.Writer,
		&slog.HandlerOptions{
			AddSource:   opts.Verbose,
			Level:       LevelFromString(opts.Level),
			ReplaceAttr: replaceAttr,
		}))
}

// removeTimeAttr removes the timestamp attribute from logs.
func removeTimeAttr(_ []string, a slog.Attr) slog.Attr {
	// Remove timestamp from logs
	if a.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return a
}
