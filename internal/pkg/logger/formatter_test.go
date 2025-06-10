package logger

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FormatAsTable(t *testing.T) {
	tests := []struct {
		name     string
		headers  []string
		body     [][]string
		expected string
	}{
		{
			name:    "Padding and alignment",
			headers: []string{"Short", "Longer Header"},
			body: [][]string{
				{"A", "Some text"},
				{"Longer text", "B"},
			},
			expected: strings.Join([]string{
				"SHORT         LONGER HEADER",
				"A             Some text    ",
				"Longer text   B            ",
			}, "\n"),
		},
		{
			name:    "Empty body",
			headers: []string{"Col1", "Col2"},
			body:    [][]string{},
			expected: strings.Join([]string{
				"COL1   COL2",
			}, "\n"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatAsTable(tc.headers, tc.body)

			assert.Equal(t, tc.expected, got, "Result does not match expected output")
		})
	}
}

func Test_FormatAsJSON(t *testing.T) {
	tests := []struct {
		name      string
		attrs     []string
		body      [][]string
		expected  string
		wantError bool
	}{
		{
			name:  "Basic JSON",
			attrs: []string{"name", "age", "role"},
			body: [][]string{
				{"Alice", "25", "Engineer"},
				{"Bob", "30", "Manager"},
			},
			expected:  `[{"age":"25","name":"Alice","role":"Engineer"},{"age":"30","name":"Bob","role":"Manager"}]`,
			wantError: false,
		},
		{
			name:      "Empty body",
			attrs:     []string{"col1", "col2"},
			body:      [][]string{},
			expected:  "[]",
			wantError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := FormatAsJSON(tc.attrs, tc.body)

			if tc.wantError {
				assert.Error(t, err, "Expected an error to be returned")
			} else {
				assert.NoError(t, err, "Unexpected error occurred")

				assert.Equal(t, tc.expected, got, "Result does not match expected output")
			}
		})
	}
}
