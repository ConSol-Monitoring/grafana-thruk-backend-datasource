package plugin

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	// need a blank import to define the logfmt.
	_ "github.com/jsternberg/zap-logfmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ErrInvalidLogLevel error type.
var ErrInvalidLogLevel = errors.New("invalid log level, has to be between [0-7]")

// ErrCouldNotBuildLogger error type.
var ErrCouldNotBuildLogger = errors.New("could not build logger from the configuration")

//nolint:funlen // is long, tries to create/touch the file first for possible permission errors
func createLoggerFromDatasourceSettings(jsonData *DatasourceSettingsJSONDataPartial) (*zap.SugaredLogger, error) {
	if jsonData == nil {
		return nil, fmt.Errorf("%w , argument: jsonData", ErrArgumentNil)
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
		return nil, fmt.Errorf("%w logLevel: %d", ErrInvalidLogLevel, jsonData.LogLevel)
	}

	// if not running on OMD, it will use root /var/log
	logPath := jsonData.LogPath
	if logPath == "" {
		logPath = "${OMD_ROOT}/var/log/grafana/consolmonitoring-thruk-datasource.log"
	}

	// Expand environment variables and ~ in the path
	// This can be used with environment variables like ${OMD_ROOT}
	expandedPath := os.ExpandEnv(logPath)
	expandedPath = os.Expand(expandedPath, func(key string) string {
		// Handle ~ expansion manually
		if key == "~" {
			home, _ := os.UserHomeDir()

			return home
		}
		// Let os.ExpandEnv handle other env vars
		return ""
	})

	// Create directories if they don't exist
	dir := filepath.Dir(expandedPath)
	if dir != "." {
		//nolint:mnd // permission bits
		err := os.MkdirAll(dir, 0o0750)
		if err != nil {
			return nil, fmt.Errorf("error when making directories for the logfile %q: %w", expandedPath, err)
		}
	}

	filename := expandedPath

	//nolint:gosec,mnd  // filename to open is dynamic due to user input, and that is intended , permission bits
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o0640)
	if err != nil {
		return nil, fmt.Errorf("error when opening the logfile %q: %w", filename, err)
	}

	err = file.Close()
	if err != nil {
		return nil, fmt.Errorf("error when closing the file descriptor: %w", err)
	}

	config := zap.NewProductionConfig()

	config.Encoding = "logfmt" // same format as grafanas own logs

	config.EncoderConfig.TimeKey = "t"
	// time.RFC3339Nano is what grafana logfmt uses
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339Nano)

	config.EncoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		// include the function name in the caller field for higher log levels
		if jsonData.LogLevel >= 4 && caller.Function != "" {
			enc.AppendString(caller.Function + " (" + caller.TrimmedPath() + ")")
		} else {
			enc.AppendString(caller.TrimmedPath())
		}
	}

	config.OutputPaths = append(config.OutputPaths, filename)

	config.Level.SetLevel(logLevel)

	loggerNormal, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("%w, %w", ErrCouldNotBuildLogger, err)
	}

	return loggerNormal.Sugar(), nil
}
