package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewLoggersNoPath(t *testing.T) {
	t.Parallel()

	loggers, err := NewLoggers(&DatasourceSettingsJSONDataPartial{LogLevel: 0, LogPath: ""}, "ds-1")
	if err != nil {
		t.Fatalf("expected no error without a log path: %v", err)
	}

	if loggers.fileLogger != nil {
		t.Fatal("expected no file logger when no log path is configured")
	}

	if loggers.fileClose != nil {
		t.Fatal("expected no close function when no log path is configured")
	}

	// Close must be a safe no-op when no file is configured.
	loggers.Close()
}

func TestNewLoggersWritablePath(t *testing.T) {
	t.Parallel()

	logPath := filepath.Join(t.TempDir(), "plugin.log")

	loggers, err := NewLoggers(&DatasourceSettingsJSONDataPartial{
		LogLevel: 7,
		LogPath:  logPath,
	}, "ds-1")
	if err != nil {
		t.Fatalf("expected logger creation to succeed: %v", err)
	}

	if loggers.fileLogger == nil {
		t.Fatal("expected a file logger when a log path is configured")
	}

	if loggers.fileClose == nil {
		t.Fatal("expected a close function when a log path is configured")
	}

	loggers.debugf("lifecycle test message")

	// Closing must flush and close the log file without error.
	loggers.Close()
	loggers.Close()

	_, statErr := os.Stat(logPath)
	if statErr != nil {
		t.Fatalf("expected log file to exist: %v", statErr)
	}
}

func TestNewLoggersParentIsFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	parentAsFile := filepath.Join(tmpDir, "not-a-dir")

	err := os.WriteFile(parentAsFile, []byte("x"), 0o600)
	if err != nil {
		t.Fatalf("failed to create parent file: %v", err)
	}

	logPath := filepath.Join(parentAsFile, "plugin.log")

	_, err = NewLoggers(&DatasourceSettingsJSONDataPartial{
		LogLevel: 7,
		LogPath:  logPath,
	}, "ds-1")
	if err == nil {
		t.Fatal("expected error when the log path parent is a file")
	}

	if !strings.Contains(err.Error(), logPath) {
		t.Fatalf("error %q should contain the log path %q", err.Error(), logPath)
	}
}

func TestNewLoggersPathIsDirectory(t *testing.T) {
	t.Parallel()

	logPath := t.TempDir()

	_, err := NewLoggers(&DatasourceSettingsJSONDataPartial{
		LogLevel: 7,
		LogPath:  logPath,
	}, "ds-1")
	if err == nil {
		t.Fatal("expected error when the log path is a directory")
	}

	if !strings.Contains(err.Error(), logPath) {
		t.Fatalf("error %q should contain the log path %q", err.Error(), logPath)
	}
}
