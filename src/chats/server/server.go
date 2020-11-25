package server

import (
	"chats/application"
	"chats/models"
	"chats/system"
	"context"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"log"
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
	logs                *system.Logs
	actualAccounts      map[uuid.UUID]time.Time
	actualAccountsMutex sync.Mutex
}

func NewServer(app *application.Application) *WsServer {
	return &WsServer{
		apiTopic:       system.BusTopic(),
		port:           os.Getenv("WEBSOCKET_PORT"),
		hub:            NewHub(app),
		logs:           system.Init(),
		shutdownSleep:  getShutdownSleep(),
		actualAccounts: make(map[uuid.UUID]time.Time),
	}
}

func getShutdownSleep() time.Duration {
	timeout, _ := strconv.ParseInt(os.Getenv("SHUTDOWN_SLEEP"), 10, 0)
	return time.Duration(timeout) * time.Second
}

func (ws *WsServer) Run() {
	if system.Cron() {

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
		go ws.Consumer()

		// send actual accounts to CRON topic
		go ws.provider()

		// http server
		ws.listenAndServe()

	}
}

func (ws *WsServer) Shutdown(ctx context.Context) {

	// TODO: close grpc

	if system.Cron() {
		ws.hub.app.Sdk.Shutdown()
		log.Println("nats connection has been closed")
	} else {
		ws.hub.removeAllSessions()
		log.Println("all wsSessions has been removed")

		ws.hub.app.Sdk.Shutdown()
		log.Println("nats connection has been closed")

		_ = ws.httpServer.server.Shutdown(ctx)
		log.Println("HTTP server has been closed")

		ws.grpcServer.GracefulStop()
		log.Println("gRPC server has been stopped")

	}

	time.Sleep(ws.shutdownSleep)

	// TODO: in v2 there is no Close method
	//err := ws.hub.app.DB.Instance.Close()
	//if err != nil {
	//	log.Println("db connection has been closed with sentry")
	//} else {
	//	log.Println("db connection has been closed")
	//}

	err := ws.hub.app.DB.Redis.Instance.Close()
	if err != nil {
		log.Println("redis connection has been closed with sentry")
	} else {
		log.Println("redis connection has been closed")
	}

	system.ErrHandler.Close()
	log.Println("sentry connection has been closed")

	if !system.Cron() {
		err = ws.httpServer.server.Shutdown(ctx)
		if err != nil {
			log.Println("sentry connection has been shutdown with sentry")
		}
	}
}

