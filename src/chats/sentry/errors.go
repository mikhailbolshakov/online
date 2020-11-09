package sentry

import (
	"github.com/getsentry/raven-go"
	"log"
	"strconv"
)

const appError = "Общая ошибка приложения"
const appErrorCode = 200

func SetError(err *SystemError) {
	raven.CaptureError(err.Error, map[string]string{
		"Message": err.Message,
		"Code":    strconv.Itoa(err.Code),
		"Data":    string(err.Data),
	})

	log.Println("Tools.Error message:", err.Message, "Code:", err.Code, "Data:", string(err.Data))
}

func SetPanic(f func()) {
	raven.CapturePanic(f, map[string]string{
		"Message": appError,
		"Code":    strconv.Itoa(appErrorCode),
	})
}
