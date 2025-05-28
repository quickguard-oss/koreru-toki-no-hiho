package testhelper

import (
	"io"
	"log/slog"
	"testing"
)

/*
nullLogger is a logger instance that discards all log entries.
*/
var nullLogger = slog.New(
	slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}),
)

/*
DisableLogging temporarily replaces the default logger with a null logger
that discards all output. It returns a function that, when called,
restores the original logger.

Example:

	restoreLogger := testhelper.DisableLogging()

	// Test code with logging disabled

	restoreLogger()
*/
func DisableLogging() func() {
	l := slog.Default()

	slog.SetDefault(nullLogger)

	return func() {
		slog.SetDefault(l)
	}
}

/*
PreserveLogger ensures that the default logger is restored to its original state
after the test completes.

Example:

	func TestSomething(t *testing.T) {
		testhelper.PreserveLogger(t)

		// Modify logger settings for test
		slog.SetDefault(customLogger)

		// Test code runs with custom logger
		// Original logger will be restored automatically after test completes
	}
*/
func PreserveLogger(t *testing.T) {
	t.Helper()

	l := slog.Default()

	t.Cleanup(func() {
		slog.SetDefault(l)
	})
}
