package glog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

//go:noinline
func BenchmarkGlog(b *testing.B) {
	b.ReportAllocs()
	l := New()
	l.logger.Logger.Out = ioutil.Discard
	for i := 0; i < b.N; i++ {
		l.logger.Info("multi",
			"bool", true, "string", "str", "int", 42,
			"float", 3.14, "struct", struct{ X, Y int }{93, 76})
	}
}

//go:noinline
func BenchmarkGlog1(b *testing.B) {
	b.ReportAllocs()
	l := New()
	l.logger.Logger.Out = ioutil.Discard
	for i := 0; i < 1; i++ {
		l.logger.Info("multi",
			"bool", true, "string", "str", "int", 42,
			"float", 3.14, "struct", struct{ X, Y int }{93, 76})
	}
}

//go:noinline
func BenchmarkGlog10(b *testing.B) {
	b.ReportAllocs()
	l := New()
	l.logger.Logger.Out = ioutil.Discard
	for i := 0; i < 10; i++ {
		l.logger.Info("multi",
			"bool", true, "string", "str", "int", 42,
			"float", 3.14, "struct", struct{ X, Y int }{93, 76})
	}
}

type bufferWriteCloser struct {
	*bytes.Buffer
}

func (bwc *bufferWriteCloser) Close() error {
	return nil
}

func TestGlog(t *testing.T) {
	// Prepare the logger instance to be tested
	buf := bytes.Buffer{}

	old := getOutputFunc
	defer func() {
		getOutputFunc = old
	}()

	getOutputFunc = func() io.WriteCloser {
		return &bufferWriteCloser{&buf}
	}
	l := New()

	// Test logging methods
	tests := []struct {
		name    string
		method  func(format string, v ...interface{})
		message string
	}{
		{"Infof", l.Infof, "info message"},
		{"Debugf", l.Debugf, "debug message"},
		{"Warnf", l.Warnf, "warning message"},
		{"Errorf", l.Errorf, "error message"},
		{"Panicf", func(format string, v ...interface{}) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("panic recovered: %v", r)
				}
			}()
			l.Panicf(format, v...)
		}, "panic message"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.method(tt.message)
			assert.Contains(t, buf.String(), tt.message)
			buf.Reset()
		})
	}

	// Test logging methods with context
	testWithContext := []struct {
		name    string
		method  func(ctx context.Context, format string, v ...interface{})
		message string
		context context.Context
	}{
		{"InfoWithContext", l.InfoWithContext, "info message", context.Background()},
		{"DebugWithContext", l.DebugWithContext, "debug message", context.Background()},
		{"WarnWithContext", l.WarnWithContext, "warning message", context.Background()},
		{"ErrorWithContext", l.ErrorWithContext, "error message", context.Background()},
		{"PanicWithContext", func(ctx context.Context, format string, v ...interface{}) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("panic recovered: %v", r)
				}
			}()
			l.PanicWithContext(ctx, format, v...)
		}, "panic message", context.Background()},
	}

	for _, tt := range testWithContext {
		t.Run(tt.name, func(t *testing.T) {
			tt.method(tt.context, tt.message)
			assert.Contains(t, buf.String(), tt.message)
			buf.Reset()
		})
	}

	// Test WithField method
	l2 := l.WithField("foo", "bar")
	l2.Infof("message with foo field")
	assert.Contains(t, buf.String(), "foo")
	assert.Contains(t, buf.String(), "bar")

	// Test WithField overwrite
	l2 = l2.WithField("foo", "pub")
	l2.Infof("message with foo field")
	assert.Contains(t, buf.String(), "bar/pub")

	// Test WithContext method
	ctx := context.WithValue(context.Background(), "request_id", "123")
	l3 := l.withContext(ctx)
	l3.Infof(" with request_id context")
	assert.Contains(t, buf.String(), "request_id")
	assert.Contains(t, buf.String(), "123")

	buf.Reset()
	l4 := l.withContext(context.Background())
	l4.Infof("with no context")
	assert.NotContains(t, buf.String(), "request_id")

	// // Test Close method
	err := l.Close()
	if err != nil {
		t.Errorf("unexpected error when closing logger: %v", err)
	}
	if _, ok := l.writer.(io.Closer); !ok {
		t.Errorf("unexpected writer: expected io.WriteCloser, got %T", l.writer)
	}

	// Test Close method write nil
	l5 := New()
	l5.writer = nil

	assert.Nil(t, l5.Close())
}

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name          string
		env           map[string]string
		expectedLevel string
		expectedType  string
	}{
		{
			name:          "default",
			env:           map[string]string{},
			expectedLevel: "debug",
			expectedType:  "json",
		},
		// {
		// 	name: "file",
		// 	env: map[string]string{
		// 		"LOG_OUTPUT": "file://test.log",
		// 	},
		// 	expectedLevel: "debug",
		// 	expectedType:  "json",
		// },
		{
			name: "json",
			env: map[string]string{
				"LOG_FORMAT": "json",
			},
			expectedLevel: "debug",
			expectedType:  "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := os.Environ()
			for k, v := range tt.env {
				env = append(env, k+"="+v)
			}
			os.Clearenv()
			for _, kv := range env {
				parts := strings.SplitN(kv, "=", 2)
				os.Setenv(parts[0], parts[1])
			}
			defer func() {
				os.Clearenv()
				for _, kv := range env {
					parts := strings.SplitN(kv, "=", 2)
					os.Setenv(parts[0], parts[1])
				}
			}()

			l := New()

			// Check log level
			assert.Equal(t, tt.expectedLevel, l.logger.Logger.Level.String())

			// Check log formatter
			var formatter string
			if tt.expectedType == "text" {
				formatter = "*logrus.TextFormatter"
			} else {
				formatter = "*logrus.JSONFormatter"
			}
			assert.IsType(t, &logrus.Logger{}, l.logger.Logger)
			assert.IsType(t, &logrus.Entry{}, l.logger)
			assert.IsType(t, &logrus.JSONFormatter{}, l.logger.Logger.Formatter)
			assert.Implements(t, (*logrus.Formatter)(nil), l.logger.Logger.Formatter)
			assert.Equal(t, formatter, fmt.Sprintf("%T", l.logger.Logger.Formatter))

			// Check log output
			if tt.env["LOG_OUTPUT"] != "" && strings.HasPrefix(tt.env["LOG_OUTPUT"], filePrefix) {
				// If output is a file, check if file is created
				_, err := os.Stat(tt.env["LOG_OUTPUT"][len(filePrefix):])
				assert.NoError(t, err)
			} else {
				// Otherwise, check if output is set to os.Stdout
				assert.NotEmpty(t, l.logger.Logger.Out)
			}
		})
	}
}

