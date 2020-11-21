package system

import (
	"fmt"
	"github.com/getsentry/raven-go"
	"log"
	"os"
	"strconv"
	"time"
)

type Sentry struct {
	Location *time.Location
}

type Options struct {
	Protocol string
	Login string
	Password string
	Url string
	Port string
	ProjectId string
	ReconnectTime time.Duration
}

func (opt *Options) IsEnabled() bool {
	return opt.Url != ""
}

func (opt *Options) DSN() string {
	return fmt.Sprintf("%s://%s:%s@%s:%s/%s",
		opt.Protocol,
		opt.Login,
		opt.Password,
		opt.Url,
		opt.Port,
		opt.ProjectId)
}

func getFromEnv() *Options {
	return &Options{
		Protocol:  os.Getenv("SENTRY_PROTOCOL"),
		Login:     os.Getenv("SENTRY_LOGIN"),
		Password:  os.Getenv("SENTRY_PASSWORD"),
		Url:       os.Getenv("SENTRY_URL"),
		Port:      os.Getenv("SENTRY_PORT"),
		ProjectId: os.Getenv("SENTRY_PROJECT_ID"),
		ReconnectTime: ReconnectTime(),
	}
}

func InitSentry(options *Options) (*Sentry, error) {

	opt := options

	if opt == nil {
		opt = getFromEnv()
	}

	if opt.IsEnabled() {
		err := raven.SetDSN(opt.DSN())
		if err != nil {
			log.Println(err)
			log.Println("reconnect after " + opt.ReconnectTime.String() + " seconds...")
			time.Sleep(opt.ReconnectTime)
			return InitSentry(opt)
		}
		log.Println("Sentry connected.")

		loc, err := getLocation()
		if err != nil {
			return nil, err
		}

		return &Sentry{
			Location: loc,
		}, nil
	} else {
		log.Println("Sentry disabled.")
		return nil, nil
	}

}

func (s *Sentry) Close() {
	raven.Close()
}

func getLocation() (*time.Location, error) {
	return time.LoadLocation("Europe/Moscow")
}

func (s *Sentry) SetError(err *Error) {
	raven.CaptureError(err.Error, map[string]string{
		"Message": err.Message,
		"Code":    strconv.Itoa(err.Code),
		"Data":    string(err.Data),
	})
	log.Println("Error message:", err.Message, "Code:", err.Code, "Data:", string(err.Data))
}

func (s *Sentry) SetPanic(f func()) {
	raven.CapturePanic(f, map[string]string{
		"Message": ApplicationError,
		"Code":    strconv.Itoa(ApplicationErrorCode),
	})
}