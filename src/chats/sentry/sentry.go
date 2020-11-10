package sentry

import (
	"github.com/getsentry/raven-go"
	"log"
	"time"
)

type Params struct {
	Sentry        string
	ReconnectTime time.Duration
}

type Sentry struct {
	Location *time.Location
}

func Init(params Params) (*Sentry, error) {
	err := raven.SetDSN(params.Sentry)
	if err != nil {
		log.Println(err)
		log.Println("reconnect after " + params.ReconnectTime.String() + " seconds...")
		time.Sleep(params.ReconnectTime)
		return Init(params)
	}
	log.Println("Sentry connected.")

	loc, err := getLocation()
	if err != nil {
		return nil, err
	}

	return &Sentry{
		Location: loc,
	}, nil
}

func Close() {
	raven.Close()
}

func getLocation() (*time.Location, error) {
	return time.LoadLocation("Europe/Moscow")
}
