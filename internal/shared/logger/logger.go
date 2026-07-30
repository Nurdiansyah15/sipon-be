package logger

import (
	"log/slog"
	"os"
)

func New(env, format string) *slog.Logger {
	var handler slog.Handler
	opts := &slog.HandlerOptions{}
	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		if format == "json" {
			handler = slog.NewJSONHandler(os.Stdout, opts)
		} else {
			handler = slog.NewTextHandler(os.Stdout, opts)
		}
	}
	return slog.New(handler)
}
