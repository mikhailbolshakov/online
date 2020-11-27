package server

import (
	"chats/app"
	"context"
	"github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"os"
	"strconv"
	"sync"
	"time"
)

type WsServer struct {
	apiTopic            string
	shutdownSleep       time.Duration
	port                string
	hub                 *Hub
	httpServer          *httpServer
	grpcServer 			*grpc.Server
	actualAccounts      map[uuid.UUID]time.Time
	actualAccountsMutex sync.Mutex
}

var wsServer = &WsServer{}

func NewServer(app *app.Application) *WsServer {
	wsServer = &WsServer{
		apiTopic:       app.Inf.Nats.BusTopic(),
		port:           os.Getenv("WEBSOCKET_PORT"),
		hub:            NewHub(),
		shutdownSleep:  getShutdownSleep(),
		actualAccounts: make(map[uuid.UUID]time.Time),
	}
	return wsServer
}

func getShutdownSleep() time.Duration {
	timeout, _ := strconv.ParseInt(os.Getenv("SHUTDOWN_SLEEP"), 10, 0)
	return time.Duration(timeout) * time.Second
}

func (ws *WsServer) Run() {

	if app.Instance.Env.Cron() {

		// push для непрочитанных сообщений
		go ws.userServiceMessageManager()

		// переводит в offline
		ws.consumer()

	} else {

		// gRPC connection listener
		go ws.Grpc()

		//
		go ws.hub.Run()

		// listens internal topic
		go ws.internalConsumer()

		// send actual accounts to CRON topic
		go ws.provider()

		// http server
		ws.listenAndServe()

	}
}

func (ws *WsServer) Shutdown(ctx context.Context) {

	if app.Instance.Env.Cron() {
		app.Instance.Inf.Nats.Shutdown()
		app.L().Debug("nats connection has been closed")
	} else {
		ws.hub.removeAllSessions()
		app.L().Debug("all wsSessions has been removed")

		app.Instance.Inf.Nats.Shutdown()
		app.L().Debug("nats connection has been closed")

		_ = ws.httpServer.server.Shutdown(ctx)
		app.L().Debug("HTTP server has been closed")

		ws.grpcServer.GracefulStop()
		app.L().Debug("gRPC server has been stopped")

	}

	time.Sleep(ws.shutdownSleep)

	// TODO: in v2 there is no Close method
	//err := app.GetDB().Instance.Close()
	//if err != nil {
	//	app.L().Debug("db connection has been closed with sentry")
	//} else {
	//	app.L().Debug("db connection has been closed")
	//}

	err := app.Instance.Inf.DB.Redis.Instance.Close()
	if err != nil {
		app.L().Debug("redis connection has been closed with sentry")
	} else {
		app.L().Debug("redis connection has been closed")
	}

	app.E().Close()
	app.L().Debug("sentry connection has been closed")

	if !app.Instance.Env.Cron() {
		err = ws.httpServer.server.Shutdown(ctx)
		if err != nil {
			app.L().Debug("sentry connection has been shutdown with sentry")
		}
	}
}



