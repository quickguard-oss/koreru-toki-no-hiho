/*
Package logger provides functionality for configuring and managing application logging.

This package offers utilities to set up structured logging with different log levels
using the standard library's slog package.
*/
package logger

import (
	"io"
	"log/slog"
	"os"
)

/*
output defines the default destination for log output.
By default, logs are written to standard error (os.Stderr).
*/
var output io.Writer = os.Stderr

/*
SetLogger configures the default logger with the specified log level and format.
*/
func SetLogger(isDebug bool, isJson bool) {
	var option *slog.HandlerOptions

	if isDebug {
		option = &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		}
	} else {
		option = &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}
	}

	var handler slog.Handler

	if isJson {
		handler = slog.NewJSONHandler(output, option)
	} else {
		handler = slog.NewTextHandler(output, option)
	}

	slog.SetDefault(
		slog.New(handler),
	)
}

/*
GetLogger returns the current logger instance.
*/
func GetLogger() *slog.Logger {
	return slog.Default()
}
