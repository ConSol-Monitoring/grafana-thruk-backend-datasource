package plugin

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	zaplogfmt "github.com/jsternberg/zap-logfmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ErrInvalidLogLevel error type.
var ErrInvalidLogLevel = errors.New("invalid log level, has to be between [0-7]")

// ErrCouldNotBuildLogger error type.
var ErrCouldNotBuildLogger = errors.New("could not build logger from the configuration")

const (
	// logMaxSizeMB is the maximum log file size in megabytes before rotation.
	logMaxSizeMB = 100
	// logMaxBackups is the number of rotated files to retain.
	logMaxBackups = 3
	// logMaxAgeDays is how many days to retain rotated files.
	logMaxAgeDays = 28
)

// createFileLogger builds an optional Zap logger that writes to the configured log file.
// It returns a nil logger and a nil close function when no log path is configured, in which
// case all logging goes through Grafana's own pipeline via the SDK logger instead.
//
//nolint:funlen // expands the path, creates directories and configures rotation before building the logger
func createFileLogger(jsonData *DatasourceSettingsJSONDataPartial) (*zap.SugaredLogger, func(), error) {
	if jsonData == nil {
		return nil, nil, fmt.Errorf("%w , argument: jsonData", ErrArgumentNil)
	}

	// Only write a log file when a path is explicitly configured.
	if jsonData.LogPath == "" {
		return nil, nil, nil
	}

	var logLevel zapcore.Level

	// imitiate syslog(3) log levels
	// emerg, alert, crit, err, warning, notice, info, debug
	// 0    , 1    , 2   , 3  , 4      , 5     , 6   , 7

	//nolint:mnd
	switch jsonData.LogLevel {
	case 0:
		logLevel = zapcore.FatalLevel
	case 1, 2, 3:
		logLevel = zapcore.PanicLevel
	case 4:
		logLevel = zapcore.WarnLevel
	case 5, 6:
		logLevel = zapcore.InfoLevel
	case 7:
		logLevel = zapcore.DebugLevel
	default:
		return nil, nil, fmt.Errorf("%w logLevel: %d", ErrInvalidLogLevel, jsonData.LogLevel)
	}

	// Expand environment variables and ~ in the path.
	// This can be used with environment variables like ${OMD_ROOT}.
	expandedPath := os.ExpandEnv(jsonData.LogPath)
	expandedPath = os.Expand(expandedPath, func(key string) string {
		// Handle ~ expansion manually.
		if key == "~" {
			home, _ := os.UserHomeDir()

			return home
		}

		return ""
	})

	// Create directories if they don't exist.
	dir := filepath.Dir(expandedPath)
	if dir != "." {
		//nolint:mnd // permission bits
		err := os.MkdirAll(dir, 0o0750)
		if err != nil {
			return nil, nil, fmt.Errorf("error when making directories for the logfile %q: %w", expandedPath, err)
		}
	}

	// Touch the file eagerly so an unusable path fails here, at datasource creation,
	// instead of silently on the first write. lumberjack opens the file lazily.
	//nolint:gosec,mnd // filename to open is dynamic due to user input, and that is intended , permission bits
	file, err := os.OpenFile(expandedPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o0640)
	if err != nil {
		return nil, nil, fmt.Errorf("error when opening the logfile %q: %w", expandedPath, err)
	}

	err = file.Close()
	if err != nil {
		return nil, nil, fmt.Errorf("error when closing the file descriptor: %w", err)
	}

	encoderConfig := zap.NewProductionEncoderConfig()

	encoderConfig.TimeKey = "t"
	// time.RFC3339Nano is what grafana logfmt uses.
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339Nano)

	encoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		// Include the function name in the caller field for higher log levels.
		if jsonData.LogLevel >= 4 && caller.Function != "" {
			enc.AppendString(caller.Function + " (" + caller.TrimmedPath() + ")")
		} else {
			enc.AppendString(caller.TrimmedPath())
		}
	}

	// lumberjack rotates the log file so a long-running debug deployment cannot grow it
	// without bound.
	rotator := &lumberjack.Logger{
		Filename:   expandedPath,
		MaxSize:    logMaxSizeMB,
		MaxBackups: logMaxBackups,
		MaxAge:     logMaxAgeDays,
		LocalTime:  true,
		Compress:   true,
	}

	core := zapcore.NewCore(
		zaplogfmt.NewEncoder(encoderConfig),
		zapcore.AddSync(rotator),
		logLevel,
	)

	logger := zap.New(core, zap.AddCaller()).Sugar()

	return logger, func() {
		_ = logger.Sync()
		_ = rotator.Close()
	}, nil
}
