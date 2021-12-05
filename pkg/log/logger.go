package log

import (
	"io"

	"github.com/sirupsen/logrus"
)

type Logger interface {
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
}

type Log struct {
	logger *logrus.Logger
}

func NewLog(logger *logrus.Logger, level logrus.Level, out io.Writer) *Log {
	logger.SetLevel(level)
	logger.SetOutput(out)
	return &Log{logger: logger}
}

func (l *Log) Info(args ...interface{}) {
	l.logger.Info(args...)
}

func (l *Log) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *Log) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

func (l *Log) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

func (l *Log) Error(args ...interface{}) {
	l.logger.Error(args...)
}

func (l *Log) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func (l *Log) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

func (l *Log) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}
