package app

import (
	"os"
	"strconv"
	"time"
)

type Env struct {}

func (e *Env) Cron() bool {
	return os.Getenv("CRON") == "1"
}

func (e *Env) CronStep() time.Duration {
	num := os.Getenv("CRON_STEP")
	cronStep, err := strconv.ParseInt(num, 10, 0)
	if err != nil {
		cronStep = defaultCronStep
	}

	return time.Duration(cronStep) * time.Second
}

