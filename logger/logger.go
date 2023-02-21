package logger

import (
	"context"
	"fmt"
	"io"

	"github.com/bombsimon/logrusr/v2"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
)

var (
	standardLogger Logger = &logger{fl: logrus.StandardLogger()}
)

func StandardLogger() Logger {
	return standardLogger
}

func New() Logger {
	return &logger{fl: logrus.New()}
}

func FromLogrus(fl logrus.Ext1FieldLogger) Logger {
	return &logger{fl: fl}
}

type Logger interface {
	WithContext(ctx context.Context) Logger

	WithField(key string, value interface{}) Logger
	WithFields(kv ...interface{}) Logger
	WithError(err error) Logger

	SetLevel(level logrus.Level) Logger
	WriterLevel(level logrus.Level) *io.PipeWriter

	SetOutput(w io.Writer) Logger

	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Printf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	Trace(args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Print(args ...interface{})
	Warn(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})

	Traceln(args ...interface{})
	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Println(args ...interface{})
	Warnln(args ...interface{})
	Warningln(args ...interface{})
	Errorln(args ...interface{})
	Fatalln(args ...interface{})
	Panicln(args ...interface{})

	Logr() logr.Logger
	FieldLogger() logrus.FieldLogger
	Logger() *logrus.Logger
}

type logger struct {
	fl logrus.Ext1FieldLogger
}

func (l *logger) Tracef(format string, args ...interface{}) {
	l.fl.Tracef(format, args...)
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.fl.Debugf(format, args...)
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.fl.Infof(format, args...)
}

func (l *logger) Printf(format string, args ...interface{}) {
	l.fl.Printf(format, args...)
}

func (l *logger) Warnf(format string, args ...interface{}) {
	l.fl.Warnf(format, args...)
}

func (l *logger) Warningf(format string, args ...interface{}) {
	l.fl.Warningf(format, args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.fl.Errorf(format, args...)
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	l.fl.Fatalf(format, args...)
}

func (l *logger) Panicf(format string, args ...interface{}) {
	l.fl.Panicf(format, args...)
}

func (l *logger) Trace(args ...interface{}) {
	l.fl.Trace(args...)
}

func (l *logger) Debug(args ...interface{}) {
	l.fl.Debug(args...)
}

func (l *logger) Info(args ...interface{}) {
	l.fl.Info(args...)
}

func (l *logger) Print(args ...interface{}) {
	l.fl.Print(args...)
}

func (l *logger) Warn(args ...interface{}) {
	l.fl.Warn(args...)
}

func (l *logger) Warning(args ...interface{}) {
	l.fl.Warning(args...)
}

func (l *logger) Error(args ...interface{}) {
	l.fl.Error(args...)
}

func (l *logger) Fatal(args ...interface{}) {
	l.fl.Fatal(args...)
}

func (l *logger) Panic(args ...interface{}) {
	l.fl.Panic(args...)
}

func (l *logger) Traceln(args ...interface{}) {
	l.fl.Traceln(args...)
}

func (l *logger) Debugln(args ...interface{}) {
	l.fl.Debugln(args...)
}

func (l *logger) Infoln(args ...interface{}) {
	l.fl.Infoln(args...)
}

func (l *logger) Println(args ...interface{}) {
	l.fl.Println(args...)
}

func (l *logger) Warnln(args ...interface{}) {
	l.fl.Warnln(args...)
}

func (l *logger) Warningln(args ...interface{}) {
	l.fl.Warningln(args...)
}

func (l *logger) Errorln(args ...interface{}) {
	l.fl.Errorln(args...)
}

func (l *logger) Fatalln(args ...interface{}) {
	l.fl.Fatalln(args...)
}

func (l *logger) Panicln(args ...interface{}) {
	l.fl.Panicln(args...)
}

func (l *logger) WriterLevel(level logrus.Level) *io.PipeWriter {
	return l.Logger().WriterLevel(level)
}

func (l *logger) SetLevel(level logrus.Level) Logger {
	l.Logger().SetLevel(level)
	return l
}

func (l *logger) WithContext(ctx context.Context) Logger {
	switch t := l.fl.(type) {
	case *logrus.Logger:
		return &logger{fl: t.WithContext(ctx)}
	case *logrus.Entry:
		return &logger{fl: t.Logger.WithContext(ctx)}
	}
	panic(fmt.Sprintf("unexpected logger type %T", l.fl))
}

func (l *logger) WithField(key string, value interface{}) Logger {
	return &logger{fl: l.fl.WithField(key, value)}
}

func (l *logger) WithFields(kv ...interface{}) Logger {
	log := &logger{fl: l.fl}
	for i := 0; i < len(kv); i += 2 {
		log = &logger{fl: log.fl.WithField(fmt.Sprintf("%v", kv[i]), kv[i+1])}
	}
	return log
}

func (l *logger) WithError(err error) Logger {
	return &logger{fl: l.fl.WithError(err)}
}

func (l *logger) Logr() logr.Logger {
	return logrusr.New(l.fl)
}

func (l *logger) FieldLogger() logrus.FieldLogger {
	return l.fl
}

func (l *logger) Logger() *logrus.Logger {
	switch t := l.fl.(type) {
	case *logrus.Logger:
		return t
	case *logrus.Entry:
		return t.Logger
	}
	panic(fmt.Sprintf("unexpected logger type %T", l.fl))
}

func (l *logger) SetOutput(w io.Writer) Logger {
	l.Logger().SetOutput(w)
	return l
}
