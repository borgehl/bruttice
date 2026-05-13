package logx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
	"os"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	Logger *logrus.Logger
	entry  *logrus.Entry
	fields logrus.Fields
}

func NewLogger(fields logrus.Fields, hooks []logrus.Hook) *Logger {
	logrusLogger := logrus.New()

	for _, hook := range hooks {
		logrusLogger.AddHook(hook)
	}

	var logLevelMap = map[string]logrus.Level{
		"debug": logrus.DebugLevel,
		"info":  logrus.InfoLevel,
		"warn":  logrus.WarnLevel,
		"error": logrus.ErrorLevel,
	}
	logrusLogger.SetLevel(logLevelMap[util.GetLogLevel("LOG_LEVEL")])
	logrusLogger.SetOutput(os.Stdout)

	return &Logger{
		Logger: logrusLogger,
		fields: fields,
	}
}

func (l *Logger) SetField(name, value string) {
	l.fields[name] = value
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.tagged().Debugf(format, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.tagged().Infof(format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.tagged().Warnf(format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.tagged().Errorf(format, args...)
}

func (l *Logger) Panicf(format string, args ...interface{}) {
	l.tagged().Panicf(format, args...)
}

func (l *Logger) Debug(args ...interface{}) {
	l.tagged().Debug(args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.tagged().Info(args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.tagged().Warn(args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.tagged().Error(args...)
}

func (l *Logger) Panic(args ...interface{}) {
	l.tagged().Panic(args...)
}

func (l *Logger) WithFields(fields map[string]interface{}) {
	l.entry = l.Logger.WithFields(fields)
}

func (l *Logger) tagged() (entry *logrus.Entry) {
	if l.entry == nil {
		return l.Logger.WithFields(l.fields)
	}
	entry = l.entry.WithFields(l.fields)
	l.entry = nil
	return
}
