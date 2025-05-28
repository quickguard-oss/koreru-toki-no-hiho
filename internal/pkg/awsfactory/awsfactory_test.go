package awsfactory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_loadAWSConfig(t *testing.T) {
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "dummy-key-id")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "dummy-secret-key")

	resetConfiguration()

	assert.Equal(t, 0, counter, "Counter should start at 0")

	err := loadAWSConfig()

	assert.NoError(t, err, "Should not return error when loading AWS config")
	assert.Equal(t, 1, counter, "Counter should be incremented to 1")

	err = loadAWSConfig()

	assert.NoError(t, err, "Should not return error when loading AWS config again")
	assert.Equal(t, 1, counter, "Counter should still be 1 (config loaded only once)")

	resetConfiguration()

	assert.Equal(t, 0, counter, "Counter should be reset to 0")
}
