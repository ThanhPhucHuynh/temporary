package glog

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type HookFormatter func(*Hook, *logrus.Entry) error

type Hook struct {
	SkipDepth  int
	SkipKey    string
	LogLevels  []logrus.Level
	Formatter  HookFormatter
	Release    bool
	PackageLog string

	FilePrefixIgnore string
}

func (hook *Hook) Levels() []logrus.Level {
	return hook.LogLevels
}

func (hook *Hook) Fire(entry *logrus.Entry) error {
	if hook.SkipKey != "" {
		if skipValue, ok := entry.Data[hook.SkipKey]; ok {
			if skipInt, ok := skipValue.(int); ok {
				hook.SkipDepth = skipInt
			}
			if hook.Release {
				delete(entry.Data, hook.SkipKey)
			}
		}
	}
	return hook.Formatter(hook, entry)
}

func NewHook(options ...Option) *Hook {
	hook := &Hook{
		Formatter: fileFormatter,
		Release:   true,
	}

	for _, option := range options {
		option(hook)
	}

	if len(hook.LogLevels) == 0 {
		hook.LogLevels = logrus.AllLevels
	}

	return hook
}

func fileFormatter(hook *Hook, entry *logrus.Entry) error {
	f := GetCaller(hook.SkipDepth)

	for strings.Contains(f.File, hook.PackageLog) {
		f = GetCaller(hook.SkipDepth + 1)
	}

	filePathStr := f.File
	if hook.FilePrefixIgnore != "" {
		pathArr := strings.Split(filePathStr, hook.FilePrefixIgnore)
		filePathStr = pathArr[len(pathArr)-1]
	}

	if f != nil {
		entry.Data["file"] = fmt.Sprintf("%s:%d", filePathStr, f.Line)
	}

	return nil
}
