package glog

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGetCallerEmpty(t *testing.T) {

	old := getPackageNameFunc
	defer func() {
		getPackageNameFunc = old
	}()

	getPackageNameFunc = func(f string) string {
		return ""
	}

	rt := GetCaller(1)
	assert.Empty(t, rt)
}

func TestFileFormatter(t *testing.T) {
	hook := NewHook(WithSkipKey("skip"), WithRelease(true), WithLogLevels([]logrus.Level{logrus.DebugLevel}), WithFormatter(fileFormatter))

	entry := logrus.WithFields(logrus.Fields{
		"skip": 2,
	})

	buf := bytes.Buffer{}

	logrus.SetOutput(&bufferWriteCloser{&buf})
	logrus.SetLevel(logrus.InfoLevel)

	err := hook.Fire(entry)

	if err != nil {
		t.Fatalf("Error firing hook: %v", err)
	}

	mgs := "test"

	logrus.Infoln(mgs)
	assert.Contains(t, buf.String(), mgs)

}
