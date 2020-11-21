package system

import (
	"bufio"
	"chats/sdk"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const Timeout = 10
const defaultReconnectTime = 15
const defaultCronStep = 10

const SystemMsgTypeUserSubscribe = "userSubscribe"
const SystemMsgTypeUserUnsubscribe = "userUnsubscribe"

type Logs struct {
	Print bool
}

func Init() *Logs {
	return &Logs{Print: true}
}

func (log *Logs) Scan() {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		text = strings.Replace(text, "\n", "", -1)

		if strings.Compare("api", text) == 0 {
			fmt.Print("this!") //	TODO
			log.Print = !log.Print
		}
	}
}

func Location() (*time.Location, error) {
	location := "Europe/Moscow"
	return time.LoadLocation(location)
}

func BusTopic() string {
	return os.Getenv("BUS_TOPIC")
}

func SdkOptions() *sdk.Options {
	busTimeout := os.Getenv("BUS_TIMEOUT")
	timeout, err := strconv.ParseInt(busTimeout, 10, 0)
	if err != nil {
		timeout = Timeout
	}

	return &sdk.Options{
		Url:     os.Getenv("BUS_URL"),
		Token:   os.Getenv("BUS_TOKEN"),
		Timeout: time.Duration(timeout) * time.Millisecond,
		Log:     os.Getenv("SDK_LOG") == "1",
	}
}

func ReconnectTime() time.Duration {
	num := os.Getenv("RECONNECT_TO_SERVICE")
	reconnectTime, err := strconv.ParseInt(num, 10, 0)
	if err != nil {
		reconnectTime = defaultReconnectTime
	}

	return time.Duration(reconnectTime) * time.Second
}

func Cron() bool {
	return os.Getenv("CRON") == "1"
}

func CronStep() time.Duration {
	num := os.Getenv("CRON_STEP")
	cronStep, err := strconv.ParseInt(num, 10, 0)
	if err != nil {
		cronStep = defaultCronStep
	}

	return time.Duration(cronStep) * time.Second
}

func InsideTopic() string {
	return "inside." + BusTopic()
}

func CronTopic() string {
	return "cron." + BusTopic()
}

func Reconnect(msg string, attempt *uint) {
	*attempt++
	log.Println("> " + msg)
	log.Println("> Attempt: " + strconv.FormatUint(uint64(*attempt), 10))
	log.Println("> Reconnect after " + ReconnectTime().String() + "...")
	time.Sleep(ReconnectTime())
}
