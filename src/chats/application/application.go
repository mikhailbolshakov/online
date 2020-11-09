package application

import (
	"chats/database"
	"chats/models"
	"chats/sentry"
	"chats/service"
	"gitlab.medzdrav.ru/health-service/go-sdk"
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
	init, err := sdk.Init(service.SdkOptions())
	if err != nil {
		service.SetError(&models.SystemError{
			Error:   err,
			Message: service.SdkConnectionError + "; attempt: " + strconv.FormatUint(uint64(attempt), 10),
			Code:    service.SdkConnectionErrorCode,
		})

		service.Reconnect(service.SdkConnectionError, &attempt)

		return sdkInit(attempt)
	}

	return init
}