func (ws *WsServer) Consumer() {
	dataChan := make(chan []byte, 1024)

	go ws.hub.app.Sdk.
		Subject(system.InsideTopic()).
		Consumer(dataChan)

	for {
		data := <-dataChan
		log.Println("Consumer data:", string(data))
		message := &RoomMessage{}
		err := json.Unmarshal(data, message)
		if err != nil {
			system.ErrHandler.SetError(&system.Error{
				Error:   err,
				Message: system.UnmarshallingError,
				Code:    system.UnmarshallingErrorCode,
				Data:    data,
			})
		}

		if message.RoomId != uuid.Nil {
			log.Println("Message to room ", message.RoomId)
			if room, ok := ws.hub.rooms[message.RoomId]; ok {
				log.Printf("Room found: %s \n", room.roomId)
				answer, err := json.Marshal(message.Message)
				if err != nil {
					system.ErrHandler.SetError(&system.Error{
						Error:   err,
						Message: WsCreateClientResponse,
						Code:    WsCreateClientResponseCode,
						Data:    []byte("Type: " + message.Message.Type),
					})
				} else {
					sessionIds := room.getRoomSessionIds()
					log.Println("- Consumer sessionIds cnt:", len(sessionIds))
					log.Println("- GetRoom sessionIds cnt:", len(room.subscribers))
					for _, uniqueId := range sessionIds {
						if session, ok := ws.hub.sessions[uniqueId]; ok {
							go ws.hub.sendMessage(session, answer)
						}
					}
				}
			}

		} else if message.RoomId == uuid.Nil && message.AccountId != uuid.Nil {
			fmt.Println("- Consumer message to client")
			answer, err := json.Marshal(message.Message)
			if err != nil {
				system.ErrHandler.SetError(&system.Error{
					Error:   err,
					Message: WsCreateClientResponse,
					Code:    WsCreateClientResponseCode,
					Data:    []byte("Type: " + message.Message.Type),
				})
			} else {
				if client, ok := ws.hub.accountSessions[message.AccountId]; ok {
					go ws.hub.sendMessage(client, answer)
					log.Println("NATS: Send WS Message") //	TODO
				} else if message.SendPush {
					/*pushMessage := &sdk.ApiUserPushResponse{
						Type: message.Message.Type,
						Data: message.Message.Data,
					}
					go ws.hub.app.Sdk.UserPush(message.AccountId, pushMessage)
					log.Println("NATS: Send Push Message")*///	TODO
				}
			}
		} else if message.RoomId == uuid.Nil && message.AccountId == uuid.Nil {

			switch message.Message.Type {
			case system.SystemMsgTypeUserSubscribe:
				log.Println("System message (user subscribing): ", *message.Message)
				messageData := &models.WSSystemUserSubscribeRequest{}
				err := json.Unmarshal(data, messageData)
				if err != nil {
					system.ErrHandler.SetError(&system.Error{
						Error:   err,
						Message: WsCreateClientResponse,
						Code:    WsCreateClientResponseCode,
						Data:    []byte("Type: " + message.Message.Type),
					})
				} else {
					if session, ok := ws.hub.accountSessions[messageData.Message.Data.AccountId]; ok {

						log.Printf("Session %s found by account %s", session.sessionId, messageData.Message.Data.AccountId.String())

						// search for subscribers by session account
						subscribers := ws.hub.app.DB.GetAccountSubscribers(messageData.Message.Data.AccountId)
						log.Printf("Subscribers found: %v", subscribers)

						ws.hub.accountSessions[messageData.Message.Data.AccountId].SetSubscribers(subscribers)

						//	update room
						room := ws.hub.LoadRoomIfNotExists(messageData.Message.Data.RoomId)
						room.AddSession(session.sessionId)
						session.rooms[messageData.Message.Data.RoomId] = room

						log.Println("account " + messageData.Message.Data.AccountId.String() + " added to room")

					} else if room, ok := ws.hub.rooms[messageData.Message.Data.RoomId]; ok {
						subscribers := ws.hub.app.DB.GetRoomAccountSubscribers(messageData.Message.Data.RoomId)
						room.UpdateSubscribers(subscribers)
						log.Println("room " + messageData.Message.Data.RoomId.String() + " updated subscribers")

					}
				}
				break
			case system.SystemMsgTypeUserUnsubscribe:
				log.Println("WS system message: user unsubscribed", *message.Message)
				messageData := &models.WSSystemUserUnsubscribeRequest{}
				err := json.Unmarshal(data, messageData)
				if err != nil {
					system.ErrHandler.SetError(&system.Error{
						Error:   err,
						Message: WsCreateClientResponse,
						Code:    WsCreateClientResponseCode,
						Data:    []byte("Type: " + message.Message.Type),
					})
				} else {
					if cli, ok := ws.hub.accountSessions[messageData.Message.Data.AccountId]; ok {
						if room, ok := cli.rooms[messageData.Message.Data.RoomId]; ok {
							room.removeSession(cli)
						}
					}
				}
				break
			}
		}
	}
}

func createResponse(response *models.WSChatErrorResponse) []byte {
	result, err := json.Marshal(response)

	if err != nil {
		system.ErrHandler.SetError(&system.Error{
			Error:   err,
			Message: WsCreateClientResponse,
			Code:    WsCreateClientResponseCode,
		})
	}

	return result
}

