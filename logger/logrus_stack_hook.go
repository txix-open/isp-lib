package logger

import (
	"github.com/sirupsen/logrus"
)

func defaultHook() logrusStackHook {
	withCaller := make(map[logrus.Level]bool, len(logrus.AllLevels))
	for _, level := range logrus.AllLevels {
		withCaller[level] = true
	}

	withStack := make(map[logrus.Level]bool, 3)
	withStack[logrus.FatalLevel] = true
	withStack[logrus.PanicLevel] = true
	withStack[logrus.ErrorLevel] = true

	return logrusStackHook{
		CallerLevels: withCaller,
		StackLevels:  withStack,
	}
}

type logrusStackHook struct {
	CallerLevels map[logrus.Level]bool
	StackLevels  map[logrus.Level]bool
}

func (hook logrusStackHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook logrusStackHook) Fire(entry *logrus.Entry) error {
	printCaller, _ := hook.CallerLevels[entry.Level]
	printStack, _ := hook.StackLevels[entry.Level]
	if !printCaller && !printStack {
		return nil
	}

	frames := callers(10)
	hasFrames := len(frames) > 0
	if printCaller && hasFrames {
		entry.Data["caller"] = frames[0]
	} else {
		delete(entry.Data, "caller")
	}

	if printStack && hasFrames {
		entry.Data["stack"] = frames
	} else {
		delete(entry.Data, "stack")
	}

	return nil
}
