package application

import (
	"chats/database"
	"chats/system"
	"chats/sdk"
	"log"
	"strconv"
)

type Application struct {
	Sdk   *sdk.Sdk
	DB    *database.Storage
	Sentry *system.Sentry

}

func Init() *Application {

	sentry, err := system.InitSentry(nil)
	system.ErrHandler.Sentry = sentry

	if err != nil {
		log.Fatalf("Sentry initialization error: %s", err)
	}

	return &Application{
		Sdk:   sdkInit(0),
		DB:    database.Init(),
		Sentry: sentry,
	}
}

func sdkInit(attempt uint) (init *sdk.Sdk) {
	init, err := sdk.Init(system.SdkOptions())
	if err != nil {
		system.ErrHandler.SetError(&system.Error{
			Error:   err,
			Message: system.SdkConnectionError + "; attempt: " + strconv.FormatUint(uint64(attempt), 10),
			Code:    system.SdkConnectionErrorCode,
		})

		system.Reconnect(system.SdkConnectionError, &attempt)

		return sdkInit(attempt)
	}

	return init
}
