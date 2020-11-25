package system

import (
	"chats/sdk"
	"log"
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

type ErrorHandler struct {
	Sentry *Sentry
}

var ErrHandler = &ErrorHandler{}

type Error struct {
	Error   error  `json:"sentry"`
	Message string `json:"message"`
	Code    int    `json:"code"`
	Data    []byte `json:"data"`
	Stack   string `json:"stack"`
}

func E(ee error) *Error {
	return &Error{Error: ee, Message: ee.Error()}
}

func MarshalError1011(err error, params []byte) *Error {
	return &Error{
		Error:   err,
		Message: sdk.GetError(1011),
		Code:    1011,
		Data:    params,
	}
}

func UnmarshalError1010(err error, params []byte) *Error {
	return abstractError(&Error{
		Error: err,
		Data:  params,
	}, 1010)
}

func UnmarshalRequestError1201(err error, params []byte) *Error {
	return abstractError(&Error{
		Error: err,
		Data:  params,
	}, 1201)
}

func UnmarshalRequestTypeError1204(err error, params []byte) *Error {
	return abstractError(&Error{
		Error: err,
		Data:  params,
	}, 1204)
}

func abstractError(err *Error, code int) *Error {
	err.Message = sdk.GetError(code)
	err.Code = code
	return err
}

func (h *ErrorHandler) SetError(err *Error) {

	if h.Sentry != nil {
		h.Sentry.SetError(err)
	} else {
		log.Println("Tools.Error message:", err.Message, "Code:", err.Code, "Data:", string(err.Data))
	}
}

func (h *ErrorHandler) SetPanic(f func()) {

	if h.Sentry != nil {
		h.Sentry.SetPanic(f)
	} else {
		defer func() {
			if err := recover(); err != nil {
				log.Println("Panic!!!!", err)
			}
		}()
		f()
	}
}

func (h *ErrorHandler) Close() {
	if h.Sentry != nil {
		h.Sentry.Close()
	}
}
