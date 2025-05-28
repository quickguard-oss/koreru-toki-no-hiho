package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_validateStackPrefix(t *testing.T) {
	testCases := []struct {
		name     string
		prefix   string
		expected bool
	}{
		{
			name:     "Empty",
			prefix:   "",
			expected: false,
		},
		{
			name:     "Within 10 chars",
			prefix:   "Abc123",
			expected: true,
		},
		{
			name:     "Exactly 10 chars",
			prefix:   "Abcde12345",
			expected: true,
		},
		{
			name:     "11 chars (too long)",
			prefix:   "Abcde123456",
			expected: false,
		},
		{
			name:     "Contains special chars",
			prefix:   "Abc-123",
			expected: false,
		},
		{
			name:     "Contains non-ASCII chars",
			prefix:   "Abc日本語",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalStackPrefixFlag := stackPrefixFlag

			t.Cleanup(func() {
				stackPrefixFlag = originalStackPrefixFlag
			})

			stackPrefixFlag = tc.prefix

			err := validateStackPrefix()

			if tc.expected {
				assert.NoError(t, err, "Prefix '%s' should be valid", tc.prefix)
			} else {
				assert.Error(t, err, "Prefix '%s' should be invalid", tc.prefix)
			}
		})
	}
}

func Test_validateWaitTimeout(t *testing.T) {
	testCases := []struct {
		name        string
		waitTimeout time.Duration
		expected    bool
	}{
		{
			name:        "Positive timeout",
			waitTimeout: 5 * time.Minute,
			expected:    true,
		},
		{
			name:        "Zero timeout",
			waitTimeout: 0,
			expected:    false,
		},
		{
			name:        "Negative timeout",
			waitTimeout: -5 * time.Minute,
			expected:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalWaitTimeoutFlag := waitTimeoutFlag

			t.Cleanup(func() {
				waitTimeoutFlag = originalWaitTimeoutFlag
			})

			waitTimeoutFlag = tc.waitTimeout

			err := validateWaitTimeout()

			if tc.expected {
				assert.NoError(t, err, "timeout '%v' should be valid", tc.waitTimeout)
			} else {
				assert.Error(t, err, "timeout '%v' should be invalid", tc.waitTimeout)
			}
		})
	}
}

func Test_timeoutDuration(t *testing.T) {
	testCases := []struct {
		name        string
		noWait      bool
		waitTimeout time.Duration
		expected    time.Duration
	}{
		{
			name:        "noWait=true",
			noWait:      true,
			waitTimeout: 5 * time.Minute,
			expected:    0,
		},
		{
			name:        "noWait=false",
			noWait:      false,
			waitTimeout: 5 * time.Minute,
			expected:    5 * time.Minute,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalNoWaitFlag := noWaitFlag
			originalWaitTimeoutFlag := waitTimeoutFlag

			t.Cleanup(func() {
				noWaitFlag = originalNoWaitFlag
				waitTimeoutFlag = originalWaitTimeoutFlag
			})

			noWaitFlag = tc.noWait
			waitTimeoutFlag = tc.waitTimeout

			assert.Equal(t, tc.expected, timeoutDuration(), "timeoutDuration() should return %v", tc.expected)
		})
	}
}
