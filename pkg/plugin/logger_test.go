package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateLoggerWritablePath(t *testing.T) {
	t.Parallel()

	logPath := filepath.Join(t.TempDir(), "plugin.log")

	logger, err := createLoggerFromDatasourceSettings(&DatasourceSettingsJSONDataPartial{
		LogLevel: 7,
		LogPath:  logPath,
	})
	if err != nil {
		t.Fatalf("expected logger creation to succeed: %v", err)
	}

	_ = logger.Sync()
}

func TestCreateLoggerParentIsFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	parentAsFile := filepath.Join(tmpDir, "not-a-dir")

	err := os.WriteFile(parentAsFile, []byte("x"), 0o600)
	if err != nil {
		t.Fatalf("failed to create parent file: %v", err)
	}

	logPath := filepath.Join(parentAsFile, "plugin.log")

	_, err = createLoggerFromDatasourceSettings(&DatasourceSettingsJSONDataPartial{
		LogLevel: 7,
		LogPath:  logPath,
	})
	if err == nil {
		t.Fatal("expected error when the log path parent is a file")
	}

	if !strings.Contains(err.Error(), logPath) {
		t.Fatalf("error %q should contain the log path %q", err.Error(), logPath)
	}
}

func TestCreateLoggerPathIsDirectory(t *testing.T) {
	t.Parallel()

	logPath := t.TempDir()

	_, err := createLoggerFromDatasourceSettings(&DatasourceSettingsJSONDataPartial{
		LogLevel: 7,
		LogPath:  logPath,
	})
	if err == nil {
		t.Fatal("expected error when the log path is a directory")
	}

	if !strings.Contains(err.Error(), logPath) {
		t.Fatalf("error %q should contain the log path %q", err.Error(), logPath)
	}
}
