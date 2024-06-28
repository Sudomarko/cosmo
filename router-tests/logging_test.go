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
	// Get the golden file content
	goldenFilePath := "./testdata/logging_test_output.golden"
	goldenFileContent, err := os.ReadFile(goldenFilePath)
	if err != nil {
		t.Fatalf("Failed to read golden file: %s", err)
	}

	// Compare the first characters of both actual and expected output
	if len(logFileContent) == 0 || len(goldenFileContent) == 0 {
		t.Fatalf("Either log file content or golden file content is empty")
	}

	if logFileContent[0] != goldenFileContent[0] {
		t.Errorf("First character of log output does not match golden file")
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
		if !strings.Contains(string(goldenFileContent), msg) {
			t.Errorf("Golden file does not contain required message: %s", msg)
		}
	}
}
