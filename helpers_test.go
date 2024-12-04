package hop_test

import (
	"io"
	"log/slog"
)

func newTestLogger(out io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}
