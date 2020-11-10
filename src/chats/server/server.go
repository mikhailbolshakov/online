package server

import (
	"chats/application"
	"chats/infrastructure"
	"chats/models"
	"chats/sdk"
	"chats/sentry"
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


func NewServer(app *application.Application) *WsServer {
	return &WsServer{
		apiTopic:       infrastructure.BusTopic(),
		port:           os.Getenv("WEBSOCKET_PORT"),
		hub:            NewHub(app),
		logs:           infrastructure.Init(),
		shutdownSleep:  getShutdownSleep(),
		actualAccounts: make(map[uuid.UUID]time.Time),
	}
}

func getShutdownSleep() time.Duration {
	timeout, _ := strconv.ParseInt(os.Getenv("SHUTDOWN_SLEEP"), 10, 0)
	return time.Duration(timeout) * time.Second
}

func (ws *WsServer) Run() {
	if infrastructure.Cron() {
		go ws.userServiceMessageManager()
		ws.consumer()
	} else {

		go ws.hub.Run()
		go ws.ApiConsumer()
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
	if infrastructure.Cron() {
		ws.hub.app.Sdk.Shutdown()
		log.Println("nats connection has been closed")
	} else {
		ws.hub.removeAllClients()
		log.Println("all clients has been removed")

		ws.hub.app.Sdk.Shutdown()
		log.Println("nats connection has been closed")
	}

	time.Sleep(ws.shutdownSleep)

	err := ws.hub.app.DB.Instance.Close()
	if err != nil {
		log.Println("db connection has been closed with sentry")
	} else {
		log.Println("db connection has been closed")
	}

	err = ws.hub.app.DB.Redis.Instance.Close()
	if err != nil {
		log.Println("redis connection has been closed with sentry")
	} else {
		log.Println("redis connection has been closed")
	}

	sentry.Close()
	log.Println("sentry connection has been closed")

	if !infrastructure.Cron() {
		err = ws.server.Shutdown(ctx)
		if err != nil {
			log.Println("sentry connection has been shutdown with sentry")
		}
	}
}

func (ws *WsServer) Consumer() {
	dataChan := make(chan []byte, 1024)

	go ws.hub.app.Sdk.
		Subject(infrastructure.InsideTopic()).
		Consumer(dataChan)

	for {
		data := <-dataChan
		fmt.Println("- Consumer data:", string(data)) //	TODO
		message := &RoomMessage{}
		err := json.Unmarshal(data, message)
		if err != nil {
			infrastructure.SetError(&sentry.SystemError{
				Error:   err,
				Message: infrastructure.UnmarshallingError,
				Code:    infrastructure.UnmarshallingErrorCode,
				Data:    data,
			})
		}

		if message.RoomId != uuid.Nil {
			fmt.Println("- Consumer message to room", message.RoomId)
			if room, ok := ws.hub.rooms[message.RoomId]; ok {
				fmt.Println("- Consumer room exists")
				answer, err := json.Marshal(message.Message)
				if err != nil {
					infrastructure.SetError(&sentry.SystemError{
						Error:   err,
						Message: WsCreateClientResponse,
						Code:    WsCreateClientResponseCode,
						Data:    []byte("Type: " + message.Message.Type),
					})
				} else {
					subscribers := room.getRoomClientIds()
					fmt.Println("- Consumer subscribers cnt:", len(subscribers))
					fmt.Println("- Chat subscribers cnt:", len(room.subscribers))
					for _, uniqueId := range subscribers {
						if client, ok := ws.hub.clients[uniqueId]; ok {
							go ws.hub.sendMessageToClient(client, answer)
						}
					}
				}
			}

		} else if message.RoomId == uuid.Nil && message.AccountId != uuid.Nil {
			fmt.Println("- Consumer message to client")
			answer, err := json.Marshal(message.Message)
			if err != nil {
				infrastructure.SetError(&sentry.SystemError{
					Error:   err,
					Message: WsCreateClientResponse,
					Code:    WsCreateClientResponseCode,
					Data:    []byte("Type: " + message.Message.Type),
				})
			} else {
				if client, ok := ws.hub.clientsId[message.AccountId]; ok {
					go ws.hub.sendMessageToClient(client, answer)
					log.Println("NATS: Send WS Message") //	TODO
				} else if message.SendPush {
					/*pushMessage := &sdk.ApiUserPushResponse{
						Type: message.Message.Type,
						Data: message.Message.Data,
					}
					go ws.hub.app.Sdk.UserPush(message.AccountId, pushMessage)
					log.Println("NATS: Send Push Message")*/ //	TODO
				}
			}
		} else if message.RoomId == uuid.Nil && message.AccountId == uuid.Nil {
			//	system message ws only
			fmt.Println(" -----> system message ws only!", *message.Message)
			switch message.Message.Type {
			case infrastructure.SystemMsgTypeUserSubscribe:
				messageData := &models.WSSystemUserSubscribeRequest{}
				err := json.Unmarshal(data, messageData)
				if err != nil {
					infrastructure.SetError(&sentry.SystemError{
						Error:   err,
						Message: WsCreateClientResponse,
						Code:    WsCreateClientResponseCode,
						Data:    []byte("Type: " + message.Message.Type),
					})
				} else {
					if client, ok := ws.hub.clientsId[messageData.Message.Data.Account.AccountId]; ok {
						//	update client
						cliSubscribes := ws.hub.app.DB.AccountSubscribes(messageData.Message.Data.Account.AccountId)
						ws.hub.clientsId[messageData.Message.Data.Account.AccountId].SetSubscribers(cliSubscribes)

						//	update room
						room := ws.hub.CreateRoomIfNotExists(messageData.Message.Data.ChatId)
						room.AddClient(client.uniqId)
						client.rooms[messageData.Message.Data.ChatId] = room
					} else if room, ok := ws.hub.rooms[messageData.Message.Data.ChatId]; ok {
						subscribers := ws.hub.app.DB.ChatSubscribes(messageData.Message.Data.ChatId)
						room.UpdateSubscribers(subscribers)
					}
				}
				break
			case infrastructure.SystemMsgTypeUserUnsubscribe:
				messageData := &models.WSSystemUserUnsubscribeRequest{}
				err := json.Unmarshal(data, messageData)
				if err != nil {
					infrastructure.SetError(&sentry.SystemError{
						Error:   err,
						Message: WsCreateClientResponse,
						Code:    WsCreateClientResponseCode,
						Data:    []byte("Type: " + message.Message.Type),
					})
				} else {
					if cli, ok := ws.hub.clientsId[messageData.Message.Data.AccountId]; ok {
						if room, ok := cli.rooms[messageData.Message.Data.ChatId]; ok {
							room.removeClient(cli)
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
		infrastructure.SetError(&sentry.SystemError{
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
		infrastructure.SetError(&sentry.SystemError{
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
		infrastructure.SetError(&sentry.SystemError{
			Error:   nil,
			Message: WsEmptyToken,
			Code:    WsEmptyTokenCode,
			Data:    []byte("token: " + token),
		})

		return
	}

	//	account
	accountModel := &sdk.AccountModel{}
	sdkErr := h.app.Sdk.UserByToken(token, accountModel)
	if sdkErr != nil || accountModel.Id == uuid.Nil {
		response := &models.WSChatErrorResponse{
			Error: models.WSChatErrorErrorResponse{
				Message: WsUserIdentification,
				Code:    WsUserIdentificationCode,
			},
		}
		w.Write(createResponse(response))
		infrastructure.SetError(&sentry.SystemError{
			Error:   nil,
			Message: WsUserIdentification,
			Code:    WsUserIdentificationCode,
			Data:    []byte("token: " + token),
		})

		return
	}

	//	uniq
	uniqId, err := uuid.NewV4()
	if err != nil {
		infrastructure.SetError(&sentry.SystemError{
			Error:   err,
			Message: WsUniqueIdGenerateProblem,
			Code:    WsUniqueIdGenerateProblemCode,
		})

		return
	}

	//	rooms & subscribes
	rooms := make(map[uuid.UUID]*Room)
	subscribes := h.app.DB.AccountSubscribes(accountModel.Id)
	for chatId, _ := range subscribes {
		room := h.CreateRoomIfNotExists(chatId)
		room.AddClient(uniqId.String())
		rooms[chatId] = room
	}

	//	client
	client := &Client{
		hub:        h,
		conn:       conn,
		sendChan:   make(chan []byte, 256),
		uniqId:     uniqId.String(),
		account:    accountModel,
		rooms:      rooms,
		subscribes: subscribes,
	}

	//	websocket power
	client.hub.registerChan <- client
	go client.Write()
	go client.Read()
}
