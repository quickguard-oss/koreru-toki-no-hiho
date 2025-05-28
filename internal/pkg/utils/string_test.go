package utils

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GenerateRandomStr(t *testing.T) {
	t.Run("Not empty", func(t *testing.T) {
		randomStr := GenerateRandomStr()

		assert.NotEmpty(t, randomStr, "Random string should not be empty")
	})

	t.Run("Min length", func(t *testing.T) {
		randomStr := GenerateRandomStr()

		assert.GreaterOrEqual(t, len(randomStr), 10, "Random string should be around 10 characters or more")
	})

	t.Run("Only alphanumeric", func(t *testing.T) {
		randomStr := GenerateRandomStr()

		assert.Regexp(t, regexp.MustCompile("^[A-Za-z0-9]+$"), randomStr, "Random string should only contain alphanumeric characters")
	})

	t.Run("Unique values", func(t *testing.T) {
		randomStr := GenerateRandomStr()
		anotherRandomStr := GenerateRandomStr()

		assert.NotEqual(t, randomStr, anotherRandomStr, "Two random strings should be different (although there is an extremely low probability they could match)")
	})
}

func Test_Truncate(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		length   int
		expected string
	}{
		{
			name:     "Shorter than limit",
			input:    "123",
			length:   10,
			expected: "123",
		},
		{
			name:     "Equal to limit",
			input:    "12345",
			length:   5,
			expected: "12345",
		},
		{
			name:     "Longer than limit",
			input:    "123456789",
			length:   3,
			expected: "123",
		},
		{
			name:     "Zero length",
			input:    "12345",
			length:   0,
			expected: "",
		},
		{
			name:     "Empty string",
			input:    "",
			length:   5,
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, Truncate(tc.input, tc.length), "Truncation result does not match expected value")
		})
	}
}
