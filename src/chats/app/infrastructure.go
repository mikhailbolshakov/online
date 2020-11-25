package app

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Infrastructure struct {
	Nats   *Nats
	DB     *Storage
	Sentry *Sentry
	Logs   *LogHandler
}

func infrastructureInit() (*Infrastructure, error) {

	sentry, err := InitSentry(nil)
	if err != nil {
		log.Fatalf("Sentry initialization error: %s", err.Error())
		return nil, err
	}

	inf := &Infrastructure{
		Nats:   initNats(0),
		DB:     initStorage(),
		Sentry: sentry,
		Logs:   initLogs(),
	}

	return inf, nil
}

func (i *Infrastructure) ReconnectTime() time.Duration {
	return reconnectTime()
}

func (i *Infrastructure) Reconnect(msg string, attempt *uint) {
	reconnect(msg, attempt)
}

func reconnectTime() time.Duration {
	num := os.Getenv("RECONNECT_TO_SERVICE")
	reconnectTime, err := strconv.ParseInt(num, 10, 0)
	if err != nil {
		reconnectTime = defaultReconnectTime
	}

	return time.Duration(reconnectTime) * time.Second
}

func reconnect(msg string, attempt *uint) {
	*attempt++
	L().Debug("> " + msg)
	L().Debug("> Attempt: " + strconv.FormatUint(uint64(*attempt), 10))
	L().Debug("> Reconnect after " + reconnectTime().String() + "...")
	time.Sleep(reconnectTime())
}