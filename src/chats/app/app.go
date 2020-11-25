package app

import (
	"chats/system"
	"time"
)

var Instance = &Application{}

type Application struct {
	ErrorHandler *ErrorHandler
	Log *LogHandler
	Inf *Infrastructure
	Env *Env
}

func ApplicationInit() *Application {

	system.SetEnvironment()

	app := &Application{
		Log: &LogHandler{Level: llDebug},
		Env: &Env{},
	}

	Instance = app

	inf, err := infrastructureInit()
	if err != nil {
		panic(err.Error())
	}

	app.Inf = inf
	app.ErrorHandler = &ErrorHandler{
		Sentry: inf.Sentry,
	}

	return app

}

func (a *Application) GetLocation() (*time.Location, error) {
	return time.LoadLocation("Europe/Moscow")
}

func GetDB() *Storage {
	return Instance.Inf.DB
}

func GetNats() *Nats {
	return Instance.Inf.Nats
}

func E() *ErrorHandler {
	return Instance.ErrorHandler
}

func L() *LogHandler {
	return Instance.Log
}

