package server

import (
	"chats/application"
	"chats/models"
	"chats/service"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"gitlab.medzdrav.ru/health-service/go-sdk"
	"gitlab.medzdrav.ru/health-service/go-tools"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type WsServer struct {
	apiTopic         string
	shutdownSleep    time.Duration
	port             string
	hub              *Hub
	server           *http.Server
	logs             *service.Logs
	actualUsers      map[uint]time.Time
	actualUsersMutex sync.Mutex
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewServer(app *application.Application) *WsServer {
	return &WsServer{
		apiTopic:      service.BusTopic(),
		port:          os.Getenv("WEBSOCKET_PORT"),
		hub:           NewHub(app),
		logs:          service.Init(),
		shutdownSleep: getShutdownSleep(),
		actualUsers:   make(map[uint]time.Time),
	}
}

func getShutdownSleep() time.Duration {
	timeout, _ := strconv.ParseInt(os.Getenv("SHUTDOWN_SLEEP"), 10, 0)
	return time.Duration(timeout) * time.Second
}

func (ws *WsServer) Run() {
	if service.Cron() {
		go ws.userServiceMessageManager()
		ws.consumer()
	} else {

		go ws.hub.Run()
		go ws.ApiConsumer()
		go ws.Consumer()
		go ws.provider()

		srv := &http.Server{Addr: ws.port}
		http.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
			IndexAction(ws.hub, w, r)
		})

		//	testing. not for production.
		/*http.HandleFunc("/ws/html/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "src/chats/main/home.html")
		})*/

		ws.server = srv

		log.Fatal(srv.ListenAndServe())
	}
}

func (ws *WsServer) Shutdown(ctx context.Context) {
	if service.Cron() {
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

	tools.Close()
	log.Println("sentry connection has been closed")

	if !service.Cron() {
		err = ws.server.Shutdown(ctx)
		if err != nil {
			log.Println("sentry connection has been shutdown with sentry")
		}
	}
}

func (ws *WsServer) Consumer() {
	dataChan := make(chan []byte, 1024)

	go ws.hub.app.Sdk.
		Subject(service.InsideTopic()).
		Consumer(dataChan)

	for {
		data := <-dataChan
		fmt.Println("- Consumer data:", string(data)) //	TODO
		message := &RoomMessage{}
		err := json.Unmarshal(data, message)
		if err != nil {
			service.SetError(&models.SystemError{
				Error:   err,
				Message: service.UnmarshallingError,
				Code:    service.UnmarshallingErrorCode,
				Data:    data,
			})
		}

		if message.RoomId > 0 {
			fmt.Println("- Consumer message to room", message.RoomId)
			if room, ok := ws.hub.rooms[message.RoomId]; ok {
				fmt.Println("- Consumer room exists")
				answer, err := json.Marshal(message.Message)
				if err != nil {
					service.SetError(&models.SystemError{
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

		} else if message.RoomId == 0 && message.UserId > 0 {
			fmt.Println("- Consumer message to client")
			answer, err := json.Marshal(message.Message)
			if err != nil {
				service.SetError(&models.SystemError{
					Error:   err,
					Message: WsCreateClientResponse,
					Code:    WsCreateClientResponseCode,
					Data:    []byte("Type: " + message.Message.Type),
				})
			} else {
				if client, ok := ws.hub.clientsId[message.UserId]; ok {
					go ws.hub.sendMessageToClient(client, answer)
					log.Println("NATS: Send WS Message") //	TODO
				} else if message.SendPush {
					/*pushMessage := &sdk.ApiUserPushResponse{
						Type: message.Message.Type,
						Data: message.Message.Data,
					}
					go ws.hub.app.Sdk.UserPush(message.UserId, pushMessage)
					log.Println("NATS: Send Push Message")*/ //	TODO
				}
			}
		} else if message.RoomId == 0 && message.UserId == 0 {
			//	system message ws only
			fmt.Println(" -----> system message ws only!", *message.Message)
			switch message.Message.Type {
			case service.SystemMsgTypeUserSubscribe:
				messageData := &models.WSSystemUserSubscribeRequest{}
				err := json.Unmarshal(data, messageData)
				if err != nil {
					service.SetError(&models.SystemError{
						Error:   err,
						Message: WsCreateClientResponse,
						Code:    WsCreateClientResponseCode,
						Data:    []byte("Type: " + message.Message.Type),
					})
				} else {
					if client, ok := ws.hub.clientsId[messageData.Message.Data.UserId]; ok {
						//	update client
						cliSubscribes := ws.hub.app.DB.UserSubscribes(messageData.Message.Data.UserId)
						ws.hub.clientsId[messageData.Message.Data.UserId].SetSubscribers(cliSubscribes)

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
			case service.SystemMsgTypeUserUnsubscribe:
				messageData := &models.WSSystemUserUnsubscribeRequest{}
				err := json.Unmarshal(data, messageData)
				if err != nil {
					service.SetError(&models.SystemError{
						Error:   err,
						Message: WsCreateClientResponse,
						Code:    WsCreateClientResponseCode,
						Data:    []byte("Type: " + message.Message.Type),
					})
				} else {
					if cli, ok := ws.hub.clientsId[messageData.Message.Data.UserId]; ok {
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
		service.SetError(&models.SystemError{
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
		service.SetError(&models.SystemError{
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
		service.SetError(&models.SystemError{
			Error:   nil,
			Message: WsEmptyToken,
			Code:    WsEmptyTokenCode,
			Data:    []byte("token: " + token),
		})

		return
	}

	//	user
	userModel := &sdk.UserModel{}
	sdkErr := h.app.Sdk.UserByToken(token, userModel)
	if sdkErr != nil || userModel.Id == 0 {
		response := &models.WSChatErrorResponse{
			Error: models.WSChatErrorErrorResponse{
				Message: WsUserIdentification,
				Code:    WsUserIdentificationCode,
			},
		}
		w.Write(createResponse(response))
		service.SetError(&models.SystemError{
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
		service.SetError(&models.SystemError{
			Error:   err,
			Message: WsUniqueIdGenerateProblem,
			Code:    WsUniqueIdGenerateProblemCode,
		})

		return
	}

	//	rooms & subscribes
	rooms := make(map[uint]*Room)
	subscribes := h.app.DB.UserSubscribes(userModel.Id)
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
		user:       userModel,
		rooms:      rooms,
		subscribes: subscribes,
	}

	//	websocket power
	client.hub.registerChan <- client
	go client.Write()
	go client.Read()
}
