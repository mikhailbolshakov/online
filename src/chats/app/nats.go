package app

import (
	"chats/system"
	"encoding/json"
	"fmt"
	gonats "github.com/nats-io/go-nats"
	"os"
	"runtime"
	"strconv"
	"time"
)

const Timeout = 10
const defaultReconnectTime = 15
const defaultCronStep = 10

type Msg = gonats.Msg

type natsOptions struct {
	Url     string
	Token   string
	Timeout time.Duration
	Log     bool
}

type Nats struct {
	Connection *gonats.Conn
	Timeout    time.Duration
	Subj       string
	Log        bool
}

//func (ar *ApiRequest) String() string {
//	return "method: " + ar.Method + "; path: " + ar.Path
//}

func getNatsOptions() *natsOptions {

	busTimeout := os.Getenv("BUS_TIMEOUT")
	timeout, err := strconv.ParseInt(busTimeout, 10, 0)
	if err != nil {
		timeout = Timeout
	}

	return &natsOptions{
		Url:     os.Getenv("BUS_URL"),
		Token:   os.Getenv("BUS_TOKEN"),
		Timeout: time.Duration(timeout) * time.Millisecond,
		Log:     os.Getenv("SDK_LOG") == "1",
	}
}

func initNats(attempt uint) *Nats {

	n, err := initNatsInternal(getNatsOptions())
	if err != nil {
		Instance.ErrorHandler.SetError(system.SysErr(err, system.SdkConnectionErrorCode, nil))
		reconnect(system.GetError(system.SdkConnectionErrorCode), &attempt)
		return initNats(attempt)
	}

	fmt.Printf("Listen topic: %s \n", n.BusTopic())
	fmt.Printf("Listen inside topic: %s \n", n.InsideTopic())
	fmt.Printf("Listen cron topic: %s \n", n.CronTopic())

	return n
}

func initNatsInternal(opt *natsOptions) (*Nats, error) {

	nats := &Nats{
		Timeout: opt.Timeout,
		Log:     opt.Log,
	}

	options := gonats.Options{
		Url:            opt.Url,
		Token:          opt.Token,
		AllowReconnect: true,
		ReconnectWait:  5 * time.Second,
		MaxReconnect:   100,
	}

	natsConn, err := options.Connect()
	if err != nil {
		Instance.ErrorHandler.SetError(&system.Error{
			Error:   err,
			Message: fmt.Sprintf("url: %s, duration: %d, token: %s", opt.Url, time.Duration.String(opt.Timeout), opt.Token),
		})
		return nil, err
	}

	nats.Connection = natsConn

	return nats, nil
}

//	Shutdown NATS connection
func (n *Nats) Shutdown() {
	n.Connection.Close()
}

//	Return response
func (n *Nats) request(data []byte) ([]byte, *system.Error) {

	if len(n.Subj) == 0 {
		return nil, Instance.ErrorHandler.SetError(system.SysErr(nil, 1101, data))
	}

	msg, err := n.Connection.Request(n.Subj, data, n.Timeout)
	if err != nil {
		return nil, Instance.ErrorHandler.SetError(system.SysErr(nil, 1011, data))
	}
	if msg == nil {
		return nil, Instance.ErrorHandler.SetError(system.SysErr(nil, 1012, data))
	}

	return msg.Data, nil
}

//	Return response
func (n *Nats) Request(data []byte) ([]byte, *system.Error) {
	return n.request(data)
}

//	Without return
func (n *Nats) Publish(data []byte) *system.Error {
	publishError := n.Connection.Publish(n.Subj, data)
	if publishError != nil {
		return Instance.ErrorHandler.SetError(system.SysErr(publishError, 1201, data))
	}

	return nil
}

//	Set subject
func (n *Nats) Subject(subject string) *Nats {
	n.Subj = subject
	return n
}

//	Consumer for WS
func (n *Nats) Consumer(dataChan chan<- []byte) {
	for {
		conn := n.Connection
		_, _ = conn.Subscribe(n.Subj, func(msg *gonats.Msg) {
			n.Setlog("Msg NATS WS request on [%s]: %s\n", msg.Subject, msg.Data)
			dataChan <- msg.Data
		})

		runtime.Goexit()
	}
}

//	logger
func (n *Nats) Setlog(msg string, sbj interface{}, data []byte) {
	if n.Log {
		L().Debugf(msg, sbj, string(data))
	}
}

func (n *Nats) catchPanic(method string) {
	if r := recover(); r != nil {
		_, ok := r.(error)
		if !ok {
			err := fmt.Errorf("%v", r)
			n.Setlog("PANIC "+method+":"+err.Error(), err, []byte{})
		}
	}
}

func (n *Nats) BusTopic() string {
	return os.Getenv("BUS_TOPIC")
}

func (n *Nats) InsideTopic() string {
	return "inside." + n.BusTopic()
}

func (n *Nats) CronTopic() string {
	return "cron." + n.BusTopic()
}

type ApiErrorResponseError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}
// cron sentry response
type CronErrorResponse struct {
	Error CronErrorResponseError
}
type CronErrorResponseError ApiErrorResponseError

func (n *Nats) CronConsumer(handle func([]byte) ([]byte, *system.Error), errorChan chan *system.Error) {
	const queue = "cronConsumer"

	for {
		conn := n.Connection
		conn.QueueSubscribe(n.Subj, queue, func(msg *gonats.Msg) {
			n.Setlog(">>> Msg NATS CRON request on [%s]: %s\n", msg.Reply, msg.Data)
			response, err := handle(msg.Data)
			n.Setlog(">>> Msg NATS CRON response on [%s]: %s\n", msg.Reply, response)

			if err != nil {
				errorChan <- err
				result := &CronErrorResponse{
					Error: CronErrorResponseError{
						Message: err.Message,
						Code:    err.Code,
					},
				}
				response, _ = json.Marshal(result)
			}

			if msg.Reply != "" {
				pubError := conn.Publish(msg.Reply, response)
				if pubError != nil {
					errorChan <- E().SetError(system.SysErr(pubError, 1011, response))
				}
			}
		})

		runtime.Goexit()
	}
}


