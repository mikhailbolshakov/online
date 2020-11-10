package infrastructure

import (
	"chats/sdk"
	"chats/sentry"
)

/**
websocket 100
service	200
database 300
nats 400
redis 500
sentry 600
*/

const (
	ApplicationError        = "Общая ошибка приложения"
	ApplicationErrorCode    = 200
	UnmarshallingError      = "Ошибка при анмаршаллинге"
	UnmarshallingErrorCode  = 201
	ParseError              = "Ошибка при парсинге"
	ParseErrorCode          = 205
	LoadLocationError       = "Ошибка установки временной зоны"
	LoadLocationErrorCode   = 210
	MessageTooLongError     = "Длина сообщения превышает установленный лимит"
	MessageTooLongErrorCode = 220
	SdkConnectionError      = "Ошибка соединения с шиной"
	SdkConnectionErrorCode  = 230
	CronResponseError       = "Ошибка ответа от cron"
	CronResponseErrorCode   = 301
)

func MarshalError1011(err error, params []byte) *sentry.SystemError {
	return &sentry.SystemError{
		Error:   err,
		Message: sdk.GetError(1011),
		Code:    1011,
		Data:    params,
	}
}

func UnmarshalError1010(err error, params []byte) *sentry.SystemError {
	return abstractError(&sentry.SystemError{
		Error: err,
		Data:  params,
	}, 1010)
}

func UnmarshalRequestError1201(err error, params []byte) *sentry.SystemError {
	return abstractError(&sentry.SystemError{
		Error: err,
		Data:  params,
	}, 1201)
}

func UnmarshalRequestTypeError1204(err error, params []byte) *sentry.SystemError {
	return abstractError(&sentry.SystemError{
		Error: err,
		Data:  params,
	}, 1204)
}

func abstractError(err *sentry.SystemError, code int) *sentry.SystemError {
	err.Message = sdk.GetError(code)
	err.Code = code
	return err
}
