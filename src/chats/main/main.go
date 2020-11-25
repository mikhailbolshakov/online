package main

import (
	"chats/app"
	"chats/server"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	app.ApplicationInit()

	f := func() {

		wsServer := server.NewServer(app.Instance)
		go wsServer.Run()

		quit := make(chan os.Signal)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		wsServer.Shutdown(ctx)
	}

	app.E().SetPanic(f)

}

