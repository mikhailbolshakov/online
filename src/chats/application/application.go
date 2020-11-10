package application

import (
	"chats/database"
	"chats/sentry"
	"chats/infrastructure"
	"chats/sdk"
	"strconv"
)

type Application struct {
	Sdk   *sdk.Sdk
	DB    *database.Storage
	Sentry *sentry.Sentry
}

func Init(sentry *sentry.Sentry) *Application {
	return &Application{
		Sdk:   sdkInit(0),
		DB:    database.Init(),
		Sentry: sentry,
	}
}

func sdkInit(attempt uint) (init *sdk.Sdk) {
	init, err := sdk.Init(infrastructure.SdkOptions())
	if err != nil {
		infrastructure.SetError(&sentry.SystemError{
			Error:   err,
			Message: infrastructure.SdkConnectionError + "; attempt: " + strconv.FormatUint(uint64(attempt), 10),
			Code:    infrastructure.SdkConnectionErrorCode,
		})

		infrastructure.Reconnect(infrastructure.SdkConnectionError, &attempt)

		return sdkInit(attempt)
	}

	return init
}
