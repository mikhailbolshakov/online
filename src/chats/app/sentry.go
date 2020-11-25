package app

import (
	"chats/system"
	"fmt"
	"github.com/getsentry/raven-go"
	"os"
	"strconv"
	"time"
)

type Sentry struct {
	Location *time.Location
}

type SentryOptions struct {
	Protocol string
	Login string
	Password string
	Url string
	Port string
	ProjectId string
	ReconnectTime time.Duration
}

func (opt *SentryOptions) IsEnabled() bool {
	return opt.Url != ""
}

func (opt *SentryOptions) DSN() string {
	return fmt.Sprintf("%s://%s:%s@%s:%s/%s",
		opt.Protocol,
		opt.Login,
		opt.Password,
		opt.Url,
		opt.Port,
		opt.ProjectId)
}

func getFromEnv() *SentryOptions {
	return &SentryOptions{
		Protocol:  os.Getenv("SENTRY_PROTOCOL"),
		Login:     os.Getenv("SENTRY_LOGIN"),
		Password:  os.Getenv("SENTRY_PASSWORD"),
		Url:       os.Getenv("SENTRY_URL"),
		Port:      os.Getenv("SENTRY_PORT"),
		ProjectId: os.Getenv("SENTRY_PROJECT_ID"),
		ReconnectTime: reconnectTime(),
	}
}

func InitSentry(options *SentryOptions) (*Sentry, error) {

	opt := options

	if opt == nil {
		opt = getFromEnv()
	}

	if opt.IsEnabled() {
		err := raven.SetDSN(opt.DSN())
		if err != nil {
			L().Debug(err)
			L().Debug("reconnect after " + opt.ReconnectTime.String() + " seconds...")
			time.Sleep(opt.ReconnectTime)
			return InitSentry(opt)
		}
		L().Debug("Sentry connected.")

		loc, err := getLocation()
		if err != nil {
			return nil, err
		}

		return &Sentry{
			Location: loc,
		}, nil
	} else {
		L().Debug("Sentry disabled.")
		return nil, nil
	}

}

func (s *Sentry) Close() {
	raven.Close()
}

func getLocation() (*time.Location, error) {
	return time.LoadLocation("Europe/Moscow")
}

func (s *Sentry) SetError(err *system.Error) {
	raven.CaptureError(err.Error, map[string]string{
		"Message": err.Message,
		"Code":    strconv.Itoa(err.Code),
		"Data":    string(err.Data),
	})
	L().Debug("Error message:", err.Message, "Code:", err.Code, "Data:", string(err.Data))
}

func (s *Sentry) SetPanic(f func()) {
	raven.CapturePanic(f, map[string]string{
		"Message": system.GetError(system.ApplicationErrorCode),
		"Code":    strconv.Itoa(system.ApplicationErrorCode),
	})
}