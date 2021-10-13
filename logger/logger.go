package logger

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

var (
	StandardLogger Logger = &logger{FieldLogger: logrus.StandardLogger()}
)

func New() Logger {
	return &logger{FieldLogger: logrus.New()}
}

type Logger interface {
	WithField(key string, value interface{}) Logger
	WithFields(kv ...interface{}) Logger
	WithError(err error) Logger

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Printf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	Debug(args ...interface{})
	Info(args ...interface{})
	Print(args ...interface{})
	Warn(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})

	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Println(args ...interface{})
	Warnln(args ...interface{})
	Warningln(args ...interface{})
	Errorln(args ...interface{})
	Fatalln(args ...interface{})
	Panicln(args ...interface{})
}

type logger struct {
	logrus.FieldLogger
}

func (l *logger) WithField(key string, value interface{}) Logger {
	return &logger{FieldLogger: l.FieldLogger.WithField(key, value)}
}

func (l *logger) WithFields(kv ...interface{}) Logger {
	for i := 0; i < len(kv); i += 2 {
		l.FieldLogger = l.FieldLogger.WithField(fmt.Sprintf("%v", kv[i]), kv[i+1])
	}
	return l
}

func (l *logger) WithError(err error) Logger {
	return &logger{FieldLogger: l.FieldLogger.WithError(err)}
}
