package plugin

import (
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

// Loggers bundles a datasource's log destinations. Logs go to Grafana's
// logging pipeline; Grafana filters them by level and scope.

// Loggers struct contained multiple logging targets
// One of them was a standard file logger, but grafana backend plugins are not allowed to access filesystem.
// Exceptions are detected during plugin source code validation.
type Loggers struct {
	sdk log.Logger
}

// NewLoggers builds a datasource's loggers, scoped by datasource uid.
func NewLoggers(uid string) *Loggers {
	return &Loggers{
		sdk: log.DefaultLogger.With("datasource", uid),
	}
}

// debugf logs a formatted debug message to Grafana's logging pipeline.
func (l *Loggers) debugf(format string, args ...any) {
	l.sdk.Debug(fmt.Sprintf(format, args...))
}

// warnf logs a formatted warning message to Grafana's logging pipeline.
func (l *Loggers) warnf(format string, args ...any) {
	l.sdk.Warn(fmt.Sprintf(format, args...))
}
