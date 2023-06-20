package glog

import "github.com/sirupsen/logrus"

type Option func(*Hook)

func WithSkip(skip int) Option {
	return func(h *Hook) {
		h.SkipDepth = skip
	}
}

func WithPackageLog(str string) Option {
	return func(h *Hook) {
		h.PackageLog = str
	}
}
func WithFilePrefixIgnore(projectName string) Option {
	return func(h *Hook) {
		h.FilePrefixIgnore = projectName
	}
}

func WithLogLevels(levels []logrus.Level) Option {
	return func(h *Hook) {
		h.LogLevels = levels
	}
}

func WithFormatter(formatter HookFormatter) Option {
	return func(h *Hook) {
		h.Formatter = formatter
	}
}

func WithSkipKey(skipKey string) Option {
	return func(h *Hook) {
		h.SkipKey = skipKey
	}
}

func WithRelease(release bool) Option {
	return func(h *Hook) {
		h.Release = release
	}
}
