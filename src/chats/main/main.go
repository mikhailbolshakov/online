package main

import (
	"chats/application"
	"chats/server"
	"chats/system"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	appVersion = "0.0.1"
)

func main() {
	system.SetEnvironment()

	app := application.Init()

	f := func() {

		HelloWorld() //	todo

		wsServer := server.NewServer(app)
		go wsServer.Run()

		quit := make(chan os.Signal)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		wsServer.Shutdown(ctx)
	}

	system.ErrHandler.SetPanic(f)

}

func HelloWorld() {
	fmt.Printf("Service Chats started, version: %s \n", appVersion)
	fmt.Printf("Listen topic: %s \n", system.BusTopic())
	fmt.Printf("Listen inside topic: %s \n", system.InsideTopic())
	fmt.Printf("Listen cron topic: %s \n", system.CronTopic())
}
