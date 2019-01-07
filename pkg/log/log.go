package log

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"

	"github.com/TV4/logrus-stackdriver-formatter"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

type (
	loggerKey struct{}
)

// Logger interface extends the FieldLogger interface with a new `F` method that is
// a shortcut to the `WithFields` method
type Logger interface {
	logrus.FieldLogger
	F(keyvals ...interface{}) Logger
}

type logger struct {
	logrus.FieldLogger
}

var (
	DebugLevel = logrus.DebugLevel
	InfoLevel  = logrus.InfoLevel
	WarnLevel  = logrus.WarnLevel
	ErrorLevel = logrus.ErrorLevel

	// L is an alias for the the standard logger.
	L = &logger{logrus.New()}

	// G is an alias for GetLogger.
	G = GetLogger
)

func New(service, version string, level logrus.Level) Logger {
	log := logrus.New()
	log.SetLevel(level)
	log.SetFormatter(stackdriver.NewFormatter(stackdriver.WithService(service), stackdriver.WithVersion(version)))
	return &logger{log}
}

func NewTest() (Logger, *test.Hook) {
	l := logrus.New()
	l.Out = ioutil.Discard
	l.Level = logrus.DebugLevel
	return &logger{l}, test.NewLocal(l)
}

// WithLogger returns a new context with the provided logger. Use in
// combination with logger.WithField(s) for great effect.
func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// GetLogger retrieves the current logger from the context. If no logger is
// available, the default logger is returned.
func GetLogger(ctx context.Context) Logger {
	logger := ctx.Value(loggerKey{})

	if logger == nil {
		return L
	}

	return logger.(Logger)
}

var errMissingValue = errors.New("(MISSING)")

func (l *logger) F(keyvals ...interface{}) Logger {
	// Short circuit if first arg is already a map
	if len(keyvals) == 1 && reflect.TypeOf(keyvals[0]).Kind() == reflect.Map {
		return &logger{l.WithFields(keyvals[0].(map[string]interface{}))}
	}

	fields := logrus.Fields{}
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			fields[fmt.Sprint(keyvals[i])] = keyvals[i+1]
		} else {
			fields[fmt.Sprint(keyvals[i])] = errMissingValue
		}
	}
	return &logger{l.WithFields(fields)}
}

func RequestErrors(status int) bool {
	return status >= 400
}

func RequestAll(_ int) bool {
	return true
}

func RequestNever(_ int) bool {
	return false
}
