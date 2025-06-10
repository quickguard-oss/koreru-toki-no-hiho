package logger

import (
	"testing"

	"github.com/quickguard-oss/koreru-toki-no-hiho/internal/pkg/testhelper"
)

func TestMain(m *testing.M) {
	restoreLogger := testhelper.DisableLogging()

	defer restoreLogger()

	m.Run()
}
