package utils

import (
	"github.com/sirupsen/logrus"
	"github.com/evalphobia/logrus_sentry"
)

var logger *logrus.Logger

func GetLogger() *logrus.Logger {
	if logger == nil {
		initLogger()
	}

	return logger
}

func initLogger() {
	logger = logrus.New()
	logger.Formatter = &logrus.TextFormatter{FullTimestamp: true}

	raven := GetRaven()
	hook, err := logrus_sentry.NewWithClientSentryHook(raven, []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
	})

	hook.StacktraceConfiguration.Enable = true

	if err == nil {
		logger.Hooks.Add(hook)
	}
}
