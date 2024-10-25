package slogx

import (
	"fmt"
	"log/slog"
	"os"
)

// OpenLogFile opens a log file for writing and returns a new logger based on the provided options.
func OpenLogFile(
	path string,
	format string,
	includeTime bool,
	level string,
	verbose bool,
	logger **slog.Logger,
	writer **os.File,
) error {
	var err error
	if path == "" || *writer == os.Stderr {
		writeToStdErr()
		*writer = os.Stderr
	} else {
		err = closeExistingLogFile(writer)
		if err != nil {
			return err
		}
		*writer, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o666)
		if err != nil {
			return fmt.Errorf("[Slogger] error opening log file: %w", err)
		}
	}

	newLogger := NewLogger(Options{
		Format:      format,
		IncludeTime: includeTime,
		Level:       level,
		Verbose:     verbose,
		Writer:      *writer})
	newLogger.Info("[Slogger] slogx reopened")
	*logger = newLogger
	return nil
}

func writeToStdErr() {
	_, _ = fmt.Fprintf(os.Stderr, "[Slogger] No log file specified, using stderr\n")
}

func closeExistingLogFile(file **os.File) error {
	if file != nil {
		if err := (*file).Close(); err != nil {
			return fmt.Errorf("[Slogger] error closing log file: %w", err)
		}
	}
	return nil
}
