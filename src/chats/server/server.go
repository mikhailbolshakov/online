package server

import (
	"chats/application"
	"chats/models"
	"chats/system"
	"context"
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var Server *WsServer

func NewServer(app *application.Application) *WsServer {
	Server = &WsServer{
		apiTopic:       system.BusTopic(),
		port:           os.Getenv("WEBSOCKET_PORT"),
		hub:            NewHub(app),
		logs:           system.Init(),
		shutdownSleep:  getShutdownSleep(),
		actualAccounts: make(map[uuid.UUID]time.Time),
	}
	return Server
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

		// listens to NATS topics
		go ws.ApiConsumer()

		// listens internal topic
		go ws.Consumer()

		go ws.provider()

		srv := &http.Server{Addr: ws.port}
		http.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
			log.Println("request:" + r.Host + r.URL.EscapedPath())
			IndexAction(ws.hub, w, r)
		})

		//	testing. not for production.
		http.HandleFunc("/ws/html/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "src/chats/main/home.html")
		})

		ws.server = srv

		log.Fatal(srv.ListenAndServe())
	}
}

func (ws *WsServer) Shutdown(ctx context.Context) {
	if system.Cron() {
		ws.hub.app.Sdk.Shutdown()
		log.Println("nats connection has been closed")
	} else {
		ws.hub.removeAllSessions()
		log.Println("all wsSessions has been removed")

		ws.hub.app.Sdk.Shutdown()
		log.Println("nats connection has been closed")
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
		err = ws.server.Shutdown(ctx)
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
		fmt.Println("- Consumer data:", string(data)) //	TODO
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
			fmt.Println("- Consumer message to room", message.RoomId)
			if room, ok := ws.hub.rooms[message.RoomId]; ok {
				fmt.Println("- Consumer room exists")
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
					fmt.Println("- Consumer sessionIds cnt:", len(sessionIds))
					fmt.Println("- GetRoom sessionIds cnt:", len(room.subscribers))
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
			//	system message ws only
			fmt.Println(" -----> system message ws only!", *message.Message)
			switch message.Message.Type {
			case system.SystemMsgTypeUserSubscribe:
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
						//	update session
						cliSubscribes := ws.hub.app.DB.GetAccountSubscribers(messageData.Message.Data.AccountId)
						ws.hub.accountSessions[messageData.Message.Data.AccountId].SetSubscribers(cliSubscribes)

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

func IndexAction(h *Hub, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	//	upgrade websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		system.ErrHandler.SetError(&system.Error{
			Error:   err,
			Message: WsUpgradeProblem,
			Code:    WsUpgradeProblemCode,
		})

		return
	}

	//	userToken
	token := r.URL.Query().Get("token")
	if token == "" {
		response := &models.WSChatErrorResponse{
			Error: models.WSChatErrorErrorResponse{
				Message: WsEmptyToken,
				Code:    WsEmptyTokenCode,
			},
		}
		w.Write(createResponse(response))
		system.ErrHandler.SetError(&system.Error{
			Error:   nil,
			Message: WsEmptyToken,
			Code:    WsEmptyTokenCode,
			Data:    []byte("token: " + token),
		})

		return
	}

	//	account
	//accountModel := &sdk.AccountModel{}
	//sdkErr := h.app.Sdk.UserByToken(token, accountModel)
	account, sentryErr := h.app.DB.GetAccount(uuid.FromStringOrNil(token), "")
	if sentryErr != nil || account.Id == uuid.Nil {
		response := &models.WSChatErrorResponse{
			Error: models.WSChatErrorErrorResponse{
				Message: WsUserIdentification,
				Code:    WsUserIdentificationCode,
			},
		}
		w.Write(createResponse(response))
		system.ErrHandler.SetError(&system.Error{
			Error:   nil,
			Message: WsUserIdentification,
			Code:    WsUserIdentificationCode,
			Data:    []byte("token: " + token),
		})

		return
	}

	session := InitSession(h, conn)

	//	rooms & subscribes
	rooms := make(map[uuid.UUID]*Room)
	subscribes := h.app.DB.GetAccountSubscribers(account.Id)
	for roomId, _ := range subscribes {
		room := h.LoadRoomIfNotExists(roomId)
		room.AddSession(session.sessionId)
		rooms[roomId] = room
	}
	session.account = ConvertAccountFromModel(account)
	session.rooms = rooms
	session.subscribes = subscribes

	h.registerChan <- session
	go session.Write()
	go session.Read()
}
