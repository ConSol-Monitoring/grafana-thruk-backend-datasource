package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateFileLoggerNoPath(t *testing.T) {
	t.Parallel()

	logger, closeFn, err := createFileLogger(&DatasourceSettingsJSONDataPartial{LogLevel: 0, LogPath: ""})
	if err != nil {
		t.Fatalf("expected no error without a log path: %v", err)
	}

	if logger != nil {
		t.Fatal("expected no file logger when no log path is configured")
	}

	if closeFn != nil {
		t.Fatal("expected no close function when no log path is configured")
	}
}

func TestCreateFileLoggerWritablePath(t *testing.T) {
	t.Parallel()

	logPath := filepath.Join(t.TempDir(), "plugin.log")

	logger, closeFn, err := createFileLogger(&DatasourceSettingsJSONDataPartial{
		LogLevel: 7,
		LogPath:  logPath,
	})
	if err != nil {
		t.Fatalf("expected logger creation to succeed: %v", err)
	}

	if logger == nil {
		t.Fatal("expected a file logger when a log path is configured")
	}

	if closeFn == nil {
		t.Fatal("expected a close function when a log path is configured")
	}

	logger.Debug("lifecycle test message")

	// Closing must flush and close the log file without error.
	closeFn()
	closeFn()

	_, statErr := os.Stat(logPath)
	if statErr != nil {
		t.Fatalf("expected log file to exist: %v", statErr)
	}
}

func TestCreateFileLoggerParentIsFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	parentAsFile := filepath.Join(tmpDir, "not-a-dir")

	err := os.WriteFile(parentAsFile, []byte("x"), 0o600)
	if err != nil {
		t.Fatalf("failed to create parent file: %v", err)
	}

	logPath := filepath.Join(parentAsFile, "plugin.log")

	_, _, err = createFileLogger(&DatasourceSettingsJSONDataPartial{
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

func TestCreateFileLoggerPathIsDirectory(t *testing.T) {
	t.Parallel()

	logPath := t.TempDir()

	_, _, err := createFileLogger(&DatasourceSettingsJSONDataPartial{
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
