package main

import (
	"chats/application"
	"chats/server"
	"chats/infrastructure"
	"context"
	"fmt"
	"chats/sentry"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	appVersion = "0.0.1"
)

func main() {
	infrastructure.SetEnvironment()

	snt, err := sentry.Init(sentry.Params{
		Sentry:        infrastructure.SentryOptions(),
		ReconnectTime: infrastructure.ReconnectTime(),
	})
	if err != nil {
		log.Fatal("/!\\ /!\\ /!\\", err)
	}

	sentry.SetPanic(func() {
		app := application.Init(snt)

		HelloWorld() //	todo

		wsserver := server.NewServer(app)
		go wsserver.Run()

		quit := make(chan os.Signal)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		wsserver.Shutdown(ctx)
	})
}

/**
Привет, Мир!
*/
func HelloWorld() {
	fmt.Printf("Service Chats started, version: %s \n", appVersion)
	fmt.Printf("Listen topic: %s \n", infrastructure.BusTopic())
	fmt.Printf("Listen inside topic: %s \n", infrastructure.InsideTopic())
	fmt.Printf("Listen cron topic: %s \n", infrastructure.CronTopic())
}
