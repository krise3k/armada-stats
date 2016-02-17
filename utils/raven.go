package utils

import (
	"log"
	"github.com/getsentry/raven-go"
)

func GetRaven() *raven.Client {
	sentryUrl, _ := Config.String("sentry_url")
	client, err := raven.New(sentryUrl)
	if err != nil {
		log.Printf("Can't initialize raven %s", err)
	}

	return client
}
