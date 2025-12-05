package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bombsimon/logrusr/v4"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
)

var (
	standardLogger Logger = &logger{fl: logrus.StandardLogger()}
)

const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel Level = iota
	// FatalLevel level. Logs and then calls `logger.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	TraceLevel
)

func StandardLogger() Logger {
	return standardLogger
}

func New() Logger {
	return &logger{fl: logrus.New()}
}

func FromLogrus(fl LogrusLogger) Logger {
	return &logger{fl: fl}
}

type Level = logrus.Level

type Logger interface {
	WithContext(ctx context.Context) Logger

	WithReportCaller(b bool, depth ...uint) Logger
	WithOffset(n int) Logger

	WithField(key string, value any) Logger
	WithFields(kv ...any) Logger
	WithError(err error) Logger

	SetLevel(level Level) Logger
	WriterLevel(level Level) *io.PipeWriter

	SetOutput(w io.Writer) Logger

	Tracef(format string, args ...any)
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Printf(format string, args ...any)
	Warnf(format string, args ...any)
	Warningf(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	Panicf(format string, args ...any)

	Trace(args ...any)
	Debug(args ...any)
	Info(args ...any)
	Print(args ...any)
	Warn(args ...any)
	Warning(args ...any)
	Error(args ...any)
	Fatal(args ...any)
	Panic(args ...any)

	Traceln(args ...any)
	Debugln(args ...any)
	Infoln(args ...any)
	Println(args ...any)
	Warnln(args ...any)
	Warningln(args ...any)
	Errorln(args ...any)
	Fatalln(args ...any)
	Panicln(args ...any)

	Logr() logr.Logger
	FieldLogger() logrus.FieldLogger
	Logger() *logrus.Logger

	Clone() Logger
}

type logger struct {
	fl           LogrusLogger
	reportCaller *int
	offset       int
}

func (l *logger) Tracef(format string, args ...any) {
	l.logf(TraceLevel, format, args...)
}

func (l *logger) Debugf(format string, args ...any) {
	l.logf(DebugLevel, format, args...)
}

func (l *logger) Infof(format string, args ...any) {
	l.logf(InfoLevel, format, args...)
}

func (l *logger) Printf(format string, args ...any) {
	l.logf(InfoLevel, format, args...)
}

func (l *logger) Warnf(format string, args ...any) {
	l.logf(WarnLevel, format, args...)
}

func (l *logger) Warningf(format string, args ...any) {
	l.logf(WarnLevel, format, args...)
}

func (l *logger) Errorf(format string, args ...any) {
	l.logf(ErrorLevel, format, args...)
}

func (l *logger) Fatalf(format string, args ...any) {
	l.logf(FatalLevel, format, args...)
	os.Exit(1)
}

func (l *logger) Panicf(format string, args ...any) {
	l.logf(PanicLevel, format, args...)
}

func (l *logger) Trace(args ...any) {
	l.log(TraceLevel, args...)
}

func (l *logger) Debug(args ...any) {
	l.log(DebugLevel, args...)
}

func (l *logger) Info(args ...any) {
	l.log(InfoLevel, args...)
}

func (l *logger) Print(args ...any) {
	l.log(InfoLevel, args...)
}

func (l *logger) Warn(args ...any) {
	l.log(WarnLevel, args...)
}

func (l *logger) Warning(args ...any) {
	l.log(WarnLevel, args...)
}

func (l *logger) Error(args ...any) {
	l.log(ErrorLevel, args...)
}

func (l *logger) Fatal(args ...any) {
	l.log(FatalLevel, args...)
	os.Exit(1)
}

func (l *logger) Panic(args ...any) {
	l.log(PanicLevel, args...)
}

func (l *logger) Traceln(args ...any) {
	l.logln(TraceLevel, args...)
}

func (l *logger) Debugln(args ...any) {
	l.logln(DebugLevel, args...)
}

func (l *logger) Infoln(args ...any) {
	l.logln(InfoLevel, args...)
}

func (l *logger) Println(args ...any) {
	l.logln(InfoLevel, args...)
}

func (l *logger) Warnln(args ...any) {
	l.logln(WarnLevel, args...)
}

func (l *logger) Warningln(args ...any) {
	l.logln(WarnLevel, args...)
}

func (l *logger) Errorln(args ...any) {
	l.logln(ErrorLevel, args...)
}

func (l *logger) Fatalln(args ...any) {
	l.logln(FatalLevel, args...)
	os.Exit(1)
}

func (l *logger) Panicln(args ...any) {
	l.logln(PanicLevel, args...)
}

func (l *logger) WriterLevel(level Level) *io.PipeWriter {
	return l.Logger().WriterLevel(level)
}

func (l *logger) SetLevel(level Level) Logger {
	l.Logger().SetLevel(level)
	return l
}

func (l *logger) WithContext(ctx context.Context) Logger {
	switch t := l.fl.(type) {
	case *logrus.Logger:
		return &logger{fl: t.WithContext(ctx), reportCaller: l.reportCaller, offset: l.offset}
	case *logrus.Entry:
		return &logger{fl: t.WithContext(ctx), reportCaller: l.reportCaller, offset: l.offset}
	}
	panic(fmt.Sprintf("unexpected logger type %T", l.fl))
}

func (l *logger) WithField(key string, value any) Logger {
	return &logger{fl: l.fl.WithField(key, value), reportCaller: l.reportCaller, offset: l.offset}
}

func (l *logger) WithFields(kv ...any) Logger {
	log := &logger{fl: l.fl, reportCaller: l.reportCaller, offset: l.offset}
	for i := 0; i < len(kv); i += 2 {
		log = &logger{fl: log.fl.WithField(fmt.Sprintf("%v", kv[i]), kv[i+1]), reportCaller: l.reportCaller, offset: l.offset}
	}
	return log
}

func (l *logger) WithError(err error) Logger {
	return &logger{fl: l.fl.WithError(err), reportCaller: l.reportCaller, offset: l.offset}
}

func (l *logger) WithReportCaller(b bool, depth ...uint) Logger {
	if !b {
		return &logger{fl: l.fl, reportCaller: nil, offset: l.offset}
	}
	var d int
	if len(depth) > 0 {
		d = int(depth[0])
	} else {
		d = 0
	}
	return &logger{fl: l.fl, reportCaller: &d, offset: l.offset}
}

func (l *logger) WithOffset(n int) Logger {
	return &logger{fl: l.fl, reportCaller: l.reportCaller, offset: n}
}

func (l *logger) Logr() logr.Logger {
	return logrusr.New(l.fl, logrusr.WithFormatter(func(i any) any {
		return fmt.Sprintf("%v", i)
	}))
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

func (l *logger) Clone() Logger {
	n := logrus.New()
	switch t := l.fl.(type) {
	case *logrus.Logger:
		n.Level = t.Level
		n.Out = t.Out
		n.Formatter = t.Formatter
		n.Hooks = t.Hooks
		return &logger{fl: n, reportCaller: l.reportCaller, offset: l.offset}
	case *logrus.Entry:
		t = t.Dup()
		n.Level = t.Logger.Level
		n.Out = t.Logger.Out
		n.Formatter = t.Logger.Formatter
		n.Hooks = t.Logger.Hooks
		t.Logger = n
		return &logger{fl: t, reportCaller: l.reportCaller, offset: l.offset}
	}
	panic(fmt.Sprintf("unexpected logger type %T", l.fl))
}

func (l *logger) logf(level logrus.Level, format string, args ...any) {
	l.withCaller().Logf(l.level(level), format, args...)
}

func (l *logger) log(level logrus.Level, args ...any) {
	l.withCaller().Log(l.level(level), args...)
}

func (l *logger) logln(level logrus.Level, args ...any) {
	l.withCaller().Logln(l.level(level), args...)
}

func (l *logger) withCaller() LogrusLogger {
	if l.reportCaller == nil {
		return l.fl
	}
	pcs := make([]uintptr, 1)
	runtime.Callers(4+*l.reportCaller, pcs)
	f, _ := runtime.CallersFrames(pcs).Next()
	pkg := getPackageName(f.Function)
	return l.fl.WithField("caller", fmt.Sprintf("%s/%s:%d", pkg, filepath.Base(f.File), f.Line)).WithField("func", f.Func.Name())
}

func (l *logger) level(lvl Level) logrus.Level {
	if lvl > 3 {
		return lvl + logrus.Level(l.offset)
	}
	return lvl
}

// getPackageName reduces a fully qualified function name to the package name
// There really ought to be a better way...
func getPackageName(f string) string {
	for {
		lastPeriod := strings.LastIndex(f, ".")
		lastSlash := strings.LastIndex(f, "/")
		if lastPeriod > lastSlash {
			f = f[:lastPeriod]
		} else {
			break
		}
	}

	return f
}

type LogrusLogger interface {
	logrus.FieldLogger
	logrus.Ext1FieldLogger

	Log(level logrus.Level, args ...any)
	Logf(level logrus.Level, format string, args ...any)
	Logln(level logrus.Level, args ...any)
}
