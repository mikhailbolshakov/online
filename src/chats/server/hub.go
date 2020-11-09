package server

import (
	"chats/application"
	"chats/models"
	"chats/service"
	"encoding/json"
	"gitlab.medzdrav.ru/health-service/go-sdk"
	"log"
	"sync"
	"time"
)

type Hub struct {
	clients        map[string]*Client
	clientsId      map[uint]*Client
	users          map[uint]bool
	rooms          map[uint]*Room
	roomMutex      sync.Mutex
	registerChan   chan *Client
	unregisterChan chan *Client
	messageChan    chan *RoomMessage
	router         *Router
	app            *application.Application
}

func NewHub(app *application.Application) *Hub {

	return &Hub{
		clients:        make(map[string]*Client),
		clientsId:      make(map[uint]*Client),
		users:          make(map[uint]bool),
		rooms:          make(map[uint]*Room),
		registerChan:   make(chan *Client),
		unregisterChan: make(chan *Client),
		messageChan:    make(chan *RoomMessage),
		router:         SetRouter(),
		app:            app,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.registerChan:
			h.clients[client.uniqId] = client
			h.clientsId[client.user.Id] = client
			h.users[client.user.Id] = true
			log.Println(">>> client register:", client.user.Id) //	TODO
			h.checkConnectionStatus(client.user.Id, true)
		case client := <-h.unregisterChan:
			go h.clientConnectionChange(client)
			h.onClientDisconnect(client)
			log.Println(">>> client unregister:", client.user.Id) //	TODO
			h.checkConnectionStatus(client.user.Id, false)
		case message := <-h.messageChan:
			log.Println(">>> message to client:", message.Message) //	TODO

			{ //	Scaling
				answer, err := json.Marshal(message)
				if err != nil {
					log.Println("Не удалось сформировать ответ клиенту")
					log.Println(err)
				} else {
					go h.app.Sdk.Subject(service.InsideTopic()).Publish(answer) //	return sentry
				}
			}
		}
	}
}

func (h *Hub) onClientDisconnect(client *Client) {
	if _, ok := h.clients[client.uniqId]; ok {
		delete(h.clients, client.uniqId)
		h.removeClientFromRooms(client)
		close(client.sendChan)

		if !h.userHasClients(client.user) {
			delete(h.users, client.user.Id)
			delete(h.clientsId, client.user.Id)
		}
	}
}

func (h *Hub) userHasClients(user *sdk.UserModel) bool {
	for _, otherClient := range h.clients {
		if otherClient.user.Id == user.Id {
			return true
		}
	}

	return false
}

func (h *Hub) removeClientFromRooms(client *Client) {
	for _, room := range client.rooms {
		room.removeClient(client)
	}
}

func (h *Hub) removeAllClients() {
	for _, client := range h.clients {
		h.onClientDisconnect(client)
	}
}

func (h *Hub) CreateRoomIfNotExists(chatId uint) *Room {
	h.roomMutex.Lock()
	defer h.roomMutex.Unlock()

	subscribers := h.app.DB.ChatSubscribes(chatId)
	if _, ok := h.rooms[chatId]; !ok {
		room := CreateRoom(chatId, subscribers)
		h.rooms[chatId] = room

		return room
	} else {
		h.rooms[chatId].subscribers = subscribers
	}

	return h.rooms[chatId]
}

func (h *Hub) onClientMessage(message []byte, client *Client) {
	h.router.Dispatch(h, client, message)
}

func (h *Hub) SendMessageToRoom(message *RoomMessage) {
	h.messageChan <- message
}

func (h *Hub) sendMessageToClient(client *Client, message []byte) {
	client.send(message)
}

func (h *Hub) checkConnectionStatus(userId uint, online bool) {
	if !online {
		if _, ok := h.clientsId[userId]; !ok {
			go h.app.Sdk.ChangeConnectionStatus(userId, online)
		}
	} else {
		go h.app.Sdk.ChangeConnectionStatus(userId, online)
	}
}

func (h *Hub) clientConnectionChange(c *Client) {
	time.Sleep(10 * time.Second)

	//	todo вынести в отдельный метод
	cronGetUserStatusRequestMessage := &models.CronGetUserStatusRequest{
		Type: MessageTypeGetUserStatus,
		Data: models.CronGetUserStatusRequestData{
			UserId: c.user.Id,
		},
	}

	request, err := json.Marshal(cronGetUserStatusRequestMessage)
	if err != nil {
		service.SetError(service.MarshalError1011(err, nil))
		return
	}

	response, cronRequestError := h.app.Sdk.
		Subject(service.CronTopic()).
		Request(request)

	if cronRequestError != nil {
		service.SetError(&models.SystemError{
			Error:   cronRequestError.Error,
			Message: service.CronResponseError,
			Code:    service.CronResponseErrorCode,
			Data:    append(request, response...),
		})
		return
	}

	cronGetUserResponse := &models.CronGetUserStatusResponse{}
	err = json.Unmarshal(response, cronGetUserResponse)
	if err != nil {
		service.SetError(service.MarshalError1011(err, response))
		return
	}

	if !cronGetUserResponse.Online {
		//if _, ok := h.clientsId[c.user.Id]; !ok {	//	plan B
		opponentId := h.app.DB.LastOpponentId(c.user.Id)

		if opponentId > 0 {
			data := new(models.UserStatusModel)
			data.UserId = c.user.Id

			roomMessage := &RoomMessage{
				UserId: opponentId,
				Message: &models.WSChatResponse{
					Type: EventClientConnectionError,
					Data: &data,
				},
			}

			go h.SendMessageToRoom(roomMessage)
		}

		/*for chatId, item := range c.subscribes {
			if item.UserType == database.UserTypeClient {
				data := new(models.UserStatusModel)
				data.UserId = c.user.Id

				roomMessage := &RoomMessage{
					RoomId: chatId,
					Message: &models.WSChatResponse{
						Type: EventClientConnectionError,
						Data: &data,
					},
				}

				go h.SendMessageToRoom(roomMessage)
			}
		}*/
	}
}
