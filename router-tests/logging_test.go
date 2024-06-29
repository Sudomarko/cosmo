package integration_test

import (
	"os"
	"strings"
	"testing"

	"github.com/wundergraph/cosmo/router/pkg/logging"
	"go.uber.org/zap/zapcore"
)

func TestNewZapLogger(t *testing.T) {
	logFile, err := os.CreateTemp("", "test_log_file.json")
	if err != nil {
		t.Fatalf("Failed to create temp log file: %s", err)
	}
	defer os.Remove(logFile.Name())

	logger := logging.New(false, logFile, false, zapcore.InfoLevel)

	logger.Info("This is an info message")
	logger.Warn("This is a warning message")
	logger.Error("This is an error message")

	logger.Sync()

	logFileContent, err := os.ReadFile(logFile.Name())
	if err != nil {
		t.Fatalf("Failed to read log file: %s", err)
	}

	requiredMessages := []string{
		`"msg":"This is an info message"`,
		`"msg":"This is a warning message"`,
		`"msg":"This is an error message"`,
	}

	for _, msg := range requiredMessages {
		if !strings.Contains(string(logFileContent), msg) {
			t.Errorf("Log output does not contain required message: %s", msg)
		}
	}
}
