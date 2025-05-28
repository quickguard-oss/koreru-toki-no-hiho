package logger

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/testhelper"
)

func Test_SetLogger(t *testing.T) {
	testCases := []struct {
		name     string
		isDebug  bool
		isJson   bool
		expected slog.Level
	}{
		{
			name:     "Debug level with text format",
			isDebug:  true,
			isJson:   false,
			expected: slog.LevelDebug,
		},
		{
			name:     "Debug level with JSON format",
			isDebug:  true,
			isJson:   true,
			expected: slog.LevelDebug,
		},
		{
			name:     "Info level with text format",
			isDebug:  false,
			isJson:   false,
			expected: slog.LevelInfo,
		},
		{
			name:     "Info level with JSON format",
			isDebug:  false,
			isJson:   true,
			expected: slog.LevelInfo,
		},
	}

	checkLevels := []slog.Level{
		slog.LevelDebug,
		slog.LevelInfo,
		slog.LevelWarn,
		slog.LevelError,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testhelper.PreserveLogger(t)

			SetLogger(tc.isDebug, tc.isJson)

			logger := GetLogger()

			t.Run("Level", func(t *testing.T) {
				enabled := []bool{
					tc.expected <= slog.LevelDebug,
					tc.expected <= slog.LevelInfo,
					tc.expected <= slog.LevelWarn,
					tc.expected <= slog.LevelError,
				}

				for i, checkLevel := range checkLevels {
					got := logger.Enabled(context.TODO(), checkLevel)

					assert.Equal(t, enabled[i], got,
						"Logger with level '%v' should have enabled=%v for level '%v'", tc.expected, enabled[i], checkLevel,
					)
				}
			})

			t.Run("Format", func(t *testing.T) {
				handler := logger.Handler()

				if tc.isJson {
					assert.IsType(t, &slog.JSONHandler{}, handler, "Logger should be of type JSONHandler")
				} else {
					assert.IsType(t, &slog.TextHandler{}, handler, "Logger should be of type TextHandler")
				}
			})
		})
	}
}
