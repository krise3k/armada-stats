package utils

import (
	"github.com/getsentry/raven-go"
	"log"
)

func GetRaven() *raven.Client {
	sentryUrl, _ := Config.String("sentry_url")
	client, err := raven.New(sentryUrl)
	if err != nil {
		log.Printf("Can't initialize raven %s", err)
	}

	version, _ := Config.String("version")
	client.SetRelease(version)

	return client
}