func TestGetOutput(t *testing.T) {
	// Test case where os.Create succeeds
	old := osCreate
	defer func() {
		osCreate = old
	}()

	osCreate = func(name string) (*os.File, error) {
		return nil, os.ErrNotExist
	}

	os.Setenv("LOG_OUTPUT", "file://test.log")
	defer os.Unsetenv("LOG_OUTPUT")

	out := getOutput()
	assert.Equal(t, nil, out)
}

func TestGetOutputWithFile(t *testing.T) {
	os.Setenv("LOG_OUTPUT", "file:test.log")

	output := getOutput()

	file, ok := output.(*os.File)
	assert.True(t, ok)
	assert.NotNil(t, file)

	os.Unsetenv("LOG_OUTPUT")
}

func TestNew(t *testing.T) {
	l := New()
	assert.NotNil(t, l)
	assert.NotNil(t, l.logger)
	assert.NotNil(t, l.writer)
}

func TestGetFormatterJSON(t *testing.T) {
	os.Setenv("LOG_FORMAT", "")
	os.Setenv("LOG_TIME_FORMAT", "2006-01-02 15:04:05")
	formatter := getFormatter()

	assert.IsType(t, &logrus.JSONFormatter{}, formatter)
}

func TestGetFormatterText(t *testing.T) {
	os.Setenv("LOG_FORMAT", "text")
	os.Setenv("LOG_TIME_FORMAT", "2006-01-02 15:04:05")
	formatter := getFormatter()

	assert.IsType(t, &logrus.TextFormatter{}, formatter)
}

func TestGetLevelDebug(t *testing.T) {
	os.Setenv("LOG_LEVEL", "")
	level := getLevel()
	assert.Equal(t, logrus.DebugLevel, level)
}

func TestGetLevelError(t *testing.T) {
	os.Setenv("LOG_LEVEL", "error")
	level := getLevel()

	assert.Equal(t, logrus.ErrorLevel, level)
}

func TestGetOutputStdout(t *testing.T) {
	os.Setenv("LOG_OUTPUT", "")
	output := getOutput()

	assert.Equal(t, os.Stdout, output)
}

func TestGetOutputFile(t *testing.T) {
	fname := "logfile.txt"
	buffer := new(bytes.Buffer)
	os.Setenv("LOG_OUTPUT", "file://logfile.txt")
	os.Create(fname)
	output := getOutput()

	assert.NotEqual(t, os.Stdout, output)
	assert.Equal(t, strings.Contains(output.(*os.File).Name(), fname), true)
	output.Close()
	buffer.WriteString("")
	os.Remove(fname)
}
