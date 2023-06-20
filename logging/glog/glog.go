package glog

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var filePrefix = "file://"

var osCreate = os.Create
var getOutputFunc = getOutput

// Logger is an interface of logging operations
type Logger interface {
	Infof(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Panicf(format string, v ...interface{})

	InfoWithContext(ctx context.Context, format string, v ...interface{})
	DebugWithContext(ctx context.Context, format string, v ...interface{})
	WarnWithContext(ctx context.Context, format string, v ...interface{})
	ErrorWithContext(ctx context.Context, format string, v ...interface{})
	PanicWithContext(ctx context.Context, format string, v ...interface{})

	WithField(field string, value interface{}) Logger
	WithFieldNoAdds(field string, value interface{}) Logger
	Close() error
}

// glog is a logger implementation
type glog struct {
	logger *logrus.Entry
	writer io.WriteCloser
}

// env
// GLOG_PREFIXIGNORE
// PackageLog

func newInit() *logrus.Logger {
	logger := logrus.New()

	// if os.Getenv("GLOG_IS_RUN_CONTAINER") == "true" {
	// 	logger.AddHook(NewHook(WithSkip(1), WithFilePrefixIgnore("src/app"), WithPackageLog("app/pkg/glog")))
	// } else {
	// 	logger.AddHook(NewHook(WithSkip(1), WithFilePrefixIgnore("hdss-rest"), WithPackageLog("hdss-rest/pkg/glog")))
	// }

	logger.SetFormatter(getFormatter())
	logger.SetLevel(getLevel())

	return logger
}

func New() *glog {
	l := &glog{}
	logger := newInit()

	out := getOutputFunc()
	logger.SetOutput(out)
	l.writer = out
	l.logger = logrus.NewEntry(logger)
	return l
}

func NewWithKafkaHook(kh *KafkaHook) *glog {
	l := &glog{}
	logger := newInit()

	out := getOutputFunc()
	logger.SetOutput(out)
	l.writer = out

	logger.Hooks.Add(kh)
	l.logger = logrus.NewEntry(logger)

	return l
}

func getFormatter() logrus.Formatter {
	var formatter logrus.Formatter

	timeFormat := os.Getenv("LOG_TIME_FORMAT")
	if timeFormat == "" {
		timeFormat = "2006-01-02 15:04:05"
	}

	formatter = &logrus.JSONFormatter{
		TimestampFormat: timeFormat, //,time.RFC1123,
	}

	if os.Getenv("LOG_FORMAT") == "text" {
		formatter = &logrus.TextFormatter{
			TimestampFormat: timeFormat,
			FullTimestamp:   true,
		}
	}

	return formatter
}

func getLevel() logrus.Level {
	lvl, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		lvl = logrus.DebugLevel
	}
	return lvl
}

// getOutput returns the log output based on the LOG_OUTPUT environment variable.
func getOutput() io.WriteCloser {
	out := os.Getenv("LOG_OUTPUT")
	if strings.HasPrefix(out, filePrefix) {
		name := out[len(filePrefix):]
		f, err := osCreate(name)
		if err != nil {
			log.Printf("log: failed to create log file: %s, err: %v\n", name, err)
			return nil
		}
		return f
	}
	return os.Stdout
}

// Infof print info with format.
func (l *glog) Infof(format string, v ...interface{}) {
	l.logger.Infof(format, v...)
}

// Debugf print debug with format.
func (l *glog) Debugf(format string, v ...interface{}) {
	l.logger.Debugf(format, v...)
}

// Warnf print warning with format.
func (l *glog) Warnf(format string, v ...interface{}) {
	l.logger.Warnf(format, v...)
}

// Errorf print error with format.
func (l *glog) Errorf(format string, v ...interface{}) {
	l.logger.Errorf(format, v...)
}

// Panicf panic with format.
func (l *glog) Panicf(format string, v ...interface{}) {
	l.logger.Panicf(format, v...)
}

// InfoWithContext print info log with context
func (l *glog) InfoWithContext(ctx context.Context, format string, v ...interface{}) {
	l.withContext(ctx).Infof(format, v...)
}

// DebugWithContext print debug with context
func (l *glog) DebugWithContext(ctx context.Context, format string, v ...interface{}) {
	l.withContext(ctx).Debugf(format, v...)
}

// WarnWithContext print warning with context
func (l *glog) WarnWithContext(ctx context.Context, format string, v ...interface{}) {
	l.withContext(ctx).Warnf(format, v...)
}

// ErrorWithContext print error with context
func (l *glog) ErrorWithContext(ctx context.Context, format string, v ...interface{}) {
	l.withContext(ctx).Errorf(format, v...)
}

// PanicWithContext panic with context
func (l *glog) PanicWithContext(ctx context.Context, format string, v ...interface{}) {
	l.withContext(ctx).Panicf(format, v...)
}

func (l *glog) withContext(ctx context.Context) Logger {
	if requestID := ctx.Value("reid"); requestID != nil {
		return l.WithField("reid", requestID)
	}
	return l
}

// WithField return a new logger with field
func (l *glog) WithField(field string, value interface{}) Logger {

	fieldMap := l.logger.Data
	val, ok := fieldMap[field]
	if ok {
		value = fmt.Sprintf("%s/%s", val, value)
	}

	nl := l.logger.WithField(field, value)
	return &glog{
		logger: nl,
	}
}

func (l *glog) WithFieldNoAdds(field string, value interface{}) Logger {
	nl := l.logger.WithField(field, value)
	return &glog{
		logger: nl,
	}
}

// Close close the underlying writer
func (l *glog) Close() error {
	if l.writer != nil {
		return l.writer.Close()
	}
	return nil
}
