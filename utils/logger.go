package utils

import (
	"github.com/Sirupsen/logrus"
	"github.com/evalphobia/logrus_sentry"
)

func GetLogger() *logrus.Logger {
	raven := GetRaven()

	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{FullTimestamp: true}

	hook, err := logrus_sentry.NewWithClientSentryHook(raven, []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
	})

	hook.StacktraceConfiguration.Enable = true

	if err == nil {
		log.Hooks.Add(hook)
	}
	return log
}