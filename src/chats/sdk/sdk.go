package sdk

import (
	"encoding/json"
	"fmt"
	gonats "github.com/nats-io/go-nats"
	"log"
	"runtime"
	"time"
)

type Msg = gonats.Msg

type Options struct {
	Url     string
	Token   string
	Timeout time.Duration
	Log     bool
}

type Sdk struct {
	connection *gonats.Conn
	timeout    time.Duration
	subject    string
	log        bool
}

func (ar *ApiRequest) String() string {
	return "method: " + ar.Method + "; path: " + ar.Path
}

//	Initialize NATS connection
func Init(opt *Options) (*Sdk, error) {
	sdk := &Sdk{
		timeout: opt.Timeout,
		log:     opt.Log,
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
		sdk.errorLog(sdk.Error(err, 1001, []byte(
			"url: "+opt.Url+
				"; duration: "+time.Duration.String(opt.Timeout)+
				"; token: "+opt.Token)))
		return nil, err
	}

	sdk.connection = natsConn

	return sdk, nil
}

//	Shutdown NATS connection
func (sdk *Sdk) Shutdown() {
	sdk.connection.Close()
}

//	Return response
func (sdk *Sdk) request(data []byte) ([]byte, *Error) {
	if len(sdk.subject) == 0 {
		return nil, sdk.errorLog(sdk.Error(nil, 1101, data))
	}

	msg, err := sdk.connection.Request(sdk.subject, data, sdk.timeout)
	if err != nil {
		return nil, sdk.errorLog(sdk.Error(err, 1011, data))
	}
	if msg == nil {
		return nil, sdk.errorLog(sdk.Error(err, 1012, data))
	}

	return msg.Data, nil
}

//	Return response
func (sdk *Sdk) Request(data []byte) ([]byte, *Error) {
	return sdk.request(data)
}

//	Without return
func (sdk *Sdk) Publish(data []byte) *Error {
	publishError := sdk.connection.Publish(sdk.subject, data)
	if publishError != nil {
		return sdk.errorLog(sdk.Error(publishError, 1201, data))
	}

	return nil
}

//	Set subject
func (sdk *Sdk) Subject(subject string) *Sdk {
	sdk.subject = subject
	return sdk
}

//	Consumer for API
func (sdk *Sdk) ApiConsumer(handle func([]byte) ([]byte, *Error), errorChan chan *Error) {
	const queue = "apiConsumer"

	for {
		conn := sdk.connection
		conn.QueueSubscribe(sdk.subject, queue, func(msg *gonats.Msg) {
			sdk.setlog(">>> Msg NATS API request on [%s]: %s\n", msg.Subject, msg.Data)
			response, err := handle(msg.Data)
			sdk.setlog(">>> Msg NATS API response on [%s]: %s\n", msg.Reply, response)

			if err != nil {
				errorChan <- err
				result := &ApiErrorResponse{
					Error: ApiErrorResponseError{
						Message: err.Message,
						Code:    err.Code,
					},
				}
				response, _ = json.Marshal(result)
			}

			pubError := conn.Publish(msg.Reply, response)
			if pubError != nil {
				errorChan <- sdk.errorLog(sdk.Error(pubError, 1011, response))
			}
		})

		runtime.Goexit()
	}
}

func (sdk *Sdk) CronConsumer(handle func([]byte) ([]byte, *Error), errorChan chan *Error) {
	const queue = "cronConsumer"

	for {
		conn := sdk.connection
		conn.QueueSubscribe(sdk.subject, queue, func(msg *gonats.Msg) {
			sdk.setlog(">>> Msg NATS CRON request on [%s]: %s\n", msg.Reply, msg.Data)
			response, err := handle(msg.Data)
			sdk.setlog(">>> Msg NATS CRON response on [%s]: %s\n", msg.Reply, response)

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
					errorChan <- sdk.errorLog(sdk.Error(pubError, 1011, response))
				}
			}
		})

		runtime.Goexit()
	}
}

//	Consumer for WS
func (sdk *Sdk) Consumer(dataChan chan<- []byte) {
	for {
		conn := sdk.connection
		conn.Subscribe(sdk.subject, func(msg *gonats.Msg) {
			sdk.setlog("Msg NATS WS request on [%s]: %s\n", msg.Subject, msg.Data)
			dataChan <- msg.Data
		})

		runtime.Goexit()
	}
}

//	logger
func (sdk *Sdk) setlog(msg string, sbj interface{}, data []byte) {
	if sdk.log {
		log.Printf(msg, sbj, string(data))
	}
}

func (sdk *Sdk) errorLog(err *Error) *Error {
	if sdk.log {
		log.Println(err.Error, err.Message, err.Code, string(err.Data))
	}

	return err
}

func (sdk *Sdk) catchPanic(method string) {
	if r := recover(); r != nil {
		_, ok := r.(error)
		if !ok {
			err := fmt.Errorf("%v", r)
			sdk.setlog("PANIC "+method+":"+err.Error(), err, []byte{})
		}
	}
}
