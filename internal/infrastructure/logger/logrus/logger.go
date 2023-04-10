package logrus

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"runtime"
	"strings"

	"fry.org/cmo/cli/internal/application/logger"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"github.com/speijnik/go-errortree"
)

// Logger provides a logrus implementation of the Service
type Logger struct {
	log   *logrus.Logger
	entry *logrus.Entry
}

func (l *Logger) SetLevel(level string) error {
	var rcerror error

	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return errortree.Add(rcerror, "SetLevel", err)
	}

	l.entry.Logger.Level = lvl

	return nil
}

// SetFormat sets the log target and format.
// Example: logrus:stdout?output=[plain|json], logrus:hooked?output=[plain|json], logrus:stderr?output=[plain|json]
func (l *Logger) SetFormat(format string) error {
	var rcerror error
	u, err := url.Parse(format)
	if err != nil {
		return errortree.Add(rcerror, "SetFormat", err)
	}
	if u.Scheme != "logrus" {
		return errortree.Add(rcerror, "SetFormat", fmt.Errorf("invalid scheme %s", u.Scheme))
	}
	output := u.Query().Get("output")
	if output == "json" {
		l.log.SetFormatter(&logrus.JSONFormatter{})
		l.entry.Logger.Formatter = &logrus.JSONFormatter{}
	} else {
		l.log.SetFormatter(&logrus.TextFormatter{})
		l.entry.Logger.Formatter = &logrus.TextFormatter{}
	}

	switch u.Opaque {
	case "stdout":
		l.entry.Logger.Out = os.Stdout
	case "stderr":
		l.entry.Logger.Out = os.Stderr
	case "hooked":
		l.log.SetOutput(io.Discard) // Send all logs to nowhere by default
		l.log.AddHook(&writer.Hook{ // Send logs with level higher than warning to stderr
			Writer: os.Stderr,
			LogLevels: []logrus.Level{
				logrus.PanicLevel,
				logrus.FatalLevel,
				logrus.ErrorLevel,
				logrus.WarnLevel,
			},
		})
		l.log.AddHook(&writer.Hook{ // Send info and debug logs to stdout
			Writer: os.Stdout,
			LogLevels: []logrus.Level{
				logrus.InfoLevel,
				logrus.DebugLevel,
			},
		})
	default:
		return errortree.Add(rcerror, "SetFormat", fmt.Errorf("unsupported format %s", format))
	}

	return nil
}

// NewLogger creates a new `log.Logger` from the provided entry
func NewLogger() logger.Logger {

	l := logrus.New()

	out := Logger{
		log:   l,
		entry: logrus.NewEntry(l),
	}

	return &out
}

// Debug logs a message at level Debug on the standard logger.
func (l *Logger) Debug(args ...interface{}) {

	l.sourced().Debug(args...)
}

// Debug logs a message at level Debug on the standard logger.
func (l *Logger) Debugln(args ...interface{}) {
	l.sourced().Debugln(args...)
}

// Debugf logs a message at level Debug on the standard logger.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.sourced().Debugf(format, args...)
}

// Info logs a message at level Info on the standard logger.
func (l *Logger) Info(args ...interface{}) {
	l.sourced().Info(args...)
}

// Info logs a message at level Info on the standard logger.
func (l *Logger) Infoln(args ...interface{}) {
	l.sourced().Infoln(args...)
}

// Infof logs a message at level Info on the standard logger.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.sourced().Infof(format, args...)
}

// Warn logs a message at level Warn on the standard logger.
func (l *Logger) Warn(args ...interface{}) {
	l.sourced().Warn(args...)
}

// Warn logs a message at level Warn on the standard logger.
func (l *Logger) Warnln(args ...interface{}) {
	l.sourced().Warnln(args...)
}

// Warnf logs a message at level Warn on the standard logger.
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.sourced().Warnf(format, args...)
}

// Error logs a message at level Error on the standard logger.
func (l *Logger) Error(args ...interface{}) {
	l.sourced().Error(args...)
}

// Error logs a message at level Error on the standard logger.
func (l *Logger) Errorln(args ...interface{}) {
	l.sourced().Errorln(args...)
}

// Errorf logs a message at level Error on the standard logger.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.sourced().Errorf(format, args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func (l *Logger) Fatal(args ...interface{}) {
	l.sourced().Fatal(args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func (l *Logger) Fatalln(args ...interface{}) {
	l.sourced().Fatalln(args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.sourced().Fatalf(format, args...)
}

func (l *Logger) With(key string, value interface{}) logger.Logger {

	out := Logger{
		entry: l.entry.WithField(key, value),
	}

	return &out
}

func (l *Logger) WithFields(fields logger.Fields) logger.Logger {

	out := Logger{
		entry: l.entry.WithFields(logrus.Fields(fields)),
	}

	return &out
}

// sourced adds a source field to the logger that contains
// the file name and line where the logging happened.
func (l *Logger) sourced() *logrus.Entry {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "<???>"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		file = file[slash+1:]
	}

	return l.entry.WithField("source", fmt.Sprintf("%s:%d", file, line))
}
